package main

import (
	"fmt"
	"time"
)

func (a app) serverUp() error {
	if err := a.compose("up", "-d", "--remove-orphans"); err != nil {
		return err
	}
	if err := a.waitServicesReady(120 * time.Second); err != nil {
		return err
	}
	fmt.Println("All containers are running/healthy.")
	return nil
}

func (a app) serverDown() error {
	return a.compose("down")
}

func (a app) serverRestart(service string, build, recreate bool) error {
	if build {
		if err := a.compose("up", "-d", "--build", "--force-recreate", service); err != nil {
			return err
		}
	} else if recreate {
		if err := a.compose("up", "-d", "--force-recreate", service); err != nil {
			return err
		}
	} else {
		if err := a.compose("restart", service); err != nil {
			return err
		}
	}
	if err := a.waitServiceReady(service, 120*time.Second); err != nil {
		return err
	}
	if service == "world" {
		if err := a.waitWorldReady(120 * time.Second); err != nil {
			return err
		}
	}
	fmt.Printf("Service %s is running/healthy.\n", service)
	return nil
}

func (a app) serverPS() error {
	return a.compose("ps")
}

func (a app) serverLogs(service string, follow bool) error {
	args := []string{"logs"}
	if follow {
		args = append(args, "--follow")
	}
	if service != "" {
		args = append(args, service)
	}
	return a.compose(args...)
}

func (a app) serverReload() error {
	return a.sendConsole("reload")
}
