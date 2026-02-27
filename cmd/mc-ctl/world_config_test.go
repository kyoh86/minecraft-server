package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadWorldConfig_DefaultWorldType(t *testing.T) {
	path := writeWorldConfigForTest(t, `
name = "testworld"
deletable = true
`)
	cfg, err := loadWorldConfig(path)
	if err != nil {
		t.Fatalf("loadWorldConfig failed: %v", err)
	}
	if cfg.WorldType != "normal" {
		t.Fatalf("expected default world_type=normal, got %q", cfg.WorldType)
	}
}

func TestLoadWorldConfig_RejectsNullLikeSeedLiteral(t *testing.T) {
	path := writeWorldConfigForTest(t, `
name = "testworld"
seed = null
deletable = true
`)
	_, err := loadWorldConfig(path)
	if err == nil {
		t.Fatal("expected parse error for seed = null, got nil")
	}
	if !strings.Contains(err.Error(), "parse world config") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadWorldConfig_RejectsEmptySeedString(t *testing.T) {
	path := writeWorldConfigForTest(t, `
name = "testworld"
seed = ""
deletable = true
`)
	_, err := loadWorldConfig(path)
	if err == nil {
		t.Fatal("expected error for empty seed string, got nil")
	}
	if !strings.Contains(err.Error(), "must not be empty string") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadWorldConfig_RejectsEmptyDimensionSeedString(t *testing.T) {
	path := writeWorldConfigForTest(t, `
name = "testworld"
deletable = true
[dimensions.nether]
seed = ""
`)
	_, err := loadWorldConfig(path)
	if err == nil {
		t.Fatal("expected error for empty dimension seed string, got nil")
	}
	if !strings.Contains(err.Error(), "must not be empty string") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func writeWorldConfigForTest(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)+"\n"), 0o644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	return path
}
