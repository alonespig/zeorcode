package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/alonespig/zeorcode/backend/internal/http/router"
	"github.com/alonespig/zeorcode/backend/internal/infra/logger"
	"github.com/alonespig/zeorcode/backend/internal/worker/reminder"

	"github.com/gin-gonic/gin"
)

// HTTP owns the HTTP server and the tasks that run in the same process.
type HTTP struct {
	handler     http.Handler
	runReminder func(context.Context)
}

func NewHTTP(routes *router.HttpServer, deadlineWorker *reminder.HomeworkDeadlineWorker) *HTTP {
	engine := gin.Default()
	routes.Register(engine)
	return &HTTP{handler: engine, runReminder: deadlineWorker.Run}
}

func (app *HTTP) Run(ctx context.Context, addr string) error {
	var config net.ListenConfig
	listener, err := config.Listen(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}
	return app.serve(ctx, listener)
}

func (app *HTTP) serve(ctx context.Context, listener net.Listener) error {
	server := &http.Server{
		Handler:      app.handler,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 120 * time.Second,
	}
	workerCtx, stopWorker := context.WithCancel(ctx)
	workerDone := make(chan struct{})
	go func() {
		defer close(workerDone)
		app.runReminder(workerCtx)
	}()
	defer func() {
		stopWorker()
		<-workerDone
	}()
	defer server.Close()

	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(listener) }()
	logger.Infow("http server started", "addr", listener.Addr().String())

	select {
	case err := <-serveErr:
		return httpServeError(err)
	case <-ctx.Done():
	}

	logger.Infow("http server shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}
	if err := httpServeError(<-serveErr); err != nil {
		return err
	}
	logger.Infow("http server stopped cleanly")
	return nil
}

func httpServeError(err error) error {
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return fmt.Errorf("serve HTTP: %w", err)
}
