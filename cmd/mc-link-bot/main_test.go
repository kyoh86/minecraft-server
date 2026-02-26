package main

import (
	"testing"
	"time"
)

func TestParseCodeEntryHash(t *testing.T) {
	t.Parallel()
	entry, err := parseCodeEntryHash(map[string]string{
		"code":            "ABCD1234",
		"type":            "uuid",
		"value":           "11111111-1111-1111-1111-111111111111",
		"expires_unix":    "1735689600",
		"claimed":         "true",
		"claimed_by":      "user123",
		"claimed_at_unix": "1735689601",
	})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if entry.Code != "ABCD1234" {
		t.Fatalf("unexpected code: %s", entry.Code)
	}
	if !entry.Claimed {
		t.Fatal("expected claimed=true")
	}
	if got, want := entry.ClaimedAt, time.Unix(1735689601, 0).UTC(); !got.Equal(want) {
		t.Fatalf("unexpected claimed_at: got=%v want=%v", got, want)
	}
}
