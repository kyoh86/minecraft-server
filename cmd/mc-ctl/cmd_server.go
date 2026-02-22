package main

import "github.com/spf13/cobra"

func newServerCmd(a app) *cobra.Command {
	cmd := &cobra.Command{Use: "server", Short: "server operations"}

	cmd.AddCommand(&cobra.Command{
		Use:   "up",
		Short: "start server containers",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.serverUp()
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "down",
		Short: "stop server containers",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.serverDown()
		},
	})

	var restartBuild bool
	restartCmd := &cobra.Command{
		Use:   "restart [service]",
		Short: "restart container service (default: world)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			service := "world"
			if len(args) == 1 {
				service = args[0]
			}
			return a.serverRestart(service, restartBuild)
		},
	}
	restartCmd.Flags().BoolVar(&restartBuild, "build", false, "rebuild image and recreate container before waiting readiness")
	cmd.AddCommand(restartCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "ps",
		Short: "show container status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.serverPS()
		},
	})

	var follow bool
	logsCmd := &cobra.Command{
		Use:   "logs [service]",
		Short: "show logs (all services or one service)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return a.serverLogs(args[0], follow)
			}
			return a.serverLogs("", follow)
		},
	}
	logsCmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow logs or not")
	cmd.AddCommand(logsCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "reload",
		Short: "reload the Minecraft server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.serverReload()
		},
	})

	return cmd
}
