package mclink

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type EntryType string

const (
	EntryTypeNick EntryType = "nick"
	EntryTypeUUID EntryType = "uuid"
)

type CodeEntry struct {
	Code      string    `json:"code"`
	Type      EntryType `json:"type"`
	Value     string    `json:"value"`
	ExpiresAt time.Time `json:"expires_at"`
	Claimed   bool      `json:"claimed"`
	ClaimedBy string    `json:"claimed_by,omitempty"`
	ClaimedAt time.Time `json:"claimed_at,omitempty"`
}

type Store struct {
	Codes map[string]CodeEntry `json:"codes"`
}

func LoadStore(path string) (Store, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Store{Codes: map[string]CodeEntry{}}, nil
		}
		return Store{}, err
	}

	var s Store
	if err := json.Unmarshal(b, &s); err != nil {
		return Store{}, err
	}
	if s.Codes == nil {
		s.Codes = map[string]CodeEntry{}
	}
	return s, nil
}

func SaveStore(path string, s Store) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
