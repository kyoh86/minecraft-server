package main

import (
	"strings"

	"github.com/spf13/cobra"
)

func newPermsCmd(a app) *cobra.Command {
	cmd := &cobra.Command{Use: "perms", Short: "LuckPerms operations"}

	cmd.AddCommand(&cobra.Command{
		Use:   "grant-admin <name>",
		Short: "grant admin group using LuckPerms",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			player := strings.TrimSpace(args[0])
			for _, c := range []string{
				"lp creategroup admin",
				"lp group admin permission set * true",
				"lp user " + player + " parent set admin",
			} {
				if err := a.sendConsole(c); err != nil {
					return err
				}
			}
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "revoke-admin <name>",
		Short: "remove admin group using LuckPerms",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.sendConsole("lp user " + strings.TrimSpace(args[0]) + " parent remove admin")
		},
	})

	return cmd
}
