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

	cmd.AddCommand(&cobra.Command{
		Use:   "restart [service]",
		Short: "restart container service (default: world)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			service := "world"
			if len(args) == 1 {
				service = args[0]
			}
			return a.serverRestart(service)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "ps",
		Short: "show container status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.serverPS()
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "logs [service]",
		Short: "follow logs (all services or one service)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return a.serverLogs(args[0])
			}
			return a.serverLogs("")
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "reload",
		Short: "reload the Minecraft server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.serverReload()
		},
	})

	return cmd
}
