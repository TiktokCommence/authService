package registry

import (
	"github.com/TiktokCommence/authService/internal/conf"
	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"time"
)

var ProviderSet = wire.NewSet(NewRegistrarServer)

func NewRegistrarServer(c *conf.Registry, logger log.Logger) *etcd.Registry {
	// ETCD源地址
	endpoints := []string{c.Addr}

	// ETCD配置信息
	etcdCfg := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: time.Second,
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
	}

	// 创建ETCD客户端
	client, err := clientv3.New(etcdCfg)
	if err != nil {
		panic(err)
	}
	logger.Log(log.LevelInfo, "msg", "service registry successfully")
	//fmt.Println("connect successfully")
	// 创建服务注册 registrar
	registrar := etcd.New(client)
	return registrar
}
