package main

import "github.com/spf13/cobra"

func newInitCmd(a app) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "initialize setup/wsl/runtime directories",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.initRuntime()
		},
	}
}
