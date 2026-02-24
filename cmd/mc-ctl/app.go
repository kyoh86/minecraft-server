package main

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
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

func (a app) localUID() string {
	const fallback = "1000"
	path := a.composeEnvFilePath()
	f, err := os.Open(path)
	if err != nil {
		return fallback
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		if strings.TrimSpace(k) == "LOCAL_UID" {
			v = strings.TrimSpace(v)
			if v != "" {
				return v
			}
		}
	}
	return fallback
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
