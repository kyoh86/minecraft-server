package main

import "github.com/spf13/cobra"

func newAssetsCmd(a app) *cobra.Command {
	cmd := &cobra.Command{Use: "assets", Short: "assets operations"}
	cmd.AddCommand(&cobra.Command{
		Use:   "stage",
		Short: "stage assets into runtime (currently datapack only)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.stageAssets()
		},
	})
	return cmd
}
