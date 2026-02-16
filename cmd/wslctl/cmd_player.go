package main

import (
	"strings"

	"github.com/spf13/cobra"
)

func newPlayerCmd(a app) *cobra.Command {
	cmd := &cobra.Command{Use: "player", Short: "player operations"}

	cmd.AddCommand(&cobra.Command{
		Use:   "op <name>",
		Short: "grant operator to a player",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.sendConsole("op " + strings.TrimSpace(args[0]))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "deop <name>",
		Short: "revoke operator from a player",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.sendConsole("deop " + strings.TrimSpace(args[0]))
		},
	})

	return cmd
}
