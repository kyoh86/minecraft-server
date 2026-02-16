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
	Function    string `yaml:"function"`
	Resettable  bool   `yaml:"resettable"`
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
