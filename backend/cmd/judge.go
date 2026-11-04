package cmd

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/alonespig/zeorcode/backend/internal/infra/cache"
	"github.com/alonespig/zeorcode/backend/internal/infra/database"
	"github.com/alonespig/zeorcode/backend/internal/infra/judgeclient"
	"github.com/alonespig/zeorcode/backend/internal/infra/logger"
	"github.com/alonespig/zeorcode/backend/internal/infra/mq"
	"github.com/alonespig/zeorcode/backend/internal/infra/redis"
	"github.com/alonespig/zeorcode/backend/internal/judgeartifact"
	"github.com/alonespig/zeorcode/backend/internal/repository"
	"github.com/alonespig/zeorcode/backend/internal/repository/outbox"
	"github.com/alonespig/zeorcode/backend/internal/worker/dispatch"
	"github.com/alonespig/zeorcode/backend/internal/worker/judge"
	judgeapi "github.com/alonespig/zeorcode/backend/pkg/judge"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var judgeConcurrency int

var judgeCmd = &cobra.Command{
	Use:   "judge",
	Short: "启动判题 worker",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		db, closeDB := database.NewMysqlClient()
		defer closeDB()
		rdb, closeRedis := redis.NewRedisClient()
		defer closeRedis()

		// 判题机地址：优先用 judge.urls（多台，负载均衡）；没配就回退单个 judge.url
		mode := viper.GetString("judge.mode")
		if mode == "" {
			mode = "direct"
		}
		var pool *judgeapi.Pool
		var agentPool *judgeclient.Pool
		switch mode {
		case "direct":
			urls := viper.GetStringSlice("judge.urls")
			if len(urls) == 0 {
				if u := viper.GetString("judge.url"); u != "" {
					urls = []string{u}
				}
			}
			pool = judgeapi.NewPool(urls)
			if pool.Size() == 0 {
				logger.Fatalw("未配置判题机地址（judge.urls 或 judge.url）")
			}
			pool.StartHealthCheck(15 * time.Second)
			defer pool.Stop()
			logger.Infow("direct judge pool ready", "instances", pool.Size(), "urls", pool.URLs())
		case "agent":
			var err error
			agentPool, err = judgeclient.New(ctx, judgeclient.Config{
				Addresses:  configStringSlice("judge.agent_addresses"),
				Token:      viper.GetString("judge.agent_token"),
				Insecure:   viper.GetBool("judge.agent_insecure"),
				CAFile:     viper.GetString("judge.agent_ca_file"),
				ServerName: viper.GetString("judge.agent_server_name"),
			})
			if err != nil {
				logger.Fatalw("create distributed judge pool failed", "err", err)
			}
			defer agentPool.Close()
			logger.Infow("distributed judge pool ready", "instances", len(configStringSlice("judge.agent_addresses")))
		default:
			logger.Fatalw("unsupported judge mode", "mode", mode)
		}

		submissionRepo := repository.NewSubmissionRepo(db)
		repos := judge.Repositories{
			Submissions:    submissionRepo,
			Problems:       repository.NewProblemRepo(db),
			Contests:       repository.NewContestRepo(db),
			RemoteAccounts: repository.NewRemoteAccountRepo(db),
		}
		queue := mq.New(rdb)
		w := judge.NewWorker(db, pool, repos, judgeConcurrency, queue, cache.NewCache(rdb))
		if agentPool != nil {
			w.UseDistributedNodes(agentPool, judgeartifact.NewStore(viper.GetString("judge.data_dir")))
		}
		relay, err := dispatch.New(
			outbox.New(db),
			queue,
			dispatchRelayOwner(),
			dispatch.DefaultConfig(),
		)
		if err != nil {
			logger.Fatalw("create submission outbox relay failed", "err", err)
		}
		relayDone := make(chan struct{})
		go func() {
			defer close(relayDone)
			relay.Run(ctx, func(err error) {
				logger.Errorw("submission outbox relay failed", "err", err)
			})
		}()

		logger.Infow("judge worker starting", "concurrency", w.Concurrency())
		w.Run(ctx)
		<-relayDone
		logger.Infow("judge worker stopped")
	},
}

func configStringSlice(key string) []string {
	values := viper.GetStringSlice(key)
	if len(values) != 1 {
		return values
	}
	return strings.FieldsFunc(values[0], func(r rune) bool {
		return r == ',' || r == ';' || r == ' ' || r == '\t' || r == '\n'
	})
}

func dispatchRelayOwner() string {
	host, err := os.Hostname()
	if err != nil || host == "" {
		host = "worker"
	}
	hostHash := sha256.Sum256([]byte(host))
	return fmt.Sprintf("judge-relay-%x-%d", hostHash[:8], os.Getpid())
}

func init() {
	judgeCmd.Flags().IntVarP(&judgeConcurrency, "concurrency", "c", judge.DefaultConcurrency, "评测并发 worker 数量")
}
