package main

import "github.com/spf13/cobra"

func newInitCmd(a app) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "initialize runtime directories and secrets",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.init()
		},
	}
}
