package main

import "path/filepath"

func (a app) compose(args ...string) error {
	composeFile := filepath.Join(a.wslDir, "docker-compose.yml")
	base := []string{"compose", "-f", composeFile}
	base = append(base, args...)
	return runCommand("docker", base...)
}

func (a app) serverUp() error {
	if err := a.syncStaticRuntimeConfigs(); err != nil {
		return err
	}
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

func (a app) syncStaticRuntimeConfigs() error {
	if _, err := a.syncWorldGuardRegionsConfig(primaryWorldName); err != nil {
		return err
	}
	cfgs, err := a.listWorldConfigs()
	if err != nil {
		return err
	}
	for _, cfgPath := range cfgs {
		cfg, err := loadWorldConfig(cfgPath)
		if err != nil {
			return err
		}
		if _, err := a.syncWorldGuardRegionsConfig(cfg.Name); err != nil {
			return err
		}
	}
	return nil
}
