package main

import "github.com/spf13/cobra"

func newSpawnCmd(a app) *cobra.Command {
	cmd := &cobra.Command{Use: "spawn", Short: "spawn profile and layout operations"}

	var spawnProfileWorld string
	profileCmd := &cobra.Command{
		Use:   "profile",
		Short: "probe surface Y and set anchor/spawn for managed worlds",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.worldSpawnProfile(spawnProfileWorld)
		},
	}
	profileCmd.Flags().StringVar(&spawnProfileWorld, "world", "", "target world name")
	cmd.AddCommand(profileCmd)

	var spawnStageWorld string
	stageCmd := &cobra.Command{
		Use:   "stage",
		Short: "apply spawn runtime files and reload",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.worldSpawnStage(spawnStageWorld)
		},
	}
	stageCmd.Flags().StringVar(&spawnStageWorld, "world", "", "target world name")
	cmd.AddCommand(stageCmd)

	var spawnApplyWorld string
	applyCmd := &cobra.Command{
		Use:   "apply",
		Short: "apply hub layout function at profiled spawn coordinates",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.worldSpawnApply(spawnApplyWorld)
		},
	}
	applyCmd.Flags().StringVar(&spawnApplyWorld, "world", "", "target world name")
	cmd.AddCommand(applyCmd)

	return cmd
}
