package lib

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

type EtcdHelper struct {
	etcdClient *clientv3.Client
	leaseID    clientv3.LeaseID
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

func (sh *EtcdHelper) Register(serviceName, serviceHost string, servicePort uint, refreshSecond int64) error {
	lease, err := sh.etcdClient.Grant(context.Background(), refreshSecond*3)
	if err != nil {
		return err
	}
	sh.leaseID = lease.ID
	key := fmt.Sprintf("%s:%s:%d", serviceName, serviceHost, servicePort)
	value := fmt.Sprintf("%s:%d", serviceHost, servicePort)
	//设置key,value
	if _, err = sh.etcdClient.Put(context.Background(), key, value, clientv3.WithLease(sh.leaseID)); err != nil {
		return err
	}
	keepAliveChan, err := sh.etcdClient.KeepAlive(context.Background(), sh.leaseID)
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

func (sh *EtcdHelper) UnRegister() error {
	defer sh.etcdClient.Close()
	if _, err := sh.etcdClient.Revoke(context.Background(), sh.leaseID); err != nil {
		return err
	}
	fmt.Println("注销服务")
	return nil
}
