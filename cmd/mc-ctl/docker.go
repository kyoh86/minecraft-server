package main

import "path/filepath"

func (a app) composeFilePath() string {
	return filepath.Join(a.baseDir, "infra", "docker-compose.yml")
}

func (a app) composeEnvFilePath() string {
	return filepath.Join(a.baseDir, "infra", ".env")
}

func (a app) composeBaseArgs() []string {
	return []string{"compose", "-f", a.composeFilePath()}
}

func (a app) compose(args ...string) error {
	base := a.composeBaseArgs()
	base = append(base, args...)
	return runCommand("docker", base...)
}

func (a app) composeOutput(args ...string) (string, error) {
	base := a.composeBaseArgs()
	base = append(base, args...)
	stdout, _, err := runCommandOutput("docker", base...)
	return stdout, err
}

func dockerInspect(containerID string) (string, error) {
	stdout, _, err := runCommandOutput(
		"docker", "inspect",
		"--format", "{{.State.Status}} {{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}",
		containerID,
	)
	return stdout, err
}
