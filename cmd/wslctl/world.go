package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/goccy/go-yaml"
)

func (a app) initRuntime() error {
	runtimeDir := filepath.Join(a.wslDir, "runtime", "world")
	if err := os.MkdirAll(runtimeDir, 0o755); err != nil {
		return err
	}
	fmt.Printf("Initialized: %s\n", filepath.Join(a.wslDir, "runtime"))
	return nil
}

func (a app) syncWorldDatapack() error {
	srcDir := filepath.Join(a.wslDir, "datapacks", "world-base")
	dstRoot := filepath.Join(a.wslDir, "runtime", "world", "world", "datapacks")
	dstDir := filepath.Join(dstRoot, "world-base")

	if !fileExists(srcDir) {
		return fmt.Errorf("missing datapack source: %s", srcDir)
	}
	if err := os.MkdirAll(dstRoot, 0o755); err != nil {
		return err
	}
	if err := os.RemoveAll(dstDir); err != nil {
		return err
	}
	if err := copyDir(srcDir, dstDir); err != nil {
		return err
	}

	fmt.Printf("synced datapack to %s\n", dstDir)
	return nil
}

func (a app) applyWorldSettings() error {
	if err := a.syncWorldDatapack(); err != nil {
		return err
	}
	if err := a.sendConsole("reload"); err != nil {
		return err
	}
	if err := a.sendConsole("function mcserver:world_settings"); err != nil {
		return err
	}
	fmt.Println("applied world settings function mcserver:world_settings")
	return nil
}

func (a app) worldsBootstrap() error {
	if err := a.syncWorldDatapack(); err != nil {
		return err
	}
	if err := a.sendConsole("reload"); err != nil {
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
		if err := a.ensureWorld(cfg, false); err != nil {
			return err
		}
		if err := a.applyWorldInitFunction(cfg); err != nil {
			return err
		}
	}

	fmt.Printf("bootstrapped worlds from %s\n", filepath.Join(a.wslDir, "worlds"))
	return nil
}

func (a app) worldReset(target string) error {
	target = strings.TrimSpace(target)
	if target == "" {
		return errors.New("world name is required")
	}

	cfgPath := filepath.Join(a.wslDir, "worlds", target, "world.env.yml")
	cfg, err := loadWorldConfig(cfgPath)
	if err != nil {
		return err
	}
	if !cfg.Resettable {
		return fmt.Errorf("world '%s' is not resettable", target)
	}

	registered, err := a.worldRegisteredInMultiverse(target)
	if err != nil {
		return err
	}
	if registered {
		if err := a.sendConsole(fmt.Sprintf("mv unload %s --remove-players", target)); err != nil {
			return err
		}
		if err := a.sendConsole(fmt.Sprintf("mv remove %s", target)); err != nil {
			return err
		}
	}

	for _, p := range []string{
		filepath.Join(a.wslDir, "runtime", "world", target),
		filepath.Join(a.wslDir, "runtime", "world", target+"_nether"),
		filepath.Join(a.wslDir, "runtime", "world", target+"_the_end"),
	} {
		if err := os.RemoveAll(p); err != nil {
			return err
		}
	}

	if err := a.ensureWorld(cfg, true); err != nil {
		return err
	}
	if err := a.sendConsole("reload"); err != nil {
		return err
	}
	if err := a.applyWorldInitFunction(cfg); err != nil {
		return err
	}

	fmt.Printf("reset world '%s'\n", target)
	return nil
}

func (a app) ensureWorld(cfg worldConfig, forceCreate bool) error {
	if cfg.Name == "" || cfg.Environment == "" {
		return fmt.Errorf("invalid world config: name/environment required")
	}

	if !forceCreate {
		registered, err := a.worldRegisteredInMultiverse(cfg.Name)
		if err != nil {
			return err
		}
		if registered {
			return nil
		}

		onDisk := fileExists(filepath.Join(a.wslDir, "runtime", "world", cfg.Name))
		if onDisk {
			return a.sendConsole(fmt.Sprintf("mv import %s %s", cfg.Name, cfg.Environment))
		}
	}

	parts := []string{"mv", "create", cfg.Name, cfg.Environment}
	if seed := formatSeed(cfg.Seed); seed != "" {
		parts = append(parts, "-s", seed)
	}
	if cfg.WorldType != "" {
		parts = append(parts, "-t", strings.ToUpper(cfg.WorldType))
	}
	return a.sendConsole(strings.Join(parts, " "))
}

func (a app) applyWorldInitFunction(cfg worldConfig) error {
	fn := cfg.Function
	if fn == "" {
		fn = fmt.Sprintf("mcserver:worlds/%s/init", cfg.Name)
	}
	return a.sendConsole("function " + fn)
}

func (a app) listWorldConfigs() ([]string, error) {
	root := filepath.Join(a.wslDir, "worlds")
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var cfgs []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		cfg := filepath.Join(root, e.Name(), "world.env.yml")
		if fileExists(cfg) {
			cfgs = append(cfgs, cfg)
		}
	}
	sort.Strings(cfgs)
	return cfgs, nil
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
	return cfg, nil
}

func (a app) worldRegisteredInMultiverse(worldName string) (bool, error) {
	path := filepath.Join(a.wslDir, "runtime", "world", "plugins", "Multiverse-Core", "worlds.yml")
	if !fileExists(path) {
		return false, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(worldName) + `:`)
	return re.Match(b), nil
}
