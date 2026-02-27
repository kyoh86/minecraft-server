package main

import (
	"fmt"
	"strings"
)

type dimensionTarget struct {
	Name        string
	Environment string
	Seed        any
	LinkKind    string
}

func managedDimensionTargets(cfg worldConfig) ([]dimensionTarget, error) {
	targets := make([]dimensionTarget, 0, 2)
	used := map[string]struct{}{cfg.Name: {}}
	addTarget := func(defaultName, env, linkKind string, dim *worldDimension) error {
		if dim == nil {
			return nil
		}
		name := strings.TrimSpace(dim.Name)
		if name == "" {
			name = defaultName
		}
		if err := validateWorldName(name); err != nil {
			return fmt.Errorf("invalid dimensions.%s.name %q: %w", linkKind, name, err)
		}
		if _, exists := used[name]; exists {
			return fmt.Errorf("duplicate managed world name %q in dimensions", name)
		}
		used[name] = struct{}{}
		targets = append(targets, dimensionTarget{
			Name:        name,
			Environment: env,
			Seed:        dim.Seed,
			LinkKind:    linkKind,
		})
		return nil
	}
	if err := addTarget(cfg.Name+"_nether", "nether", "nether", cfg.Dimensions.Nether); err != nil {
		return nil, err
	}
	if err := addTarget(cfg.Name+"_the_end", "the_end", "end", cfg.Dimensions.End); err != nil {
		return nil, err
	}
	return targets, nil
}
