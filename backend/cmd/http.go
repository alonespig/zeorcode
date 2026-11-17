package cmd

import (
	"fmt"
	"os/signal"
	"syscall"

	"github.com/alonespig/zeorcode/backend/internal/di"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "启动http服务",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()
		app, cleanup, err := di.InitHTTPApp()
		if err != nil {
			return fmt.Errorf("init HTTP app: %w", err)
		}
		defer cleanup()
		return app.Run(ctx, viper.GetString("server.port"))
	},
}
