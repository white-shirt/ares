package registry

import (
	"fmt"
	"log"
	"sync"

	"github.com/coreos/etcd/clientv3"
	etcdnaming "github.com/coreos/etcd/clientv3/naming"
	"google.golang.org/grpc/naming"
)

type etcdRegistry struct {
	Address  string
	client   *clientv3.Client
	resolver *etcdnaming.GRPCResolver
	options  Options

	data map[string]string
	mu   sync.Mutex
}

// NewETCDRegistry constructs an ETCD Registry.
func NewETCDRegistry(opts ...Option) Registry {
	var options Options
	for _, o := range opts {
		o(&options)
	}

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   options.Addrs,
		DialTimeout: options.Timeout,
		TLS:         options.TLSConfig,
	})

	if err != nil {
		// 初始化registry是在app startup阶段，该阶段的非容忍错误直接panic
		log.Fatalf("create etcd client failed: %s\n", err)
	}

	return &etcdRegistry{
		options:  options,
		client:   client,
		resolver: &etcdnaming.GRPCResolver{Client: client},
		data:     make(map[string]string),
	}
}

func (r *etcdRegistry) mustRegister(key, val string) {
	if err := r.Register(key, val); err != nil {
		panic(err)
	}
}

func (r *etcdRegistry) RegisterApp(info AppInfo) error {
	// Register registers service from provided service name and address.
	if r.client == nil || r.resolver == nil {
		return ErrUninitialRegistry
	}
	defer func() {
		if err := recover(); err != nil {
			r.UnregisterApp()
		}
	}()
	// 注册服务信息
	for srvLabel, srvAddr := range info.Services {
		key := fmt.Sprintf("%s:%s", srvLabel, info.AppName)
		r.mustRegister(key, srvAddr)
	}
	// 注册节点信息
	r.mustRegister(fmt.Sprintf("node:%s", info.AppName), info.UUID)
	r.mustRegister(fmt.Sprintf("info:%s:%s", info.AppName, info.UUID), info.Desc())
	return nil
}

func (r *etcdRegistry) UnregisterApp() error {
	r.UnregisterAll()
	return nil
}

func (r *etcdRegistry) Register(key string, val string) error {
	// Register registers service from provided service name and address.
	if r.client == nil || r.resolver == nil {
		return ErrUninitialRegistry
	}

	err := r.resolver.Update(r.client.Ctx(), key, naming.Update{
		Op:   naming.Add,
		Addr: val,
	})

	if err != nil {
		return err
	}

	r.mu.Lock()
	r.data[key] = val
	r.mu.Unlock()
	return nil
}

func (r *etcdRegistry) Unregister(key string, val string) error {
	if r.client == nil || r.resolver == nil {
		return ErrUninitialRegistry
	}

	err := r.resolver.Update(r.client.Ctx(), key, naming.Update{
		Op:   naming.Delete,
		Addr: val,
	})
	if err != nil {
		return err
	}

	r.mu.Lock()
	delete(r.data, key)
	r.mu.Unlock()
	return nil
}

func (r *etcdRegistry) UnregisterAll() {
	for k, v := range r.data {
		r.Unregister(k, v)
	}
}

func (r *etcdRegistry) Watch() (Watcher, error) {
	return nil, nil
}

func (r *etcdRegistry) String() string {
	return "etcd"
}
