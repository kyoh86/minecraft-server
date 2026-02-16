package main

import "github.com/spf13/cobra"

func newRootCmd(a app) *cobra.Command {
	root := &cobra.Command{
		Use:   "wslctl",
		Short: "WSL Minecraft server helper",
		Long: "wslctl manages runtime directories, datapack sync, world bootstrap/reset, " +
			"player operator state, and LuckPerms admin grants for this repository.",
	}
	root.AddCommand(newInitCmd(a))
	root.AddCommand(newDatapackCmd(a))
	root.AddCommand(newWorldCmd(a))
	root.AddCommand(newPlayerCmd(a))
	root.AddCommand(newPermsCmd(a))
	return root
}
