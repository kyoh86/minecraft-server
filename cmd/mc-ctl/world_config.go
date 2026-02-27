package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

func (a app) loadWorldPolicy(worldName string) (worldPolicy, bool, error) {
	path := filepath.Join(a.baseDir, "worlds", worldName, "config.toml")
	if !fileExists(path) {
		return worldPolicy{}, false, nil
	}
	cfg, err := loadWorldConfig(path)
	if err != nil {
		return worldPolicy{}, false, fmt.Errorf("parse world policy %s: %w", path, err)
	}
	policy := worldPolicy{Name: cfg.Name, MVSet: cfg.MVSet}
	if policy.Name != "" && policy.Name != worldName {
		return worldPolicy{}, false, fmt.Errorf("world policy name mismatch: %s != %s", policy.Name, worldName)
	}
	return policy, true, nil
}

func loadWorldConfig(path string) (worldConfig, error) {
	if !fileExists(path) {
		return worldConfig{}, fmt.Errorf("missing world config: %s", path)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return worldConfig{}, err
	}
	var cfg worldConfig
	if err := toml.Unmarshal(b, &cfg); err != nil {
		return worldConfig{}, fmt.Errorf("parse world config %s: %w", path, err)
	}
	if err := validateSeedValue("seed", cfg.Seed); err != nil {
		return worldConfig{}, fmt.Errorf("invalid world config %s: %w", path, err)
	}
	if cfg.Dimensions.Nether != nil {
		if err := validateSeedValue("dimensions.nether.seed", cfg.Dimensions.Nether.Seed); err != nil {
			return worldConfig{}, fmt.Errorf("invalid world config %s: %w", path, err)
		}
	}
	if cfg.Dimensions.End != nil {
		if err := validateSeedValue("dimensions.end.seed", cfg.Dimensions.End.Seed); err != nil {
			return worldConfig{}, fmt.Errorf("invalid world config %s: %w", path, err)
		}
	}
	if strings.TrimSpace(cfg.WorldType) == "" {
		cfg.WorldType = "normal"
	}
	return cfg, nil
}

func validateSeedValue(path string, seed any) error {
	if s, ok := seed.(string); ok && strings.TrimSpace(s) == "" {
		return fmt.Errorf("%s must not be empty string; omit it for random seed", path)
	}
	return nil
}
