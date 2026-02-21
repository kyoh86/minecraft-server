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

type app struct {
	repoRoot string
	baseDir  string
}

func newApp() (app, error) {
	repoRoot, err := findRepoRoot()
	if err != nil {
		return app{}, err
	}
	return app{repoRoot: repoRoot, baseDir: repoRoot}, nil
}

func (a app) composeFilePath() string {
	return filepath.Join(a.baseDir, "infra", "docker-compose.yml")
}

func findRepoRoot() (string, error) {
	cwd, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}
	for dir := cwd; ; dir = filepath.Dir(dir) {
		if fileExists(filepath.Join(dir, "infra", "docker-compose.yml")) &&
			fileExists(filepath.Join(dir, "worlds", "env.schema.json")) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("could not find repo root (infra/docker-compose.yml not found)")
		}
	}
}
