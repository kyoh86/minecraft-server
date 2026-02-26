package main

import (
	"fmt"
	"strings"
	"time"
)

func (a app) sendConsole(command string) error {
	if err := validateConsoleCommand(command); err != nil {
		return err
	}
	if err := a.waitWorldReady(90 * time.Second); err != nil {
		return err
	}
	return a.compose("exec", "-T", "--user", a.localUID(), "world", "mc-send-to-console", command)
}

func (a app) waitWorldReady(timeout time.Duration) error {
	started := time.Now()
	deadline := time.Now().Add(timeout)
	lastReport := time.Time{}

	for {
		status := "world=missing"
		containerID, err := a.composeOutput("ps", "-q", "world")
		if err == nil {
			containerID = strings.TrimSpace(containerID)
			if containerID != "" {
				state, err := dockerInspect(containerID)
				if err == nil {
					parts := strings.Fields(strings.TrimSpace(state))
					if len(parts) >= 2 {
						pipeReady := a.worldConsolePipeReady()
						status = fmt.Sprintf("world=%s/%s pipe=%t", parts[0], parts[1], pipeReady)
						if parts[0] == "running" && (parts[1] == "healthy" || parts[1] == "none") && pipeReady {
							return nil
						}
					}
				}
			}
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("world container is not ready within %s", timeout)
		}
		if lastReport.IsZero() || time.Since(lastReport) >= 3*time.Second {
			fmt.Printf("Waiting for world readiness (%s elapsed): %s\n", time.Since(started).Truncate(time.Second), status)
			lastReport = time.Now()
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func (a app) waitServicesReady(timeout time.Duration) error {
	started := time.Now()
	deadline := time.Now().Add(timeout)
	lastReport := time.Time{}

	servicesOut, err := a.composeOutput("config", "--services")
	if err != nil {
		return err
	}
	services := []string{}
	for _, line := range strings.Split(servicesOut, "\n") {
		name := strings.TrimSpace(line)
		if name != "" {
			services = append(services, name)
		}
	}
	if len(services) == 0 {
		return fmt.Errorf("no services found in compose file: %s", a.composeFilePath())
	}

	for {
		allReady := true
		statuses := make([]string, 0, len(services))
		for _, service := range services {
			containerID, err := a.composeOutput("ps", "-q", service)
			containerID = strings.TrimSpace(containerID)
			if err != nil || containerID == "" {
				allReady = false
				statuses = append(statuses, fmt.Sprintf("%s=missing", service))
				continue
			}

			state, err := dockerInspect(containerID)
			if err != nil {
				allReady = false
				statuses = append(statuses, fmt.Sprintf("%s=inspect_error", service))
				continue
			}

			parts := strings.Fields(strings.TrimSpace(state))
			if len(parts) < 2 {
				allReady = false
				statuses = append(statuses, fmt.Sprintf("%s=unknown", service))
				continue
			}
			status := parts[0]
			health := parts[1]
			ready := status == "running" && (health == "healthy" || health == "none")
			if !ready {
				allReady = false
			}
			statuses = append(statuses, fmt.Sprintf("%s=%s/%s", service, status, health))
		}

		if allReady {
			return nil
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("services are not ready within %s: %s", timeout, strings.Join(statuses, ", "))
		}
		if lastReport.IsZero() || time.Since(lastReport) >= 3*time.Second {
			fmt.Printf("Waiting for services (%s elapsed): %s\n", time.Since(started).Truncate(time.Second), strings.Join(statuses, ", "))
			lastReport = time.Now()
		}
		time.Sleep(1 * time.Second)
	}
}

func (a app) waitServiceReady(service string, timeout time.Duration) error {
	started := time.Now()
	deadline := time.Now().Add(timeout)
	lastReport := time.Time{}

	service = strings.TrimSpace(service)
	if service == "" {
		return fmt.Errorf("service name is required")
	}

	for {
		status := service + "=missing"
		containerID, err := a.composeOutput("ps", "-q", service)
		containerID = strings.TrimSpace(containerID)
		if err == nil && containerID != "" {
			state, err := dockerInspect(containerID)
			if err == nil {
				parts := strings.Fields(strings.TrimSpace(state))
				if len(parts) >= 2 {
					status = fmt.Sprintf("%s=%s/%s", service, parts[0], parts[1])
					if parts[0] == "running" && (parts[1] == "healthy" || parts[1] == "none") {
						return nil
					}
				}
			}
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("service %q is not ready within %s", service, timeout)
		}
		if lastReport.IsZero() || time.Since(lastReport) >= 3*time.Second {
			fmt.Printf("Waiting for service readiness (%s elapsed): %s\n", time.Since(started).Truncate(time.Second), status)
			lastReport = time.Now()
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func (a app) worldConsolePipeReady() bool {
	_, err := a.composeOutput("exec", "-T", "world", "sh", "-lc", "test -p /tmp/minecraft-console-in")
	return err == nil
}
