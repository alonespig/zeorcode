package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	judgev1 "github.com/alonespig/zeorcode/judge-node/api/judge/v1"
	"github.com/alonespig/zeorcode/judge-node/internal/judgenode"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type config struct {
	listenAddr      string
	nodeID          string
	goJudgeURL      string
	artifactBaseURL string
	artifactToken   string
	rpcToken        string
	cacheDir        string
	concurrency     int
	tlsCertFile     string
	tlsKeyFile      string
	insecure        bool
}

func main() {
	if err := run(); err != nil {
		slog.Error("judge node stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := parseConfig()
	if err := validateConfig(cfg); err != nil {
		return err
	}
	cache, err := judgenode.NewArtifactCache(cfg.cacheDir, cfg.artifactBaseURL, cfg.artifactToken)
	if err != nil {
		return err
	}
	service, err := judgenode.NewServer(cfg.nodeID, cfg.goJudgeURL, cache, cfg.concurrency)
	if err != nil {
		return err
	}

	options := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(16 << 20),
		grpc.UnaryInterceptor(judgenode.UnaryAuthInterceptor(cfg.rpcToken)),
		grpc.StreamInterceptor(judgenode.StreamAuthInterceptor(cfg.rpcToken)),
	}
	if !cfg.insecure {
		creds, err := credentials.NewServerTLSFromFile(cfg.tlsCertFile, cfg.tlsKeyFile)
		if err != nil {
			return fmt.Errorf("load judge node TLS certificate: %w", err)
		}
		options = append(options, grpc.Creds(creds))
	}
	server := grpc.NewServer(options...)
	judgev1.RegisterJudgeNodeServiceServer(server, service)
	listener, err := net.Listen("tcp", cfg.listenAddr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", cfg.listenAddr, err)
	}
	defer listener.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	serveErr := make(chan error, 1)
	go func() {
		slog.Info("judge node started", "node", cfg.nodeID, "listen", cfg.listenAddr, "concurrency", cfg.concurrency)
		serveErr <- server.Serve(listener)
	}()
	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
	}

	stopped := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(stopped)
	}()
	select {
	case <-stopped:
	case <-time.After(30 * time.Second):
		server.Stop()
	}
	return nil
}

func parseConfig() config {
	hostname, _ := os.Hostname()
	cfg := config{}
	flag.StringVar(&cfg.listenAddr, "listen", envOr("ZEORCODE_JUDGE_NODE_LISTEN", "127.0.0.1:60051"), "gRPC listen address")
	flag.StringVar(&cfg.nodeID, "node-id", envOr("ZEORCODE_JUDGE_NODE_ID", hostname), "stable node identifier")
	flag.StringVar(&cfg.goJudgeURL, "go-judge-url", envOr("ZEORCODE_JUDGE_NODE_GO_JUDGE_URL", "http://127.0.0.1:5050"), "local go-judge HTTP URL")
	flag.StringVar(&cfg.artifactBaseURL, "artifact-base-url", os.Getenv("ZEORCODE_JUDGE_ARTIFACT_BASE_URL"), "central backend URL used to download problem artifacts")
	flag.StringVar(&cfg.artifactToken, "artifact-token", os.Getenv("ZEORCODE_JUDGE_ARTIFACT_TOKEN"), "problem artifact bearer token")
	flag.StringVar(&cfg.rpcToken, "rpc-token", os.Getenv("ZEORCODE_JUDGE_NODE_TOKEN"), "gRPC bearer token")
	flag.StringVar(&cfg.cacheDir, "cache-dir", envOr("ZEORCODE_JUDGE_NODE_CACHE_DIR", "/var/lib/zeorcode-judge/problems"), "problem artifact cache directory")
	flag.IntVar(&cfg.concurrency, "concurrency", envInt("ZEORCODE_JUDGE_NODE_CONCURRENCY", 1), "maximum concurrent submissions")
	flag.StringVar(&cfg.tlsCertFile, "tls-cert", os.Getenv("ZEORCODE_JUDGE_NODE_TLS_CERT"), "server TLS certificate")
	flag.StringVar(&cfg.tlsKeyFile, "tls-key", os.Getenv("ZEORCODE_JUDGE_NODE_TLS_KEY"), "server TLS private key")
	flag.BoolVar(&cfg.insecure, "insecure", envBool("ZEORCODE_JUDGE_NODE_INSECURE"), "explicitly allow plaintext gRPC")
	flag.Parse()
	return cfg
}

func validateConfig(cfg config) error {
	if strings.TrimSpace(cfg.rpcToken) == "" || strings.TrimSpace(cfg.artifactToken) == "" {
		return errors.New("judge node and artifact tokens are required")
	}
	if cfg.concurrency <= 0 || cfg.concurrency > 64 {
		return errors.New("judge node concurrency must be between 1 and 64")
	}
	if cfg.insecure {
		if cfg.tlsCertFile != "" || cfg.tlsKeyFile != "" {
			return errors.New("do not configure TLS files together with --insecure")
		}
		return nil
	}
	if cfg.tlsCertFile == "" || cfg.tlsKeyFile == "" {
		return errors.New("TLS certificate and key are required unless --insecure is explicitly set")
	}
	return nil
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	var value int
	if _, err := fmt.Sscan(os.Getenv(key), &value); err == nil && value > 0 {
		return value
	}
	return fallback
}

func envBool(key string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
