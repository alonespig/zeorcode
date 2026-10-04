//go:build wireinject
// +build wireinject

package di

import (
	"zoj/internal/http/handler"
	"zoj/internal/http/router"
	"zoj/internal/infra/cache"
	"zoj/internal/infra/database"
	"zoj/internal/infra/llm"
	"zoj/internal/infra/mail"
	"zoj/internal/infra/mq"
	"zoj/internal/infra/redis"
	"zoj/internal/infra/session"
	"zoj/internal/problem/adapter/testdatastore"
	"zoj/internal/repository"
	"zoj/internal/service"

	"github.com/google/wire"
)

func InitHttpServer() (*router.HttpServer, func(), error) {
	wire.Build(
		database.NewMysqlClient,
		redis.NewRedisClient,
		session.New,
		service.NewTokenService,
		mq.New,
		cache.CacheSet,
		mail.MailSet,
		llm.ProviderSet,
		repository.RepoSet,
		testdatastore.New,
		service.ServerSet,
		wire.Bind(new(service.SubmissionStore), new(*repository.SubmissionRepo)),
		wire.Bind(new(service.SubmissionUserReader), new(*repository.UserRepo)),
		wire.Bind(new(service.SubmissionProblemReader), new(*repository.ProblemRepo)),
		wire.Bind(new(service.SubmissionQueue), new(*mq.MQ)),
		wire.Bind(new(service.SubmissionCache), new(*cache.Cache)),
		wire.Bind(new(service.ProblemTestDataReader), new(*testdatastore.Store)),
		wire.Bind(new(service.LanguageResolver), new(*service.LanguageService)),
		handler.HandlerSet,
		router.NewHttpServer,
	)
	return nil, nil, nil
}
