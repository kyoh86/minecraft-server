package mclink

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"

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

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	lockPath := path + ".lock"
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}
	defer lockFile.Close()
	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer func() { _ = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN) }()

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
	tmpFile, err := os.CreateTemp(filepath.Dir(path), ".allowlist-*.yml")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	if _, err := tmpFile.Write(out); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := tmpFile.Chmod(0o644); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}

func containsFold(values []string, needle string) bool {
	for _, v := range values {
		if strings.EqualFold(v, needle) {
			return true
		}
	}
	return false
}
