package main

import (
	"flag"
	"github.com/TiktokCommence/authService/internal/conf"
	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	_ "go.uber.org/automaxprocs"
	"google.golang.org/protobuf/types/known/durationpb"
	"os"
	"time"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string = "auth_service"
	// Version is the version of the compiled software.
	Version string = "v1"
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, gs *grpc.Server, r *etcd.Registry) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
		),
		kratos.Registrar(r),
	)
}

func main() {
	flag.Parse()
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	watchToken(c, &bc)

	app, cleanup, err := wireApp(bc.Server, bc.Data, bc.Token, bc.Registry, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}

func watchToken(c config.Config, bc *conf.Bootstrap) {
	if err := c.Watch("token.secret", func(key string, value config.Value) {
		log.Infof("config changed: %s = %v\n", key, value)
		secret, err := c.Value("token.secret").String()
		if err != nil {
			log.Error(err)
			return
		}
		bc.Token.Secret = secret
	}); err != nil {
		log.Error(err)
	}
	if err := c.Watch("token.expiration", func(key string, value config.Value) {
		log.Infof("config changed: %s = %v\n", key, value)
		expirStr, err := c.Value("token.expiration").String()
		if err != nil {
			log.Error(err)
			return
		}
		expir, err := time.ParseDuration(expirStr)
		if err != nil {
			log.Error(err)
			return
		}
		bc.Token.Expiration = durationpb.New(expir)
	}); err != nil {
		log.Error(err)
	}
}
