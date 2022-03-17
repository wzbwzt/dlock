package etcd

import (
	"context"
	"fmt"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/micro/go-micro/util/log"
)

type Register struct {
	endpoint []string
	path     string
	timeout  time.Duration
}

type Option func(*Register)

var (
	etcdcli        *clientv3.Client
	lockSession    *concurrency.Session
	sessionTimeout = 15
)

func NewRegister(opts ...Option) *Register {
	register := &Register{}
	for _, opt := range opts {
		opt(register)
	}
	if register.timeout == 0 {
		register.timeout = 5
	}
	register.Init()

	return register
}

func WithEndpoint(endpoint []string) Option {
	return func(r *Register) {
		r.endpoint = endpoint
	}
}

func WithTimeOut(duration time.Duration) Option {
	return func(r *Register) {
		r.timeout = duration
	}
}

func (r *Register) Init() {
	success := make(chan struct{})
	go func() {
		select {
		case <-success:
			return
		case <-time.After(time.Second * 5):
			panic(fmt.Errorf("与etcd服务[%v]建立会话失败", r.endpoint))
		}
	}()

	var err error
	etcdcli, err = clientv3.New(clientv3.Config{
		Endpoints:   r.endpoint,
		DialTimeout: r.timeout,
	})
	if err != nil {
		panic(err)
	}
	lockSession, err = concurrency.NewSession(etcdcli, concurrency.WithTTL(sessionTimeout))
	if err != nil {
		panic(err)
	}
	close(success)

	// 监听session有效性
	go func() {
		for {
			<-lockSession.Done()
			retry := 0
			for {
				retry++
				log.Warnf("lock session gone, trying to recreate, take %d", retry)
				ctx, cancel := context.WithCancel(etcdcli.Ctx())
				success := make(chan struct{})
				go func() {
					select {
					case <-success:
						return
					case <-time.After(time.Second * 5):
						cancel()
					}
				}()
				lockSession, err = concurrency.NewSession(etcdcli,
					concurrency.WithTTL(sessionTimeout), concurrency.WithContext(ctx))
				if err == nil {
					log.Warnf("lock session recovered, rectify old locks")
					close(success)
					rectifyLocks() // 修复旧锁
					break
				}
				log.Errorf("create lock session failed: %s", err)
				time.Sleep(time.Second)
			}
		}
	}()
}

//获取etcd注册客户端
func (r *Register) GetEtcdClient() *clientv3.Client {
	return etcdcli
}
