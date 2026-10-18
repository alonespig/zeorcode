//go:build wireinject
// +build wireinject

package di

import (
	"github.com/alonespig/zeorcode/backend/internal/http/handler"
	"github.com/alonespig/zeorcode/backend/internal/http/router"
	"github.com/alonespig/zeorcode/backend/internal/infra/cache"
	"github.com/alonespig/zeorcode/backend/internal/infra/database"
	"github.com/alonespig/zeorcode/backend/internal/infra/judgepool"
	"github.com/alonespig/zeorcode/backend/internal/infra/llm"
	"github.com/alonespig/zeorcode/backend/internal/infra/mail"
	"github.com/alonespig/zeorcode/backend/internal/infra/mq"
	"github.com/alonespig/zeorcode/backend/internal/infra/redis"
	"github.com/alonespig/zeorcode/backend/internal/infra/session"
	"github.com/alonespig/zeorcode/backend/internal/repository"
	"github.com/alonespig/zeorcode/backend/internal/service"
	"github.com/alonespig/zeorcode/backend/pkg/judge"

	"github.com/google/wire"
)

func InitHttpServer() (*router.HttpServer, func(), error) {
	wire.Build(
		database.NewMysqlClient,
		redis.NewRedisClient,
		session.New,
		service.NewTokenService,
		mq.New,
		judgepool.New,
		cache.CacheSet,
		mail.MailSet,
		llm.ProviderSet,
		repository.RepoSet,
		repository.NewProblemTestDataStore,
		service.ServerSet,
		wire.Bind(new(service.SubmissionStore), new(*repository.SubmissionRepo)),
		wire.Bind(new(service.SubmissionUserReader), new(*repository.UserRepo)),
		wire.Bind(new(service.SubmissionProblemReader), new(*repository.ProblemRepo)),
		wire.Bind(new(service.SubmissionQueue), new(*mq.MQ)),
		wire.Bind(new(service.SubmissionCache), new(*cache.Cache)),
		wire.Bind(new(service.ProblemTestDataReader), new(*repository.ProblemTestDataStore)),
		wire.Bind(new(service.LanguageResolver), new(*service.LanguageService)),
		wire.Bind(new(service.CodeRunPool), new(*judge.Pool)),
		handler.HandlerSet,
		router.NewHttpServer,
	)
	return nil, nil, nil
}
