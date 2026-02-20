package mclink

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/goccy/go-yaml"
)

type Allowlist struct {
	UUIDs []string `yaml:"uuids"`
	Nicks []string `yaml:"nicks"`
}

func AddAllowlistEntry(path string, typ EntryType, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New("value must not be empty")
	}

	cfg := Allowlist{}
	if b, err := os.ReadFile(path); err == nil {
		if err := yaml.Unmarshal(b, &cfg); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	switch typ {
	case EntryTypeNick:
		if !containsFold(cfg.Nicks, value) {
			cfg.Nicks = append(cfg.Nicks, value)
		}
	case EntryTypeUUID:
		if !containsFold(cfg.UUIDs, value) {
			cfg.UUIDs = append(cfg.UUIDs, value)
		}
	default:
		return errors.New("unsupported entry type")
	}
	slices.Sort(cfg.Nicks)
	slices.Sort(cfg.UUIDs)

	out, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o644)
}

func containsFold(values []string, needle string) bool {
	for _, v := range values {
		if strings.EqualFold(v, needle) {
			return true
		}
	}
	return false
}
