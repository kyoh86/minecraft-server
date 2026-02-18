package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/goccy/go-yaml"
)

const primaryWorldName = "mainhall"
const spawnProfilePath = "runtime/world/.wslctl/spawn-profile.yml"

func (a app) initRuntime() error {
	runtimeDir := filepath.Join(a.baseDir, "runtime", "world")
	if err := os.MkdirAll(runtimeDir, 0o755); err != nil {
		return err
	}
	fmt.Printf("Initialized: %s\n", filepath.Join(a.baseDir, "runtime"))
	return nil
}

type spawnProfile struct {
	Worlds map[string]spawnProfileWorld `yaml:"worlds"`
}

type spawnProfileWorld struct {
	SurfaceY int `yaml:"surface_y"`
	AnchorY  int `yaml:"anchor_y"`
}

type spawnTemplateData struct {
	Worlds map[string]spawnTemplateWorld
}

type spawnTemplateWorld struct {
	SurfaceY       int
	AnchorY        int
	ReturnGateMinY int
	ReturnGateMaxY int
	RegionMinY     int
	RegionMaxY     int
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

	fmt.Printf("ensured worlds from %s\n", filepath.Join(a.baseDir, "worlds"))
	return nil
}

func (a app) worldRegenerate(target string) error {
	target = strings.TrimSpace(target)
	if target == "" {
		return errors.New("world name is required")
	}

	cfgPath := filepath.Join(a.baseDir, "worlds", target, "world.env.yml")
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
		filepath.Join(a.baseDir, "runtime", "world", target),
		filepath.Join(a.baseDir, "runtime", "world", target+"_nether"),
		filepath.Join(a.baseDir, "runtime", "world", target+"_the_end"),
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

	cfgPath := filepath.Join(a.baseDir, "worlds", target, "world.env.yml")
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
		filepath.Join(a.baseDir, "runtime", "world", target),
		filepath.Join(a.baseDir, "runtime", "world", target+"_nether"),
		filepath.Join(a.baseDir, "runtime", "world", target+"_the_end"),
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
	if err := a.ensureRuntimeDatapackScaffold(); err != nil {
		return err
	}
	if target != "" {
		if target != primaryWorldName {
			cfgPath := filepath.Join(a.baseDir, "worlds", target, "world.env.yml")
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
	if err := a.pruneMainhallExtraDimensions(); err != nil {
		return err
	}
	fmt.Printf("setup worlds from %s\n", filepath.Join(a.baseDir, "worlds"))
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

		onDisk := fileExists(filepath.Join(a.baseDir, "runtime", "world", cfg.Name))
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
	path := filepath.Join(a.baseDir, "worlds", worldName, "setup.commands")
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

func (a app) worldSpawnProfile() error {
	worldNames, err := a.listManagedWorldNames()
	if err != nil {
		return err
	}
	profile := spawnProfile{Worlds: map[string]spawnProfileWorld{}}
	for _, worldName := range worldNames {
		y, ok, err := a.resolveWorldSurfaceY(worldName)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("surface y could not be resolved for world %q", worldName)
		}
		anchorY := y - 32
		dimension := worldDimensionID(worldName)
		worldTag := "mcserver_spawn_anchor_" + worldName
		if err := a.sendConsole(fmt.Sprintf("execute in %s run forceload add 0 0", dimension)); err != nil {
			return err
		}
		if err := a.sendConsole(fmt.Sprintf("execute in %s run kill @e[type=minecraft:marker,tag=%s]", dimension, worldTag)); err != nil {
			return err
		}
		if err := a.sendConsole(fmt.Sprintf("execute in %s run summon minecraft:marker 0 %d 0 {Tags:[\"mcserver_spawn_anchor\",\"%s\"],NoGravity:1b,Invisible:1b,Invulnerable:1b}", dimension, anchorY, worldTag)); err != nil {
			return err
		}
		if err := a.sendConsole(fmt.Sprintf("execute in %s run setworldspawn 0 %d 0", dimension, y)); err != nil {
			return err
		}
		if err := a.sendConsole(fmt.Sprintf("mvsetspawn %s:0,%d,0 --unsafe", worldName, y)); err != nil {
			return err
		}
		if err := a.sendConsole(fmt.Sprintf("execute in %s run forceload remove 0 0", dimension)); err != nil {
			return err
		}
		profile.Worlds[worldName] = spawnProfileWorld{
			SurfaceY: y,
			AnchorY:  anchorY,
		}
	}
	if err := a.saveSpawnProfile(profile); err != nil {
		return err
	}
	fmt.Printf("profiled world spawn data for %d worlds\n", len(profile.Worlds))
	return nil
}

func (a app) worldSpawnStage() error {
	worldNames, err := a.listManagedWorldNames()
	if err != nil {
		return err
	}
	profile, err := a.loadSpawnProfile()
	if err != nil {
		return err
	}
	data, err := buildSpawnTemplateData(worldNames, profile)
	if err != nil {
		return err
	}
	if err := a.ensureRuntimeDatapackScaffold(); err != nil {
		return err
	}
	for _, worldName := range worldNames {
		src := filepath.Join(a.baseDir, "worlds", worldName, "worldguard.regions.yml.tmpl")
		dst := filepath.Join(a.baseDir, "runtime", "world", "plugins", "WorldGuard", "worlds", worldName, "regions.yml")
		if err := renderTemplateFile(src, dst, data); err != nil {
			return err
		}
	}
	portalsSrc := filepath.Join(a.baseDir, "worlds", primaryWorldName, "portals.yml.tmpl")
	portalsDst := filepath.Join(a.baseDir, "runtime", "world", "plugins", "Multiverse-Portals", "portals.yml")
	if err := renderTemplateFile(portalsSrc, portalsDst, data); err != nil {
		return err
	}
	if err := a.sendConsole("reload"); err != nil {
		return err
	}
	if err := a.sendConsole("wg reload"); err != nil {
		return err
	}
	if err := a.sendConsole("mvp config enforce-portal-access false"); err != nil {
		return err
	}
	if err := a.sendConsole("mv reload"); err != nil {
		return err
	}
	fmt.Printf("staged world spawn runtime configs for %d worlds\n", len(worldNames))
	return nil
}

func (a app) worldSpawnApply() error {
	worldNames, err := a.listManagedWorldNames()
	if err != nil {
		return err
	}
	profile, err := a.loadSpawnProfile()
	if err != nil {
		return err
	}
	if _, err := buildSpawnTemplateData(worldNames, profile); err != nil {
		return err
	}
	if err := a.sendConsole("reload"); err != nil {
		return err
	}
	if err := a.sendConsole("execute in minecraft:overworld run forceload add -1 -1 0 0"); err != nil {
		return err
	}
	if err := a.sendConsole("execute in minecraft:overworld run function mcserver:mainhall/hub_layout"); err != nil {
		return err
	}
	if err := a.sendConsole("execute in minecraft:overworld run forceload remove -1 -1 0 0"); err != nil {
		return err
	}
	for _, worldName := range worldNames {
		p := profile.Worlds[worldName]
		dimension := worldDimensionID(worldName)
		if err := a.sendConsole(fmt.Sprintf("execute in %s run forceload add 0 0", dimension)); err != nil {
			return err
		}
		if err := a.sendConsole(fmt.Sprintf("execute in %s run execute positioned 0 %d 0 run function mcserver:world/hub_layout", dimension, p.SurfaceY)); err != nil {
			return err
		}
		if err := a.sendConsole(fmt.Sprintf("execute in %s run forceload remove 0 0", dimension)); err != nil {
			return err
		}
	}
	fmt.Printf("applied spawn layout for %d worlds\n", len(worldNames)+1)
	return nil
}

func (a app) resolveWorldSurfaceY(worldName string) (int, bool, error) {
	dimension := worldDimensionID(worldName)
	composeFile := a.composeFilePath()
	tag := fmt.Sprintf("mcserver_yprobe_%d", time.Now().UnixNano())
	re := regexp.MustCompile(`Marker has the following entity data:\s*(-?\d+(?:\.\d+)?)d?`)

	if err := a.sendConsole(fmt.Sprintf("execute in %s run forceload add 0 0", dimension)); err != nil {
		return 0, false, err
	}
	defer func() { _ = a.sendConsole(fmt.Sprintf("execute in %s run forceload remove 0 0", dimension)) }()

	if err := a.sendConsole(fmt.Sprintf("execute in %s run summon minecraft:marker 0 0 0 {Tags:[\"%s\"]}", dimension, tag)); err != nil {
		return 0, false, err
	}
	defer func() {
		_ = a.sendConsole(fmt.Sprintf("execute in %s run kill @e[type=minecraft:marker,tag=%s]", dimension, tag))
	}()

	if err := a.sendConsole(fmt.Sprintf("execute in %s as @e[type=minecraft:marker,tag=%s,limit=1] at @s run execute positioned over motion_blocking_no_leaves run tp @s ~ ~ ~", dimension, tag)); err != nil {
		return 0, false, err
	}
	if err := a.sendConsole(fmt.Sprintf("execute in %s run data get entity @e[type=minecraft:marker,tag=%s,limit=1] Pos[1]", dimension, tag)); err != nil {
		return 0, false, err
	}
	time.Sleep(300 * time.Millisecond)
	out, err := runCommandOutput("docker", "compose", "-f", composeFile, "logs", "--since=5s", "world")
	if err != nil {
		return 0, false, err
	}
	matches := re.FindAllStringSubmatch(out, -1)
	if len(matches) == 0 {
		return 0, false, nil
	}
	last := matches[len(matches)-1]
	yf, err := strconv.ParseFloat(last[1], 64)
	if err != nil {
		return 0, false, err
	}
	return int(yf), true, nil
}

func (a app) ensureRuntimeDatapackScaffold() error {
	srcDir := filepath.Join(a.baseDir, "datapacks", "world-base")
	dstRoot := filepath.Join(a.baseDir, "runtime", "world", primaryWorldName, "datapacks")
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
	return nil
}

func (a app) listManagedWorldNames() ([]string, error) {
	cfgs, err := a.listWorldConfigs()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(cfgs))
	for _, cfgPath := range cfgs {
		cfg, err := loadWorldConfig(cfgPath)
		if err != nil {
			return nil, err
		}
		if cfg.Name == primaryWorldName {
			continue
		}
		names = append(names, cfg.Name)
	}
	sort.Strings(names)
	return names, nil
}

func (a app) loadSpawnProfile() (spawnProfile, error) {
	path := filepath.Join(a.baseDir, spawnProfilePath)
	if !fileExists(path) {
		return spawnProfile{}, fmt.Errorf("spawn profile not found: run `wslctl world spawn profile` first")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return spawnProfile{}, err
	}
	var p spawnProfile
	if err := yaml.Unmarshal(b, &p); err != nil {
		return spawnProfile{}, fmt.Errorf("parse spawn profile %s: %w", path, err)
	}
	if p.Worlds == nil {
		p.Worlds = map[string]spawnProfileWorld{}
	}
	return p, nil
}

func (a app) saveSpawnProfile(p spawnProfile) error {
	path := filepath.Join(a.baseDir, spawnProfilePath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := yaml.Marshal(p)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func buildSpawnTemplateData(worldNames []string, profile spawnProfile) (spawnTemplateData, error) {
	data := spawnTemplateData{Worlds: map[string]spawnTemplateWorld{}}
	for _, worldName := range worldNames {
		p, ok := profile.Worlds[worldName]
		if !ok {
			return spawnTemplateData{}, fmt.Errorf("spawn profile for world %q is missing: run `wslctl world spawn profile` first", worldName)
		}
		data.Worlds[worldName] = spawnTemplateWorld{
			SurfaceY:       p.SurfaceY,
			AnchorY:        p.AnchorY,
			ReturnGateMinY: p.SurfaceY,
			ReturnGateMaxY: p.SurfaceY + 3,
			RegionMinY:     p.SurfaceY - 11,
			RegionMaxY:     p.SurfaceY + 17,
		}
	}
	return data, nil
}

func renderTemplateFile(src, dst string, data any) error {
	if !fileExists(src) {
		return fmt.Errorf("missing template file: %s", src)
	}
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	tpl, err := template.New(filepath.Base(src)).Option("missingkey=error").Parse(string(b))
	if err != nil {
		return fmt.Errorf("parse template %s: %w", src, err)
	}
	var out bytes.Buffer
	if err := tpl.Execute(&out, data); err != nil {
		return fmt.Errorf("render template %s: %w", src, err)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, out.Bytes(), 0o644)
}

func (a app) listWorldConfigs() ([]string, error) {
	root := filepath.Join(a.baseDir, "worlds")
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
	return cfg, nil
}

func (a app) worldRegisteredInMultiverse(worldName string) (bool, error) {
	path := filepath.Join(a.baseDir, "runtime", "world", "plugins", "Multiverse-Core", "worlds.yml")
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
