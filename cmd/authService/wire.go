//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/TiktokCommence/authService/internal/biz"
	"github.com/TiktokCommence/authService/internal/casbin"
	"github.com/TiktokCommence/authService/internal/conf"
	"github.com/TiktokCommence/authService/internal/redis"
	"github.com/TiktokCommence/authService/internal/registry"
	"github.com/TiktokCommence/authService/internal/server"
	"github.com/TiktokCommence/authService/internal/service"
	"github.com/TiktokCommence/authService/internal/token"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *conf.Token, *conf.Registry, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		biz.ProviderSet,
		casbin.ProviderSet,
		redis.ProviderSet,
		server.ProviderSet,
		service.ProviderSet,
		token.ProviderSet,
		registry.ProviderSet,
		newApp,
		wire.Bind(new(service.AuthHandler), new(*biz.AuthUserCase)),
		wire.Bind(new(biz.RoleHandler), new(*casbin.AuthCase)),
		wire.Bind(new(biz.SignHandler), new(*biz.Signer)),
		wire.Bind(new(biz.VerifyHandler), new(*biz.Verifyer)),
		wire.Bind(new(biz.SetKeyer), new(*redis.Cache)),
		wire.Bind(new(biz.TokenGenerater), new(*token.JWTer)),
		wire.Bind(new(biz.CacheProxy), new(*redis.Cache)),
		wire.Bind(new(biz.TokenVerifyer), new(*token.JWTer)),
	))
}
