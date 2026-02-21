package main

import "github.com/spf13/cobra"

func newRootCmd(a app) *cobra.Command {
	root := &cobra.Command{
		Use:   "wslctl",
		Short: "WSL Minecraft server helper",
		Long:  "wslctl manages assets, server operations, worlds, and players.",
	}
	root.AddCommand(newAssetCmd(a))
	root.AddCommand(newServerCmd(a))
	root.AddCommand(newWorldCmd(a))
	root.AddCommand(newPlayerCmd(a))
	root.AddCommand(newLinkCmd(a))
	return root
}
