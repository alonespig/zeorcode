package cmd

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"zoj/internal/common/logger"
	"zoj/internal/infra/cache"
	"zoj/internal/infra/conn"
	"zoj/internal/infra/mq"
	"zoj/internal/judgeworker"
	"zoj/internal/repository"
	"zoj/internal/submission/adapter/outboxstore"
	"zoj/internal/submission/application/dispatchrelay"
	"zoj/pkg/judge"

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

		db, closeDB := conn.NewMysqlClient()
		defer closeDB()
		rdb, closeRedis := conn.NewRedisClient()
		defer closeRedis()

		// 判题机地址：优先用 judge.urls（多台，负载均衡）；没配就回退单个 judge.url
		urls := viper.GetStringSlice("judge.urls")
		if len(urls) == 0 {
			if u := viper.GetString("judge.url"); u != "" {
				urls = []string{u}
			}
		}
		pool := judge.NewPool(urls)
		if pool.Size() == 0 {
			logger.Fatalw("未配置判题机地址（judge.urls 或 judge.url）")
		}
		pool.StartHealthCheck(15 * time.Second) // 后台定时探活，Next() 自动跳过挂掉的实例
		defer pool.Stop()
		logger.Infow("judge pool ready", "instances", pool.Size(), "urls", pool.URLs())

		submissionRepo := repository.NewSubmissionRepo(db)
		repos := judgeworker.Repositories{
			Submissions:    submissionRepo,
			Problems:       repository.NewProblemRepo(db),
			Contests:       repository.NewContestRepo(db),
			RemoteAccounts: repository.NewRemoteAccountRepo(db),
		}
		queue := mq.New(rdb)
		w := judgeworker.NewWorker(db, pool, repos, judgeConcurrency, queue, cache.NewCache(rdb))
		relay, err := dispatchrelay.New(
			outboxstore.New(db),
			queue,
			dispatchRelayOwner(),
			dispatchrelay.DefaultConfig(),
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

func dispatchRelayOwner() string {
	host, err := os.Hostname()
	if err != nil || host == "" {
		host = "worker"
	}
	hostHash := sha256.Sum256([]byte(host))
	return fmt.Sprintf("judge-relay-%x-%d", hostHash[:8], os.Getpid())
}

func init() {
	judgeCmd.Flags().IntVarP(&judgeConcurrency, "concurrency", "c", judgeworker.DefaultConcurrency, "评测并发 worker 数量")
}
