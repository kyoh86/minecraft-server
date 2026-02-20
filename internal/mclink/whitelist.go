package mclink

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
)

type WhitelistEntry struct {
	Type string `toml:"type"`
	Nick string `toml:"nick,omitempty"`
	UUID string `toml:"uuid,omitempty"`
}

type WhitelistFile struct {
	Enabled   bool             `toml:"enabled"`
	Servers   []string         `toml:"servers"`
	Whitelist []WhitelistEntry `toml:"whitelist"`
}

func AddWhitelistEntry(path string, typ EntryType, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New("value must not be empty")
	}

	var cfg WhitelistFile
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := toml.Unmarshal(b, &cfg); err != nil {
		return err
	}

	if hasEntry(cfg.Whitelist, typ, value) {
		return nil
	}

	entry := WhitelistEntry{Type: string(typ)}
	switch typ {
	case EntryTypeNick:
		entry.Nick = value
	case EntryTypeUUID:
		entry.UUID = value
	default:
		return errors.New("unsupported entry type")
	}
	cfg.Whitelist = append(cfg.Whitelist, entry)

	out, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o644)
}

func hasEntry(entries []WhitelistEntry, typ EntryType, value string) bool {
	for _, e := range entries {
		if e.Type != string(typ) {
			continue
		}
		switch typ {
		case EntryTypeNick:
			if strings.EqualFold(e.Nick, value) {
				return true
			}
		case EntryTypeUUID:
			if strings.EqualFold(e.UUID, value) {
				return true
			}
		}
	}
	return false
}
