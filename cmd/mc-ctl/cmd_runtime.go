package main

import "github.com/spf13/cobra"

func newAssetCmd(a app) *cobra.Command {
	cmd := &cobra.Command{Use: "asset", Short: "asset operations"}
	cmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "initialize runtime asset directories",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.assetInit()
		},
	})
	return cmd
}
