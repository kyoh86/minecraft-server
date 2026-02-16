package main

import "github.com/spf13/cobra"

func newRuntimeCmd(a app) *cobra.Command {
	cmd := &cobra.Command{Use: "runtime", Short: "runtime directory operations"}
	cmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "initialize setup/wsl/runtime directories",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.initRuntime()
		},
	})
	return cmd
}
