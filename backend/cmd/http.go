package cmd

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"zoj/internal/bootstrap"
	"zoj/internal/common/logger"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "启动http服务",
	Run: func(cmd *cobra.Command, args []string) {
		r := gin.Default()
		httpSrv, cleanup, err := bootstrap.InitHttpServer()
		if err != nil {
			logger.Fatalw("init http server failed", "err", err)
		}
		defer cleanup()
		httpSrv.Register(r)

		srv := &http.Server{
			Addr:         viper.GetString("server.port"),
			Handler:      r,
			ReadTimeout:  60 * time.Second,
			WriteTimeout: 120 * time.Second,
		}

		// 后台跑 HTTP server
		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatalw("http server crashed", "err", err)
			}
		}()
		logger.Infow("http server started", "addr", srv.Addr)

		// 等 Ctrl+C / SIGTERM
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()
		<-ctx.Done()
		logger.Infow("http server shutting down")

		// 给 10s 让在飞的请求跑完
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Errorw("http server shutdown failed", "err", err)
			return
		}
		logger.Infow("http server stopped cleanly")
	},
}
