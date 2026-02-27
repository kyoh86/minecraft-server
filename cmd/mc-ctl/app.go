package main

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type worldConfig struct {
	Name         string            `toml:"name"`
	DisplayName  string            `toml:"display_name"`
	DisplayColor string            `toml:"display_color"`
	WorldType    string            `toml:"world_type"`
	Seed         any               `toml:"seed"`
	Deletable    bool              `toml:"deletable"`
	Dimensions   worldDimensions   `toml:"dimensions"`
	MVSet        map[string]string `toml:"mv_set"`
}

type worldDimensions struct {
	Nether *worldDimension `toml:"nether"`
	End    *worldDimension `toml:"end"`
}

type worldDimension struct {
	Name string `toml:"name"`
	Seed any    `toml:"seed"`
}

type worldPolicy struct {
	Name  string            `toml:"name"`
	MVSet map[string]string `toml:"mv_set"`
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
			fileExists(filepath.Join(dir, "worlds", "config.schema.json")) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("could not find repo root (infra/docker-compose.yml not found)")
		}
	}
}
