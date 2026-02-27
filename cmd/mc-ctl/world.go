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
	"github.com/pelletier/go-toml/v2"
)

const (
	primaryWorldName = "mainhall"
	spawnProfilePath = "runtime/world/.mc-ctl/spawn-profile.toml"
	hubSchematicName = "hub.schem"
)

var reHexDisplayColor = regexp.MustCompile(`^#[0-9a-f]{6}$`)

type spawnProfile struct {
	Worlds map[string]spawnProfileWorld `toml:"worlds"`
}

type spawnProfileWorld struct {
	SurfaceY int `toml:"surface_y"`
	AnchorY  int `toml:"anchor_y"`
}

type spawnTemplateData struct {
	Worlds     map[string]spawnTemplateWorld
	WorldItems []spawnTemplateWorldItem
	Mainhall   mainhallLayout
}

type mainhallLayout struct {
	OuterMinX      int
	OuterMaxX      int
	InnerFloorMinX int
	InnerFloorMaxX int
	RoomMinX       int
	RoomMaxX       int
	RoomAirMinX    int
	RoomAirMaxX    int
	WallMinX       int
	WallMaxX       int
}

type spawnTemplateWorld struct {
	SurfaceY       int
	AnchorY        int
	ReturnGateMinY int
	ReturnGateMaxY int
	RegionMinY     int
	RegionMaxY     int
}

type spawnTemplateWorldItem struct {
	Name                string
	DisplayName         string
	DisplayColor        string
	MainhallGateMinX    int
	MainhallGateMaxX    int
	MainhallGateCenterX int
	MainhallFrameMinX   int
	MainhallFrameMaxX   int
	MainhallParticleX1  int
	MainhallParticleX2  int
	MainhallSignX       int
	MainhallGateMinY    int
	MainhallGateMaxY    int
	MainhallGateZ       int
	MainhallGateBackZ   int
	MainhallSignZ       int
	ReturnGateMinY      int
	ReturnGateMaxY      int
	ReturnGateMinX      int
	ReturnGateMaxX      int
	ReturnGateMinZ      int
	ReturnGateMaxZ      int
}

type portalConfig struct {
	Portals map[string]portalEntry `yaml:"portals"`
}

type portalEntry struct {
	Owner                  string `yaml:"owner"`
	Location               string `yaml:"location"`
	ActionType             string `yaml:"action-type"`
	Action                 string `yaml:"action"`
	SafeTeleport           bool   `yaml:"safe-teleport"`
	TeleportNonPlayers     bool   `yaml:"teleport-non-players"`
	CheckDestinationSafety bool   `yaml:"check-destination-safety"`
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
	if err := validateWorldName(target); err != nil {
		return err
	}

	cfgPath := filepath.Join(a.baseDir, "worlds", target, "config.toml")
	cfg, err := loadWorldConfig(cfgPath)
	if err != nil {
		return err
	}
	if cfg.Name == primaryWorldName {
		return fmt.Errorf("world '%s' cannot be regenerated", primaryWorldName)
	}
	if !cfg.Deletable {
		return fmt.Errorf("world '%s' is not deletable", target)
	}
	targets, err := managedDimensionTargets(cfg)
	if err != nil {
		return err
	}

	if err := a.worldDrop(cfg.Name); err != nil {
		return err
	}
	for _, t := range targets {
		if err := a.worldDrop(t.Name); err != nil {
			return err
		}
	}

	paths := []string{filepath.Join(a.baseDir, "runtime", "world", cfg.Name)}
	for _, t := range targets {
		paths = append(paths, filepath.Join(a.baseDir, "runtime", "world", t.Name))
	}
	for _, p := range paths {
		if err := os.RemoveAll(p); err != nil {
			return err
		}
	}

	if err := a.ensureWorld(cfg, true); err != nil {
		return err
	}

	fmt.Printf("regenerated world '%s'\n", cfg.Name)
	return nil
}

func (a app) worldDrop(target string) error {
	target = strings.TrimSpace(target)
	if target == "" {
		return errors.New("world name is required")
	}
	if err := validateWorldName(target); err != nil {
		return err
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
	if err := validateWorldName(target); err != nil {
		return err
	}
	if target == primaryWorldName {
		return fmt.Errorf("world '%s' cannot be deleted", primaryWorldName)
	}
	if !yes {
		return errors.New("delete requires --yes")
	}

	cfgPath := filepath.Join(a.baseDir, "worlds", target, "config.toml")
	cfg, err := loadWorldConfig(cfgPath)
	if err != nil {
		return err
	}
	if !cfg.Deletable {
		return fmt.Errorf("world '%s' is not deletable", target)
	}
	targets, err := managedDimensionTargets(cfg)
	if err != nil {
		return err
	}

	if err := a.worldDrop(cfg.Name); err != nil {
		return err
	}
	for _, t := range targets {
		if err := a.worldDrop(t.Name); err != nil {
			return err
		}
	}
	paths := []string{filepath.Join(a.baseDir, "runtime", "world", cfg.Name)}
	for _, t := range targets {
		paths = append(paths, filepath.Join(a.baseDir, "runtime", "world", t.Name))
	}
	for _, p := range paths {
		if err := os.RemoveAll(p); err != nil {
			return err
		}
	}
	fmt.Printf("deleted world '%s'\n", cfg.Name)
	return nil
}

func (a app) worldSetup(target string) error {
	target = strings.TrimSpace(target)
	if target != "" {
		if err := validateWorldName(target); err != nil {
			return err
		}
	}
	fmt.Println("world setup: syncing runtime datapack scaffold...")
	if err := a.ensureRuntimeDatapackScaffold(); err != nil {
		return err
	}
	if target != "" {
		if target != primaryWorldName {
			cfgPath := filepath.Join(a.baseDir, "worlds", target, "config.toml")
			cfg, err := loadWorldConfig(cfgPath)
			if err != nil {
				return err
			}
			target = cfg.Name
		}
		fmt.Printf("world setup: applying setup.commands to %s...\n", target)
		if err := a.applyWorldSetupCommands(target); err != nil {
			return err
		}
		fmt.Printf("world setup: applying config.toml to %s...\n", target)
		if err := a.applyWorldPolicy(target); err != nil {
			return err
		}
		fmt.Printf("setup world '%s'\n", target)
		return nil
	}

	fmt.Printf("world setup: applying setup.commands to %s...\n", primaryWorldName)
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
		fmt.Printf("world setup: applying setup.commands to %s...\n", cfg.Name)
		if err := a.applyWorldSetupCommands(cfg.Name); err != nil {
			return err
		}
		fmt.Printf("world setup: applying config.toml to %s...\n", cfg.Name)
		if err := a.applyWorldPolicy(cfg.Name); err != nil {
			return err
		}
	}
	fmt.Printf("world setup: applying config.toml to %s...\n", primaryWorldName)
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
	if err := validateFunctionID(functionID); err != nil {
		return err
	}
	return a.sendConsole("function " + functionID)
}

func (a app) ensureWorld(cfg worldConfig, forceCreate bool) error {
	if cfg.Name == "" {
		return fmt.Errorf("invalid world config: name is required")
	}
	if err := validateWorldName(cfg.Name); err != nil {
		return err
	}

	if err := a.ensureSingleWorld(cfg.Name, "normal", cfg.WorldType, cfg.Seed, forceCreate); err != nil {
		return err
	}
	targets, err := managedDimensionTargets(cfg)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		return nil
	}
	if err := a.ensureLinkedDimensions(targets, cfg.WorldType, cfg.Seed, forceCreate); err != nil {
		return err
	}
	if err := a.linkWorldDimensions(cfg.Name, targets); err != nil {
		return err
	}
	return nil
}

func (a app) ensureSingleWorld(worldName, environment, worldType string, seed any, forceCreate bool) error {
	registered, err := a.worldRegisteredInMultiverse(worldName)
	if err != nil {
		return err
	}
	onDisk := fileExists(filepath.Join(a.baseDir, "runtime", "world", worldName))

	if !forceCreate && registered && onDisk {
		return nil
	}
	if registered && !onDisk {
		fmt.Printf("world '%s' is registered but missing on disk; removing stale registration...\n", worldName)
		if err := a.sendConsole(fmt.Sprintf("mv remove %s", worldName)); err != nil {
			return err
		}
	}

	var ensureErr error
	if onDisk {
		ensureErr = a.sendConsole(fmt.Sprintf("mv import %s %s", worldName, environment))
	} else {
		parts := []string{"mv", "create", worldName, environment}
		if formattedSeed := formatSeed(seed); formattedSeed != "" {
			parts = append(parts, "-s", formattedSeed)
		}
		if worldType != "" {
			parts = append(parts, "-t", strings.ToUpper(worldType))
		}
		ensureErr = a.sendConsole(strings.Join(parts, " "))
	}
	if ensureErr != nil {
		return ensureErr
	}
	if err := a.waitWorldEnsureReady(worldName, 10*time.Second); err != nil {
		return err
	}
	return nil
}

func (a app) ensureLinkedDimensions(targets []dimensionTarget, worldType string, baseSeed any, forceCreate bool) error {
	for _, target := range targets {
		seed := baseSeed
		if target.Seed != nil {
			seed = target.Seed
		}
		if err := a.ensureSingleWorld(target.Name, target.Environment, worldType, seed, forceCreate); err != nil {
			return err
		}
	}
	return nil
}

func (a app) linkWorldDimensions(worldName string, targets []dimensionTarget) error {
	for _, target := range targets {
		if err := a.sendConsole(fmt.Sprintf("mvnp link %s %s %s", target.LinkKind, worldName, target.Name)); err != nil {
			return err
		}
	}
	return nil
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
		if err := validateConsoleCommand(c); err != nil {
			return fmt.Errorf("invalid setup command in %s: %w", worldName, err)
		}
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
		if strings.ContainsAny(key, " \t\r\n\x00") || strings.ContainsAny(val, "\r\n\x00") {
			return fmt.Errorf("invalid world policy key/value: %q=%q", key, val)
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

func (a app) worldSpawnProfile(target string) error {
	if target != "" {
		if err := validateWorldName(target); err != nil {
			return err
		}
	}
	worldNames, err := a.listManagedWorldNames(target)
	if err != nil {
		return err
	}
	profile := spawnProfile{Worlds: map[string]spawnProfileWorld{}}
	if p, err := a.loadSpawnProfile(); err == nil {
		profile = p
	}
	for _, worldName := range worldNames {
		fmt.Printf("spawn profile: probing surface Y for %s...\n", worldName)
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
		fmt.Printf("spawn profile: applying anchor/spawn for %s (surface=%d anchor=%d)...\n", worldName, y, anchorY)
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
	fmt.Printf("profiled spawn data for %d worlds\n", len(profile.Worlds))
	return nil
}

func (a app) worldSpawnStage(target string) error {
	if target != "" {
		if err := validateWorldName(target); err != nil {
			return err
		}
	}
	targetWorlds, err := a.listManagedWorldNames(target)
	if err != nil {
		return err
	}
	worldNames, err := a.listManagedWorldNames("")
	if err != nil {
		return err
	}
	profile, err := a.loadSpawnProfile()
	if err != nil {
		return err
	}
	worldCfgs, err := a.loadWorldConfigsByNames(worldNames)
	if err != nil {
		return err
	}
	data, err := buildSpawnTemplateData(worldNames, worldCfgs, profile)
	if err != nil {
		return err
	}
	if err := a.ensureRuntimeDatapackScaffold(); err != nil {
		return err
	}
	if err := a.ensureRuntimeHubSchematic(); err != nil {
		return err
	}
	worldGuardTargets := make([]string, 0, len(targetWorlds)+1)
	if target == "" {
		worldGuardTargets = append(worldGuardTargets, primaryWorldName)
	}
	worldGuardTargets = append(worldGuardTargets, targetWorlds...)
	for _, worldName := range worldGuardTargets {
		fmt.Printf("spawn stage: rendering WorldGuard regions for %s...\n", worldName)
		if err := a.renderWorldGuardRegions(worldName, data); err != nil {
			return err
		}
	}
	if target == "" {
		fmt.Println("spawn stage: rendering portals.yml...")
		portalsSrc := filepath.Join(a.baseDir, "worlds", primaryWorldName, "portals.yml.tmpl")
		portalsDst := filepath.Join(a.baseDir, "runtime", "world", "plugins", "Multiverse-Portals", "portals.yml")
		if err := renderTemplateFile(portalsSrc, portalsDst, data); err != nil {
			return err
		}
		fmt.Println("spawn stage: rendering mainhall hub layout...")
		hubLayoutSrc := filepath.Join(a.baseDir, "worlds", primaryWorldName, "hub_layout.mcfunction.tmpl")
		hubLayoutDst := filepath.Join(a.baseDir, "runtime", "world", primaryWorldName, "datapacks", "world-base", "data", "mcserver", "function", "mainhall", "hub_layout.mcfunction")
		if err := renderTemplateFile(hubLayoutSrc, hubLayoutDst, data); err != nil {
			return err
		}
	} else {
		fmt.Println("spawn stage: patching portals.yml for target world...")
		if err := a.patchPortalsForWorlds(targetWorlds); err != nil {
			return err
		}
	}
	fmt.Println("spawn stage: reloading server/plugins...")
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
	fmt.Printf("staged spawn runtime configs for %d worlds (worldguard targets=%d)\n", len(targetWorlds), len(worldGuardTargets))
	return nil
}

func (a app) renderWorldGuardRegions(worldName string, data spawnTemplateData) error {
	src, err := a.resolveWorldGuardRegionsTemplate(worldName)
	if err != nil {
		return err
	}
	dst := filepath.Join(a.baseDir, "runtime", "world", "plugins", "WorldGuard", "worlds", worldName, "regions.yml")
	tplData := map[string]any{
		"WorldName": worldName,
		"Worlds":    data.Worlds,
	}
	return renderTemplateFile(src, dst, tplData)
}

func (a app) resolveWorldGuardRegionsTemplate(worldName string) (string, error) {
	byWorld := filepath.Join(a.baseDir, "worlds", worldName, "worldguard.regions.yml.tmpl")
	if fileExists(byWorld) {
		return byWorld, nil
	}
	if worldName == primaryWorldName {
		return "", fmt.Errorf("missing worldguard template for %q: %s", worldName, byWorld)
	}
	def := filepath.Join(a.baseDir, "worlds", "_defaults", "worldguard.regions.yml.tmpl")
	if fileExists(def) {
		return def, nil
	}
	return "", fmt.Errorf("missing worldguard template for %q: %s (and default %s)", worldName, byWorld, def)
}

func (a app) worldSpawnApply(target string) error {
	target = strings.TrimSpace(target)
	if target != "" {
		if err := validateWorldName(target); err != nil {
			return err
		}
	}
	worldNames, err := a.listManagedWorldNames(target)
	if err != nil {
		return err
	}
	profile, err := a.loadSpawnProfile()
	if err != nil {
		return err
	}
	worldCfgs, err := a.loadWorldConfigsByNames(worldNames)
	if err != nil {
		return err
	}
	if _, err := buildSpawnTemplateData(worldNames, worldCfgs, profile); err != nil {
		return err
	}
	if err := a.ensureRuntimeHubSchematic(); err != nil {
		return err
	}
	fmt.Println("spawn apply: reloading datapacks...")
	if err := a.sendConsole("reload"); err != nil {
		return err
	}
	appliedMainhall := false
	if target == "" {
		fmt.Println("spawn apply: applying mainhall hub layout...")
		if err := a.sendConsole("execute in minecraft:overworld run forceload add -1 -1 0 0"); err != nil {
			return err
		}
		if err := a.sendConsole("execute in minecraft:overworld run function mcserver:mainhall/hub_layout"); err != nil {
			return err
		}
		if err := a.sendConsole("execute in minecraft:overworld run forceload remove -1 -1 0 0"); err != nil {
			return err
		}
		appliedMainhall = true
	}
	for _, worldName := range worldNames {
		p := profile.Worlds[worldName]
		dimension := worldDimensionID(worldName)
		fmt.Printf("spawn apply: applying %s hub schematic at y=%d...\n", worldName, p.SurfaceY)
		if err := a.sendConsole(fmt.Sprintf("execute in %s run forceload add -64 -64 64 64", dimension)); err != nil {
			return err
		}
		if err := a.sendConsole(fmt.Sprintf("hubterraform apply %s %d", worldName, p.SurfaceY)); err != nil {
			return err
		}
		if err := a.applyWorldHubSchematic(worldName, p.SurfaceY); err != nil {
			return err
		}
		if err := a.sendConsole(fmt.Sprintf("execute in %s run forceload remove -64 -64 64 64", dimension)); err != nil {
			return err
		}
	}
	applied := len(worldNames)
	if appliedMainhall {
		applied++
	}
	fmt.Printf("applied spawn layout for %d worlds\n", applied)
	return nil
}

func (a app) resolveWorldSurfaceY(worldName string) (int, bool, error) {
	re := regexp.MustCompile(`hubterraform probe: world=([A-Za-z0-9_:-]+) surfaceY=(-?\d+) samples=(\d+) median=(-?\d+) p40=(-?\d+) meanFloor=(-?\d+)`)
	if err := a.sendConsole(fmt.Sprintf("hubterraform probe %s", worldName)); err != nil {
		return 0, false, err
	}
	for i := 0; i < 40; i++ {
		y, ok, err := a.readLastSurfaceProbeY(worldName, re)
		if err != nil {
			return 0, false, err
		}
		if ok {
			return y, true, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return 0, false, nil
}

func (a app) readLastSurfaceProbeY(worldName string, re *regexp.Regexp) (int, bool, error) {
	out, err := a.composeOutput("logs", "--since=5s", "world")
	if err != nil {
		return 0, false, err
	}
	matches := re.FindAllStringSubmatch(out, -1)
	if len(matches) == 0 {
		return 0, false, nil
	}
	for i := len(matches) - 1; i >= 0; i-- {
		m := matches[i]
		if len(m) < 3 || m[1] != worldName {
			continue
		}
		y, err := strconv.Atoi(m[2])
		if err != nil {
			return 0, false, err
		}
		return y, true, nil
	}
	return 0, false, nil
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

func (a app) ensureRuntimeHubSchematic() error {
	src := filepath.Join(a.baseDir, "infra", "world", "schematics", hubSchematicName)
	if !fileExists(src) {
		return fmt.Errorf("missing hub schematic source: %s", src)
	}
	dst := filepath.Join(a.baseDir, "runtime", "world", "plugins", "FastAsyncWorldEdit", "schematics", hubSchematicName)
	return copyFile(src, dst)
}

func (a app) applyWorldHubSchematic(worldName string, surfaceY int) (err error) {
	if err = a.sendConsole("//world " + worldName); err != nil {
		return err
	}
	defer func() {
		if cleanupErr := a.sendConsole("//world"); cleanupErr != nil && err == nil {
			err = cleanupErr
		}
	}()
	if err = a.sendConsole("//schem load " + strings.TrimSuffix(hubSchematicName, filepath.Ext(hubSchematicName))); err != nil {
		return err
	}
	if err = a.sendConsole(fmt.Sprintf("//pos1 0,%d,0", surfaceY)); err != nil {
		return err
	}
	return a.sendConsole("//paste")
}

func (a app) listManagedWorldNames(target string) ([]string, error) {
	target = strings.TrimSpace(target)
	if target == primaryWorldName {
		return nil, fmt.Errorf("world '%s' is not managed by spawn commands", primaryWorldName)
	}
	cfgs, err := a.listWorldConfigs()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(cfgs))
	found := false
	for _, cfgPath := range cfgs {
		cfg, err := loadWorldConfig(cfgPath)
		if err != nil {
			return nil, err
		}
		if cfg.Name == primaryWorldName {
			continue
		}
		if target != "" && cfg.Name != target {
			continue
		}
		names = append(names, cfg.Name)
		found = true
	}
	if target != "" && !found {
		return nil, fmt.Errorf("managed world '%s' is not defined", target)
	}
	sort.Strings(names)
	return names, nil
}

func (a app) loadWorldConfigsByNames(worldNames []string) (map[string]worldConfig, error) {
	cfgs := make(map[string]worldConfig, len(worldNames))
	for _, worldName := range worldNames {
		cfgPath := filepath.Join(a.baseDir, "worlds", worldName, "config.toml")
		cfg, err := loadWorldConfig(cfgPath)
		if err != nil {
			return nil, err
		}
		cfgs[worldName] = cfg
	}
	return cfgs, nil
}

func (a app) loadSpawnProfile() (spawnProfile, error) {
	path := filepath.Join(a.baseDir, spawnProfilePath)
	if !fileExists(path) {
		return spawnProfile{}, fmt.Errorf("spawn profile not found: run `mc-ctl spawn profile` first")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return spawnProfile{}, err
	}
	var p spawnProfile
	if err := toml.Unmarshal(b, &p); err != nil {
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
	b, err := toml.Marshal(p)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func buildSpawnTemplateData(worldNames []string, cfgs map[string]worldConfig, profile spawnProfile) (spawnTemplateData, error) {
	data := spawnTemplateData{
		Worlds:     map[string]spawnTemplateWorld{},
		WorldItems: make([]spawnTemplateWorldItem, 0, len(worldNames)),
		Mainhall: mainhallLayout{
			OuterMinX:      -12,
			OuterMaxX:      12,
			InnerFloorMinX: -11,
			InnerFloorMaxX: 11,
			RoomMinX:       -8,
			RoomMaxX:       8,
			RoomAirMinX:    -7,
			RoomAirMaxX:    7,
			WallMinX:       -9,
			WallMaxX:       9,
		},
	}
	for i, worldName := range worldNames {
		cfg, ok := cfgs[worldName]
		if !ok {
			return spawnTemplateData{}, fmt.Errorf("world config for %q is missing", worldName)
		}
		p, ok := profile.Worlds[worldName]
		if !ok {
			return spawnTemplateData{}, fmt.Errorf("spawn profile for world %q is missing: run `mc-ctl spawn profile` first", worldName)
		}
		data.Worlds[worldName] = spawnTemplateWorld{
			SurfaceY:       p.SurfaceY,
			AnchorY:        p.AnchorY,
			ReturnGateMinY: p.SurfaceY,
			ReturnGateMaxY: p.SurfaceY + 3,
			RegionMinY:     p.SurfaceY - 8,
			RegionMaxY:     p.SurfaceY + 12,
		}
		returnGateCenterX := -1
		gateMinX, gateMaxX := mainhallGateXForIndex(i, len(worldNames))
		centerX := (gateMinX + gateMaxX) / 2
		data.WorldItems = append(data.WorldItems, spawnTemplateWorldItem{
			Name:                worldName,
			DisplayName:         normalizeDisplayName(cfg.DisplayName, worldName),
			DisplayColor:        normalizeDisplayColor(cfg.DisplayColor),
			MainhallGateMinX:    gateMinX,
			MainhallGateMaxX:    gateMaxX,
			MainhallGateCenterX: centerX,
			MainhallFrameMinX:   centerX - 2,
			MainhallFrameMaxX:   centerX + 2,
			MainhallParticleX1:  centerX - 2,
			MainhallParticleX2:  centerX - 1,
			MainhallSignX:       centerX,
			MainhallGateMinY:    -58,
			MainhallGateMaxY:    -56,
			MainhallGateZ:       -9,
			MainhallGateBackZ:   -10,
			MainhallSignZ:       -7,
			ReturnGateMinY:      p.SurfaceY,
			ReturnGateMaxY:      p.SurfaceY + 3,
			ReturnGateMinX:      returnGateCenterX,
			ReturnGateMaxX:      returnGateCenterX + 1,
			ReturnGateMinZ:      2,
			ReturnGateMaxZ:      3,
		})
	}
	if len(data.WorldItems) > 0 {
		gateMin := data.WorldItems[0].MainhallFrameMinX
		gateMax := data.WorldItems[0].MainhallFrameMaxX
		for _, w := range data.WorldItems[1:] {
			if w.MainhallFrameMinX < gateMin {
				gateMin = w.MainhallFrameMinX
			}
			if w.MainhallFrameMaxX > gateMax {
				gateMax = w.MainhallFrameMaxX
			}
		}
		wallMin := minInt(-9, gateMin-1)
		wallMax := maxInt(9, gateMax+1)
		data.Mainhall = mainhallLayout{
			OuterMinX:      wallMin - 3,
			OuterMaxX:      wallMax + 3,
			InnerFloorMinX: wallMin - 2,
			InnerFloorMaxX: wallMax + 2,
			RoomMinX:       wallMin + 1,
			RoomMaxX:       wallMax - 1,
			RoomAirMinX:    wallMin + 2,
			RoomAirMaxX:    wallMax - 2,
			WallMinX:       wallMin,
			WallMaxX:       wallMax,
		}
	}
	return data, nil
}

func mainhallGateXForIndex(i, total int) (minX, maxX int) {
	if total <= 0 {
		return -1, 1
	}
	firstCenterX := -2 * (total - 1)
	centerX := firstCenterX + i*4
	minX = centerX - 1
	maxX = centerX + 1
	return minX, maxX
}

func normalizeDisplayName(displayName, fallback string) string {
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		return fallback
	}
	return displayName
}

func normalizeDisplayColor(color string) string {
	color = strings.TrimSpace(strings.ToLower(color))
	if color == "" {
		return "gold"
	}
	if reHexDisplayColor.MatchString(color) {
		return color
	}
	switch color {
	case "black", "dark_blue", "dark_green", "dark_aqua", "dark_red", "dark_purple", "gold",
		"gray", "dark_gray", "blue", "green", "aqua", "red", "light_purple", "yellow", "white":
		return color
	default:
		return "gold"
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (a app) patchPortalsForWorlds(targetWorlds []string) error {
	allWorlds, err := a.listManagedWorldNames("")
	if err != nil {
		return err
	}
	indexByName := map[string]int{}
	for i, name := range allWorlds {
		indexByName[name] = i
	}
	profile, err := a.loadSpawnProfile()
	if err != nil {
		return err
	}
	path := filepath.Join(a.baseDir, "runtime", "world", "plugins", "Multiverse-Portals", "portals.yml")
	if !fileExists(path) {
		return fmt.Errorf("missing portals runtime file: %s (run `mc-ctl spawn stage` without --world first)", path)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var cfg portalConfig
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return fmt.Errorf("parse portals yml %s: %w", path, err)
	}
	if cfg.Portals == nil {
		cfg.Portals = map[string]portalEntry{}
	}
	for _, worldName := range targetWorlds {
		idx, ok := indexByName[worldName]
		if !ok {
			return fmt.Errorf("managed world %q not found", worldName)
		}
		p, ok := profile.Worlds[worldName]
		if !ok {
			return fmt.Errorf("spawn profile for world %q is missing: run `mc-ctl spawn profile --world %s` first", worldName, worldName)
		}
		minX, maxX := mainhallGateXForIndex(idx, len(allWorlds))
		cfg.Portals["gate_"+worldName] = portalEntry{
			Owner:                  "console",
			Location:               fmt.Sprintf("mainhall:%d,-58,-9:%d,-56,-9", minX, maxX),
			ActionType:             "multiverse-destination",
			Action:                 "w:" + worldName,
			SafeTeleport:           true,
			TeleportNonPlayers:     false,
			CheckDestinationSafety: true,
		}
		cfg.Portals["gate_"+worldName+"_to_mainhall"] = portalEntry{
			Owner:                  "console",
			Location:               fmt.Sprintf("%s:-2,%d,2:0,%d,3", worldName, p.SurfaceY, p.SurfaceY+3),
			ActionType:             "multiverse-destination",
			Action:                 "w:mainhall",
			SafeTeleport:           true,
			TeleportNonPlayers:     false,
			CheckDestinationSafety: true,
		}
	}
	out, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o644)
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
		if e.Name() == primaryWorldName {
			continue
		}
		cfg := filepath.Join(root, e.Name(), "config.toml")
		if fileExists(cfg) {
			cfgs = append(cfgs, cfg)
		}
	}
	sort.Strings(cfgs)
	return cfgs, nil
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

func (a app) waitWorldEnsureReady(worldName string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	worldPath := filepath.Join(a.baseDir, "runtime", "world", worldName)
	for time.Now().Before(deadline) {
		onDisk := fileExists(worldPath)
		registered, err := a.worldRegisteredInMultiverse(worldName)
		if err != nil {
			return err
		}
		if onDisk && registered {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("world '%s' was not ensured (registered/on-disk check failed); check server logs", worldName)
}
