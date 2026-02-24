package main

import "fmt"

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func versionString() string {
	return fmt.Sprintf("mc-ctl %s (commit: %s, date: %s)", version, commit, date)
}
