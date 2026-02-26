package main

import (
	"strings"

	"github.com/spf13/cobra"
)

func newPlayerCmd(a app) *cobra.Command {
	cmd := &cobra.Command{Use: "player", Short: "player operations"}

	opCmd := &cobra.Command{Use: "op", Short: "operator permission operations"}
	opCmd.AddCommand(&cobra.Command{
		Use:   "grant <name>",
		Short: "grant operator to a player",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			player := strings.TrimSpace(args[0])
			if err := validatePlayerName(player); err != nil {
				return err
			}
			return a.sendConsole("op " + player)
		},
	})
	opCmd.AddCommand(&cobra.Command{
		Use:   "revoke <name>",
		Short: "revoke operator from a player",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			player := strings.TrimSpace(args[0])
			if err := validatePlayerName(player); err != nil {
				return err
			}
			return a.sendConsole("deop " + player)
		},
	})
	cmd.AddCommand(opCmd)

	adminCmd := &cobra.Command{Use: "admin", Short: "admin group operations"}
	adminCmd.AddCommand(&cobra.Command{
		Use:   "grant <name>",
		Short: "grant admin group using LuckPerms",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			player := strings.TrimSpace(args[0])
			if err := validatePlayerName(player); err != nil {
				return err
			}
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
	adminCmd.AddCommand(&cobra.Command{
		Use:   "revoke <name>",
		Short: "revoke admin group using LuckPerms",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			player := strings.TrimSpace(args[0])
			if err := validatePlayerName(player); err != nil {
				return err
			}
			return a.sendConsole("lp user " + player + " parent remove admin")
		},
	})
	cmd.AddCommand(adminCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "delink [uuid]",
		Short: "remove one entry from velocity allowlist (interactive if uuid omitted)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := ""
			if len(args) == 1 {
				target = strings.TrimSpace(args[0])
			}
			return a.playerDelink(target)
		},
	})

	return cmd
}
