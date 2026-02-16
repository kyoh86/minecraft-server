package main

import (
	"errors"
	"path/filepath"
)

type worldConfig struct {
	Name        string `yaml:"name"`
	Environment string `yaml:"environment"`
	WorldType   string `yaml:"world_type"`
	Seed        any    `yaml:"seed"`
	Deletable   bool   `yaml:"deletable"`
}

type worldPolicy struct {
	Name  string            `yaml:"name"`
	MVSet map[string]string `yaml:"mv_set"`
}

type portalsConfig struct {
	Portals map[string]portalDef `yaml:"portals"`
}

type portalDef struct {
	Owner                  string `yaml:"owner"`
	Location               string `yaml:"location"`
	ActionType             string `yaml:"action-type"`
	Action                 string `yaml:"action"`
	SafeTeleport           bool   `yaml:"safe-teleport"`
	TeleportNonPlayers     bool   `yaml:"teleport-non-players"`
	CheckDestinationSafety bool   `yaml:"check-destination-safety"`
}

type app struct {
	repoRoot string
	wslDir   string
}

func newApp() (app, error) {
	repoRoot, err := findRepoRoot()
	if err != nil {
		return app{}, err
	}
	return app{repoRoot: repoRoot, wslDir: filepath.Join(repoRoot, "setup", "wsl")}, nil
}

func findRepoRoot() (string, error) {
	cwd, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}
	for dir := cwd; ; dir = filepath.Dir(dir) {
		if fileExists(filepath.Join(dir, "setup", "wsl")) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("could not find repo root (setup/wsl not found)")
		}
	}
}
