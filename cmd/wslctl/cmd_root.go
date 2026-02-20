package main

import "github.com/spf13/cobra"

func newRootCmd(a app) *cobra.Command {
	root := &cobra.Command{
		Use:   "wslctl",
		Short: "WSL Minecraft server helper",
		Long:  "wslctl manages setup, server operations, worlds, and players.",
	}
	root.AddCommand(newSetupCmd(a))
	root.AddCommand(newServerCmd(a))
	root.AddCommand(newWorldCmd(a))
	root.AddCommand(newPlayerCmd(a))
	root.AddCommand(newLinkCmd(a))
	return root
}
