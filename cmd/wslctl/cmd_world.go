package main

import "github.com/spf13/cobra"

func newWorldCmd(a app) *cobra.Command {
	cmd := &cobra.Command{Use: "world", Short: "world operations"}

	cmd.AddCommand(&cobra.Command{
		Use:   "ensure",
		Short: "create/import worlds from setup/wsl/worlds definitions",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.worldEnsure()
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "regenerate <name>",
		Short: "delete and recreate one world when deletable=true",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.worldRegenerate(args[0])
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "drop <name>",
		Short: "unload and unregister one world without deleting world files",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.worldDrop(args[0])
		},
	})

	var deleteYes bool
	deleteCmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "drop and delete one world files when deletable=true",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.worldDelete(args[0], deleteYes)
		},
	}
	deleteCmd.Flags().BoolVar(&deleteYes, "yes", false, "confirm destructive delete")
	cmd.AddCommand(deleteCmd)

	var worldName string
	setupCmd := &cobra.Command{
		Use:   "setup",
		Short: "apply setup.commands and world policy for all worlds or one world",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.worldSetup(worldName)
		},
	}
	setupCmd.Flags().StringVar(&worldName, "world", "", "target world name")
	cmd.AddCommand(setupCmd)

	functionCmd := &cobra.Command{Use: "function", Short: "world function operations"}
	functionCmd.AddCommand(&cobra.Command{
		Use:   "run <id>",
		Short: "run an arbitrary function id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.worldFunctionRun(args[0])
		},
	})
	cmd.AddCommand(functionCmd)

	return cmd
}
