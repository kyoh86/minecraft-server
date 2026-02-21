package main

import (
	"fmt"
	"path/filepath"
)

func (a app) compose(args ...string) error {
	composeFile := a.composeFilePath()
	base := []string{"compose", "-f", composeFile}
	base = append(base, args...)
	return runCommand("docker", base...)
}

func (a app) serverUp() error {
	return a.compose("up", "-d", "--remove-orphans")
}

func (a app) serverDown() error {
	return a.compose("down")
}

func (a app) serverRestart(service string) error {
	return a.compose("restart", service)
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

func (a app) assetInit() error {
	if err := a.ensureRuntimeLayout(); err != nil {
		return err
	}
	fmt.Printf("Initialized assets: %s\n", filepath.Join(a.baseDir, "runtime"))
	return nil
}

func (a app) assetStage() error {
	if err := a.ensureRuntimeLayout(); err != nil {
		return err
	}
	fmt.Println("Staged runtime assets (directory checks only)")
	return nil
}
