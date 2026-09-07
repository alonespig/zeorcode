//go:build wireinject
// +build wireinject

package bootstrap

import (
	"zoj/internal/handler"
	"zoj/internal/infra/cache"
	"zoj/internal/infra/conn"
	"zoj/internal/infra/llm"
	"zoj/internal/infra/mail"
	"zoj/internal/infra/mq"
	"zoj/internal/infra/session"
	"zoj/internal/middleware"
	"zoj/internal/problem/adapter/testdatastore"
	"zoj/internal/repository"
	"zoj/internal/router"
	"zoj/internal/service"

	"github.com/google/wire"
)

func InitHttpServer() (*router.HttpServer, func(), error) {
	wire.Build(
		conn.NewMysqlClient,
		conn.NewRedisClient,
		session.New,
		middleware.NewAuth,
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
