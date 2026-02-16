package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/goccy/go-yaml"
)

const primaryWorldName = "mainhall"

func (a app) initRuntime() error {
	runtimeDir := filepath.Join(a.wslDir, "runtime", "world")
	if err := os.MkdirAll(runtimeDir, 0o755); err != nil {
		return err
	}
	fmt.Printf("Initialized: %s\n", filepath.Join(a.wslDir, "runtime"))
	return nil
}

func (a app) stageAssets() error {
	srcDir := filepath.Join(a.wslDir, "datapacks", "world-base")
	dstRoot := filepath.Join(a.wslDir, "runtime", "world", "mainhall", "datapacks")
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

	fmt.Printf("staged assets to %s\n", dstDir)
	return nil
}

func (a app) worldEnsure() error {
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
	}
	if err := a.pruneMainhallExtraDimensions(); err != nil {
		return err
	}

	fmt.Printf("ensured worlds from %s\n", filepath.Join(a.wslDir, "worlds"))
	return nil
}

func (a app) worldRegenerate(target string) error {
	target = strings.TrimSpace(target)
	if target == "" {
		return errors.New("world name is required")
	}

	cfgPath := filepath.Join(a.wslDir, "worlds", target, "world.env.yml")
	cfg, err := loadWorldConfig(cfgPath)
	if err != nil {
		return err
	}
	if target == primaryWorldName {
		return fmt.Errorf("world '%s' cannot be regenerated", primaryWorldName)
	}
	if !cfg.Deletable {
		return fmt.Errorf("world '%s' is not deletable", target)
	}

	if err := a.worldDrop(target); err != nil {
		return err
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

	fmt.Printf("regenerated world '%s'\n", target)
	return nil
}

func (a app) worldDrop(target string) error {
	target = strings.TrimSpace(target)
	if target == "" {
		return errors.New("world name is required")
	}
	if target == primaryWorldName {
		return fmt.Errorf("world '%s' cannot be dropped", primaryWorldName)
	}
	registered, err := a.worldRegisteredInMultiverse(target)
	if err != nil {
		return err
	}
	if !registered {
		fmt.Printf("world '%s' is not registered; skipped drop\n", target)
		return nil
	}
	if err := a.sendConsole(fmt.Sprintf("mv unload %s --remove-players", target)); err != nil {
		return err
	}
	if err := a.sendConsole(fmt.Sprintf("mv remove %s", target)); err != nil {
		return err
	}
	fmt.Printf("dropped world '%s'\n", target)
	return nil
}

func (a app) worldDelete(target string, yes bool) error {
	target = strings.TrimSpace(target)
	if target == "" {
		return errors.New("world name is required")
	}
	if target == primaryWorldName {
		return fmt.Errorf("world '%s' cannot be deleted", primaryWorldName)
	}
	if !yes {
		return errors.New("delete requires --yes")
	}

	cfgPath := filepath.Join(a.wslDir, "worlds", target, "world.env.yml")
	cfg, err := loadWorldConfig(cfgPath)
	if err != nil {
		return err
	}
	if !cfg.Deletable {
		return fmt.Errorf("world '%s' is not deletable", target)
	}

	if err := a.worldDrop(target); err != nil {
		return err
	}
	paths := []string{
		filepath.Join(a.wslDir, "runtime", "world", target),
		filepath.Join(a.wslDir, "runtime", "world", target+"_nether"),
		filepath.Join(a.wslDir, "runtime", "world", target+"_the_end"),
	}
	for _, p := range paths {
		if err := os.RemoveAll(p); err != nil {
			return err
		}
	}
	fmt.Printf("deleted world '%s'\n", target)
	return nil
}

func (a app) worldSetup(target string) error {
	target = strings.TrimSpace(target)
	if target != "" {
		if target != primaryWorldName {
			cfgPath := filepath.Join(a.wslDir, "worlds", target, "world.env.yml")
			cfg, err := loadWorldConfig(cfgPath)
			if err != nil {
				return err
			}
			target = cfg.Name
		}
		if err := a.applyWorldSetupCommands(target); err != nil {
			return err
		}
		if err := a.applyWorldPolicy(target); err != nil {
			return err
		}
		if target == primaryWorldName {
			if _, err := a.syncMainhallPortalsConfig(); err != nil {
				return err
			}
			if err := a.pruneMainhallExtraDimensions(); err != nil {
				return err
			}
		}
		fmt.Printf("setup world '%s'\n", target)
		return nil
	}

	if err := a.applyWorldSetupCommands(primaryWorldName); err != nil {
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
		if cfg.Name == primaryWorldName {
			continue
		}
		if err := a.applyWorldSetupCommands(cfg.Name); err != nil {
			return err
		}
		if err := a.applyWorldPolicy(cfg.Name); err != nil {
			return err
		}
	}
	if err := a.applyWorldPolicy(primaryWorldName); err != nil {
		return err
	}
	if _, err := a.syncMainhallPortalsConfig(); err != nil {
		return err
	}
	if err := a.pruneMainhallExtraDimensions(); err != nil {
		return err
	}
	fmt.Printf("setup worlds from %s\n", filepath.Join(a.wslDir, "worlds"))
	return nil
}

func (a app) worldFunctionRun(functionID string) error {
	functionID = strings.TrimSpace(functionID)
	if functionID == "" {
		return errors.New("function id is required")
	}
	return a.sendConsole("function " + functionID)
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

func worldDimensionID(worldName string) string {
	if worldName == primaryWorldName {
		return "minecraft:overworld"
	}
	return "minecraft:" + worldName
}

func (a app) loadWorldSetupCommands(worldName string) ([]string, bool, error) {
	path := filepath.Join(a.wslDir, "worlds", worldName, "setup.commands")
	if !fileExists(path) {
		return nil, false, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, false, err
	}
	defer f.Close()

	var commands []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "/")
		commands = append(commands, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, false, err
	}
	return commands, true, nil
}

func (a app) applyWorldSetupCommands(worldName string) error {
	commands, ok, err := a.loadWorldSetupCommands(worldName)
	if err != nil {
		return err
	}
	if !ok || len(commands) == 0 {
		return nil
	}
	dimension := worldDimensionID(worldName)
	for _, c := range commands {
		if err := a.sendConsole(fmt.Sprintf("execute in %s run %s", dimension, c)); err != nil {
			return err
		}
	}
	return nil
}

func (a app) applyWorldPolicy(worldName string) error {
	policy, ok, err := a.loadWorldPolicy(worldName)
	if err != nil {
		return err
	}
	if !ok || len(policy.MVSet) == 0 {
		return nil
	}

	var keys []string
	for k := range policy.MVSet {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		val := strings.TrimSpace(policy.MVSet[key])
		if val == "" {
			continue
		}
		if err := a.sendConsole(fmt.Sprintf("mv modify %s set %s %s", worldName, key, val)); err != nil {
			return err
		}
	}
	return nil
}

func (a app) pruneMainhallExtraDimensions() error {
	for _, name := range []string{primaryWorldName + "_nether", primaryWorldName + "_the_end"} {
		if err := a.worldDrop(name); err != nil {
			return err
		}
	}
	return nil
}

func (a app) syncMainhallPortalsConfig() (bool, error) {
	src := filepath.Join(a.wslDir, "worlds", primaryWorldName, "portals.yml")
	if !fileExists(src) {
		return false, nil
	}
	dst := filepath.Join(a.wslDir, "runtime", "world", "plugins", "Multiverse-Portals", "portals.yml")
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return false, err
	}
	in, err := os.Open(src)
	if err != nil {
		return false, err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return false, err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return false, err
	}
	fmt.Printf("synced mainhall portals config: %s -> %s\n", src, dst)
	return true, nil
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

func (a app) loadWorldPolicy(worldName string) (worldPolicy, bool, error) {
	path := filepath.Join(a.wslDir, "worlds", worldName, "world.policy.yml")
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
