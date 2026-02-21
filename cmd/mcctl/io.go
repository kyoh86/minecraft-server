package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	deadline := time.Now().Add(timeout)

	for {
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
					if len(parts) >= 2 && parts[0] == "running" && (parts[1] == "healthy" || parts[1] == "none") {
						if a.worldConsolePipeReady(composeFile) {
							return nil
						}
					}
				}
			}
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("world container is not ready within %s", timeout)
		}
		time.Sleep(500 * time.Millisecond)
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

func (a app) runtimeWorldDir() string {
	return filepath.Join(a.baseDir, "runtime", "world")
}

func (a app) runtimeVelocityDir() string {
	return filepath.Join(a.baseDir, "runtime", "velocity")
}

func (a app) runtimeRedisDir() string {
	return filepath.Join(a.baseDir, "runtime", "redis")
}

func (a app) ensureRuntimeLayout() error {
	for _, dir := range []string{
		filepath.Join(a.baseDir, "runtime"),
		filepath.Join(a.baseDir, "runtime", "mclink"),
		a.runtimeRedisDir(),
		a.runtimeWorldDir(),
		filepath.Join(a.runtimeWorldDir(), ".mcctl"),
		filepath.Join(a.runtimeWorldDir(), "config"),
		filepath.Join(a.runtimeWorldDir(), "plugins"),
		filepath.Join(a.runtimeWorldDir(), "plugins", "ClickMobs"),
		filepath.Join(a.runtimeWorldDir(), "plugins", "WorldGuard", "worlds"),
		filepath.Join(a.runtimeWorldDir(), "plugins", "Multiverse-Core"),
		filepath.Join(a.runtimeWorldDir(), "plugins", "Multiverse-Portals"),
		a.runtimeVelocityDir(),
		filepath.Join(a.runtimeVelocityDir(), "plugins"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func (a app) ensureRuntimeWritable() error {
	if err := a.ensureRuntimeLayout(); err != nil {
		if !errors.Is(err, os.ErrPermission) {
			return err
		}
		if err := a.fixRuntimeOwnership(); err != nil {
			return fmt.Errorf("repair runtime ownership (layout): %w", err)
		}
		if err := a.ensureRuntimeLayout(); err != nil {
			return err
		}
	}

	probeChecks := []struct {
		dir string
		fix func() error
	}{
		{dir: a.runtimeWorldDir(), fix: a.fixRuntimeWorldOwnership},
		{dir: a.runtimeVelocityDir(), fix: a.fixRuntimeVelocityOwnership},
	}
	for _, check := range probeChecks {
		probePath := filepath.Join(check.dir, ".writecheck")
		if err := os.WriteFile(probePath, []byte("ok"), 0o644); err != nil {
			if !errors.Is(err, os.ErrPermission) {
				return err
			}
			if err := check.fix(); err != nil {
				return fmt.Errorf("repair runtime ownership (%s): %w", check.dir, err)
			}
			if err := os.WriteFile(probePath, []byte("ok"), 0o644); err != nil {
				return err
			}
		}
		_ = os.Remove(probePath)
	}
	return nil
}

func (a app) fixRuntimeOwnership() error {
	if err := a.fixRuntimeWorldOwnership(); err != nil {
		return err
	}
	return a.fixRuntimeVelocityOwnership()
}

func (a app) fixRuntimeWorldOwnership() error {
	composeFile := a.composeFilePath()
	cmd := fmt.Sprintf("chown -R %d:%d /data", os.Getuid(), os.Getgid())
	return runCommand(
		"docker", "compose", "-f", composeFile,
		"run", "--rm", "--no-deps", "--user", "root", "--entrypoint", "sh",
		"world", "-lc", cmd,
	)
}

func (a app) fixRuntimeVelocityOwnership() error {
	composeFile := a.composeFilePath()
	cmd := fmt.Sprintf("chown -R %d:%d /server", os.Getuid(), os.Getgid())
	return runCommand(
		"docker", "compose", "-f", composeFile,
		"run", "--rm", "--no-deps", "--user", "root", "--entrypoint", "sh",
		"velocity", "-lc", cmd,
	)
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
