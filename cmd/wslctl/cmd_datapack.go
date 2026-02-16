package main

import "github.com/spf13/cobra"

func newDatapackCmd(a app) *cobra.Command {
	cmd := &cobra.Command{Use: "datapack", Short: "datapack operations"}
	cmd.AddCommand(&cobra.Command{
		Use:   "sync",
		Short: "sync setup/wsl/datapacks/world-base into runtime/world/mainhall/datapacks",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.syncWorldDatapack()
		},
	})
	return cmd
}
