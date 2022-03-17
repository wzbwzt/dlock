package dlock

import (
	"errors"
	"fmt"
	"time"

	"github.com/coreos/etcd/clientv3"
	etcdclient "github.com/wzbwzt/dlock/etcd"
)

type Dlock struct {
	//路径
	path string
	//超时控制
	timeout time.Duration
	//注册类型
	register RegisterObj
	//etcd客户端
	etcdclient *clientv3.Client
}

var defaultDlock = &Dlock{}

type DlockOption func(*Dlock)

func Init(obj RegisterObj, endpoint []string) {
	defaultDlock.register = obj

	switch obj {
	case ETCD:
		client := etcdclient.InitClient(endpoint)
		defaultDlock.etcdclient = client
	default:
		panic("无效的注册地址")
	}
}

func NewDlock(opts ...DlockOption) *Dlock {
	for _, opt := range opts {
		opt(defaultDlock)
	}
	return defaultDlock
}

func WithPath(path string) DlockOption {
	return func(d *Dlock) {
		d.path = path
	}
}

func WithTimeout(duration time.Duration) DlockOption {
	return func(d *Dlock) {
		d.timeout = duration
	}
}

//获取etcd注册客户端
func (d *Dlock) GetEtcdClient() *clientv3.Client {
	return d.etcdclient
}

func (d *Dlock) Lock() error {
	if d.timeout == 0 {
		d.timeout = time.Second * 10
	}
	switch d.register {
	case ETCD:
		err := etcdclient.Lock(d.path, d.timeout)
		if err != nil {
			return err
		}
	default:
		return errors.New(fmt.Sprintf("%v暂未开发", d.register))
	}
	return nil
}

func (d *Dlock) UnLock() error {
	switch d.register {
	case ETCD:
		err := etcdclient.Unlock(d.path)
		if err != nil {
			return err
		}
	default:
		return errors.New(fmt.Sprintf("%v暂未开发", d.register))
	}
	return nil
}
