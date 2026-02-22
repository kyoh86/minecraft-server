package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func (a app) sendConsole(command string) error {
	if err := a.waitWorldReady(90 * time.Second); err != nil {
		return err
	}
	composeFile := a.composeFilePath()
	return runCommand(
		"docker", "compose", "-f", composeFile,
		"exec", "-T", "--user", "1000", "world", "mc-send-to-console", command,
	)
}

func (a app) waitWorldReady(timeout time.Duration) error {
	composeFile := a.composeFilePath()
	started := time.Now()
	deadline := time.Now().Add(timeout)
	lastReport := time.Time{}

	for {
		status := "world=missing"
		containerID, err := runCommandOutput("docker", "compose", "-f", composeFile, "ps", "-q", "world")
		if err == nil {
			containerID = strings.TrimSpace(containerID)
			if containerID != "" {
				state, err := runCommandOutput(
					"docker", "inspect",
					"--format", "{{.State.Status}} {{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}",
					containerID,
				)
				if err == nil {
					parts := strings.Fields(strings.TrimSpace(state))
					if len(parts) >= 2 {
						pipeReady := a.worldConsolePipeReady(composeFile)
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
	composeFile := a.composeFilePath()
	started := time.Now()
	deadline := time.Now().Add(timeout)
	lastReport := time.Time{}

	servicesOut, err := runCommandOutput("docker", "compose", "-f", composeFile, "config", "--services")
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
		return fmt.Errorf("no services found in compose file: %s", composeFile)
	}

	for {
		allReady := true
		statuses := make([]string, 0, len(services))
		for _, service := range services {
			containerID, err := runCommandOutput("docker", "compose", "-f", composeFile, "ps", "-q", service)
			containerID = strings.TrimSpace(containerID)
			if err != nil || containerID == "" {
				allReady = false
				statuses = append(statuses, fmt.Sprintf("%s=missing", service))
				continue
			}

			state, err := runCommandOutput(
				"docker", "inspect",
				"--format", "{{.State.Status}} {{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}",
				containerID,
			)
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

func (a app) worldConsolePipeReady(composeFile string) bool {
	_, err := runCommandOutput(
		"docker", "compose", "-f", composeFile,
		"exec", "-T", "world", "sh", "-lc",
		"test -p /tmp/minecraft-console-in",
	)
	return err == nil
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runCommandOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	out := stdout.String()
	if serr := stderr.String(); serr != "" {
		if out != "" {
			out += "\n"
		}
		out += serr
	}
	return out, err
}

func (a app) ensureRuntimeLayout() error {
	for _, dir := range []string{
		filepath.Join(a.baseDir, "runtime"),
		filepath.Join(a.baseDir, "runtime", "mc-link"),
		filepath.Join(a.baseDir, "runtime", "limbo"),
		filepath.Join(a.baseDir, "runtime", "redis"),
		filepath.Join(a.baseDir, "runtime", "world"),
		filepath.Join(a.baseDir, "runtime", "world", "config"),
		filepath.Join(a.baseDir, "runtime", "world", "plugins"),
		filepath.Join(a.baseDir, "runtime", "world", "plugins", "ClickMobs"),
		filepath.Join(a.baseDir, "runtime", "world", "plugins", "WorldGuard", "worlds"),
		filepath.Join(a.baseDir, "runtime", "world", "plugins", "Multiverse-Core"),
		filepath.Join(a.baseDir, "runtime", "world", "plugins", "Multiverse-Portals"),
		filepath.Join(a.baseDir, "runtime", "velocity"),
		filepath.Join(a.baseDir, "runtime", "velocity", "plugins"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		probePath := filepath.Join(dir, ".writecheck")
		if err := os.WriteFile(probePath, []byte("ok"), 0o644); err != nil {
			return err
		}
		_ = os.Remove(probePath)
	}
	allowlistPath := filepath.Join(a.baseDir, "runtime", "velocity", "allowlist.yml")
	if !fileExists(allowlistPath) {
		const defaultAllowlist = "uuids: []\nnicks: []\n"
		if err := os.WriteFile(allowlistPath, []byte(defaultAllowlist), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func (a app) checkRuntimeOwnership() error {
	root := filepath.Join(a.baseDir, "runtime")
	wantUID := uint32(os.Getuid())
	wantGID := uint32(os.Getgid())

	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		st, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			return nil
		}
		if st.Uid != wantUID || st.Gid != wantGID {
			return fmt.Errorf(
				"ownership mismatch: %s is %d:%d, expected %d:%d (fix with: sudo chown -R %d:%d runtime)",
				path, st.Uid, st.Gid, wantUID, wantGID, wantUID, wantGID,
			)
		}
		return nil
	})
}

func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()

		info, err := d.Info()
		if err != nil {
			return err
		}

		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
		if err != nil {
			return err
		}
		defer out.Close()

		_, err = io.Copy(out, in)
		return err
	})
}
