package main

import "github.com/spf13/cobra"

func newRootCmd(a app) *cobra.Command {
	root := &cobra.Command{
		Use:   "mc-ctl",
		Short: "Minecraft server helper",
		Long:  "mc-ctl manages init, server operations, worlds, players, and link flows.",
	}
	root.AddCommand(newInitCmd(a))
	root.AddCommand(newServerCmd(a))
	root.AddCommand(newWorldCmd(a))
	root.AddCommand(newPlayerCmd(a))
	root.AddCommand(newLinkCmd(a))
	return root
}
