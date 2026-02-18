package main

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
	if service == "" {
		service = "world"
	}
	return a.compose("restart", service)
}

func (a app) serverPS() error {
	return a.compose("ps")
}

func (a app) serverLogs(service string) error {
	args := []string{"logs", "-f"}
	if service != "" {
		args = append(args, service)
	}
	return a.compose(args...)
}

func (a app) serverReload() error {
	return a.sendConsole("reload")
}
