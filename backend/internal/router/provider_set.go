package router

import "github.com/google/wire"

var HttpServerSet = wire.NewSet(
	NewHttpServer,
)
