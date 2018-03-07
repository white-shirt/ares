package chiwoo

import (
	"github.com/coreos/etcd/clientv3"
	etcdnaming "github.com/coreos/etcd/clientv3/naming"
	"google.golang.org/grpc"
)

type ETCDBalancer struct {
	endpoints []string
}

func NewETCDBalancer(endpoints []string) grpc.Balancer {
	cc, err := clientv3.New(clientv3.Config{
		Endpoints:        endpoints,
		AutoSyncInterval: 0,
	})

	if err != nil {
		panic("client v3 connect etcd error")
	}

	rr := &etcdnaming.GRPCResolver{Client: cc}
	return grpc.RoundRobin(rr)
}
