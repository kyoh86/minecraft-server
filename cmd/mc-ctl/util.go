package main

import (
	"fmt"
	"os"
	"strings"
)

func formatSeed(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case int:
		return fmt.Sprintf("%d", t)
	case int64:
		return fmt.Sprintf("%d", t)
	case uint64:
		return fmt.Sprintf("%d", t)
	case float64:
		return fmt.Sprintf("%.0f", t)
	case string:
		return strings.TrimSpace(t)
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", t))
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
