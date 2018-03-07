package registry

import (
	"github.com/coreos/etcd/clientv3"
)

type etcdWatcher struct {
	clientv3.WatchChan
	client *clientv3.Client
}
