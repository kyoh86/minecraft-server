package main

import "github.com/spf13/cobra"

func newWorldCmd(a app) *cobra.Command {
	cmd := &cobra.Command{Use: "world", Short: "world operations"}

	cmd.AddCommand(&cobra.Command{
		Use:   "bootstrap",
		Short: "create/import all worlds from setup/wsl/worlds and run each init function",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.worldsBootstrap()
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "reset <name>",
		Short: "reset one world from setup/wsl/worlds/<name>/world.env.yml when resettable=true",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.worldReset(args[0])
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "apply-settings",
		Short: "reload and run function mcserver:world_settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.applyWorldSettings()
		},
	})

	return cmd
}
