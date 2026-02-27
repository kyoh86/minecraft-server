package mclink

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddAllowlistEntry(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "allowlist.yml")

	if err := addAllowlistEntryUnlocked(path, EntryTypeUUID, "11111111-1111-1111-1111-111111111111"); err != nil {
		t.Fatalf("failed to add uuid entry: %v", err)
	}
	if err := addAllowlistEntryUnlocked(path, EntryTypeUUID, "11111111-1111-1111-1111-111111111111"); err != nil {
		t.Fatalf("failed to add duplicate uuid entry: %v", err)
	}
	if err := addAllowlistEntryUnlocked(path, EntryTypeNick, "Steve"); err == nil {
		t.Fatalf("nick entry should be rejected")
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read allowlist: %v", err)
	}
	got := string(b)
	if strings.Count(strings.ToLower(got), "11111111-1111-1111-1111-111111111111") != 1 {
		t.Fatalf("uuid entry was not deduplicated: %s", got)
	}
	if strings.Contains(got, "nicks:") {
		t.Fatalf("nicks key should not be written: %s", got)
	}
}
