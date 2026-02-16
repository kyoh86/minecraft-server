package main

import "github.com/spf13/cobra"

func newSetupCmd(a app) *cobra.Command {
	cmd := &cobra.Command{Use: "setup", Short: "setup operations"}
	cmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "initialize setup/wsl/runtime directories",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.initRuntime()
		},
	})
	return cmd
}
