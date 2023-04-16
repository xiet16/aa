package lib

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/clientv3"
	"sync"
	"time"
)

type EtcdHelper struct {
	etcdClient *clientv3.Client
	leaseID    clientv3.LeaseID
	serverList sync.Map
}

func NewEtcdHelper(endpoints []string) (*EtcdHelper, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	helper := &EtcdHelper{
		etcdClient: client,
	}
	return helper, nil
}

func (h *EtcdHelper) Register(serviceName, serviceHost string, servicePort uint, refreshSecond int64) error {
	lease, err := h.etcdClient.Grant(context.Background(), refreshSecond*3)
	if err != nil {
		return err
	}
	h.leaseID = lease.ID
	key := fmt.Sprintf("%s:%s:%d", serviceName, serviceHost, servicePort)
	value := fmt.Sprintf("%s:%d", serviceHost, servicePort)
	//设置key,value
	if _, err = h.etcdClient.Put(context.Background(), key, value, clientv3.WithLease(h.leaseID)); err != nil {
		return err
	}
	keepAliveChan, err := h.etcdClient.KeepAlive(context.Background(), h.leaseID)
	if err != nil {
		return err
	}
	fmt.Printf("注册服务：%s,%s,%d\n", serviceName, serviceHost, servicePort)
	go func() {
		for aliveCh := range keepAliveChan {
			fmt.Printf("续约：%d,%s\n", aliveCh.TTL, time.Now().String())
		}
	}()
	return nil
}

func (h *EtcdHelper) UnRegister() error {
	defer h.etcdClient.Close()
	if _, err := h.etcdClient.Revoke(context.Background(), h.leaseID); err != nil {
		return err
	}
	fmt.Println("注销服务")
	return nil
}

func (h *EtcdHelper) Discover(prefix string) {
	rsp, err := h.etcdClient.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		fmt.Errorf("etcd discover err :%v", err)
		return
	}
	for _, kv := range rsp.Kvs {
		h.serverList.Store(string(kv.Key), string(kv.Value))
	}
	go h.Watch(prefix)
}

func (h *EtcdHelper) Watch(prefix string) {
	wh := h.etcdClient.Watch(context.Background(), prefix, clientv3.WithPrefix())
	for w := range wh {
		for _, ev := range w.Events {
			switch ev.Type {
			case mvccpb.PUT:
				h.serverList.Store(string(ev.Kv.Key), string(ev.Kv.Value))
				break
			case mvccpb.DELETE:
				h.serverList.Delete(string(ev.Kv.Key))
				break
			default:
			}
		}
	}
}

func (h *EtcdHelper) Close() {
	h.etcdClient.Close()
}
