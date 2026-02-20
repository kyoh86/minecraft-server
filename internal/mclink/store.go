package mclink

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type EntryType string

const (
	EntryTypeNick EntryType = "nick"
	EntryTypeUUID EntryType = "uuid"
)

type CodeEntry struct {
	Code      string
	Type      EntryType
	Value     string
	ExpiresAt time.Time
	Claimed   bool
	ClaimedBy string
	ClaimedAt time.Time
}

type Store struct {
	Codes map[string]CodeEntry
}

func LoadStore(path string) (Store, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Store{Codes: map[string]CodeEntry{}}, nil
		}
		return Store{}, err
	}
	defer f.Close()

	store := Store{Codes: map[string]CodeEntry{}}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		e, ok := parseLine(line)
		if !ok {
			continue
		}
		store.Codes[e.Code] = e
	}
	if err := sc.Err(); err != nil {
		return Store{}, err
	}
	return store, nil
}

func SaveStore(path string, s Store) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	_, _ = w.WriteString("# code\ttype\tvalue\texpires_unix\tclaimed\tclaimed_by\tclaimed_at_unix\n")
	for _, e := range s.Codes {
		line := fmt.Sprintf("%s\t%s\t%s\t%d\t%t\t%s\t%d\n",
			e.Code,
			e.Type,
			e.Value,
			e.ExpiresAt.Unix(),
			e.Claimed,
			e.ClaimedBy,
			e.ClaimedAt.Unix(),
		)
		if _, err := w.WriteString(line); err != nil {
			return err
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func parseLine(line string) (CodeEntry, bool) {
	p := strings.Split(line, "\t")
	if len(p) < 7 {
		return CodeEntry{}, false
	}
	expiresUnix, err := strconv.ParseInt(p[3], 10, 64)
	if err != nil {
		return CodeEntry{}, false
	}
	claimed, err := strconv.ParseBool(p[4])
	if err != nil {
		return CodeEntry{}, false
	}
	claimedAtUnix, _ := strconv.ParseInt(p[6], 10, 64)
	return CodeEntry{
		Code:      p[0],
		Type:      EntryType(p[1]),
		Value:     p[2],
		ExpiresAt: time.Unix(expiresUnix, 0).UTC(),
		Claimed:   claimed,
		ClaimedBy: p[5],
		ClaimedAt: time.Unix(claimedAtUnix, 0).UTC(),
	}, true
}
