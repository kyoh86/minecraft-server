package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	reWorldName  = regexp.MustCompile(`^[a-z0-9_-]+$`)
	rePlayerName = regexp.MustCompile(`^[A-Za-z0-9_]{3,16}$`)
	reFunctionID = regexp.MustCompile(`^[0-9a-z_./:-]+$`)
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

func validateConsoleCommand(command string) error {
	command = strings.TrimSpace(command)
	if command == "" {
		return fmt.Errorf("empty console command")
	}
	if strings.ContainsAny(command, "\r\n\x00") {
		return fmt.Errorf("console command contains unsafe control characters")
	}
	return nil
}

func validateWorldName(name string) error {
	name = strings.TrimSpace(name)
	if !reWorldName.MatchString(name) {
		return fmt.Errorf("invalid world name: %q", name)
	}
	return nil
}

func validatePlayerName(name string) error {
	name = strings.TrimSpace(name)
	if !rePlayerName.MatchString(name) {
		return fmt.Errorf("invalid player name: %q", name)
	}
	return nil
}

func validateFunctionID(functionID string) error {
	functionID = strings.TrimSpace(functionID)
	if !reFunctionID.MatchString(functionID) {
		return fmt.Errorf("invalid function id: %q", functionID)
	}
	return nil
}
