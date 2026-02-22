package mclink

import (
	"errors"
	"io"
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
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer func() { _ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN) }()

	cfg := Allowlist{}
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	if b, err := io.ReadAll(f); err == nil {
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
	if err := f.Truncate(0); err != nil {
		return err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	if _, err := f.Write(out); err != nil {
		return err
	}
	if err := f.Chmod(0o644); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
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
