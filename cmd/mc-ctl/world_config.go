package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

func (a app) loadWorldPolicy(worldName string) (worldPolicy, bool, error) {
	path := filepath.Join(a.baseDir, "worlds", worldName, "world.policy.yml")
	if !fileExists(path) {
		return worldPolicy{}, false, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return worldPolicy{}, false, err
	}
	var policy worldPolicy
	if err := yaml.Unmarshal(b, &policy); err != nil {
		return worldPolicy{}, false, fmt.Errorf("parse world policy %s: %w", path, err)
	}
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
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return worldConfig{}, fmt.Errorf("parse world config %s: %w", path, err)
	}
	var raw map[string]any
	if err := yaml.Unmarshal(b, &raw); err == nil {
		if err := validateNoExplicitNullSeed(raw); err != nil {
			return worldConfig{}, fmt.Errorf("invalid world config %s: %w", path, err)
		}
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

func validateNoExplicitNullSeed(raw map[string]any) error {
	if v, exists := raw["seed"]; exists && v == nil {
		return fmt.Errorf("seed must not be null; omit it for random seed")
	}
	dimensions, ok := toStringAnyMap(raw["dimensions"])
	if !ok {
		return nil
	}
	for _, dimName := range []string{"nether", "end"} {
		dimValue, exists := dimensions[dimName]
		if !exists {
			continue
		}
		dimMap, ok := toStringAnyMap(dimValue)
		if !ok {
			continue
		}
		if v, exists := dimMap["seed"]; exists && v == nil {
			return fmt.Errorf("dimensions.%s.seed must not be null; omit it for fallback", dimName)
		}
	}
	return nil
}

func toStringAnyMap(v any) (map[string]any, bool) {
	switch m := v.(type) {
	case map[string]any:
		return m, true
	case map[any]any:
		out := make(map[string]any, len(m))
		for k, value := range m {
			key, ok := k.(string)
			if !ok {
				continue
			}
			out[key] = value
		}
		return out, true
	default:
		return nil, false
	}
}
