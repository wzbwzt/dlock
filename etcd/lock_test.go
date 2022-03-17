package etcd

import (
	"context"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	log "github.com/micro/go-micro/v2/logger"
)

func TestMain(m *testing.M) {
	var err error
	etcdcli, err = clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		panic(err)
	}
	lockSession, err = concurrency.NewSession(etcdcli, concurrency.WithTTL(sessionTimeout))
	if err != nil {
		panic(err)
	}

	m.Run()
}

func TestDlock(t *testing.T) {
	lock1 := "/lock/a"
	lock2 := "/lock"
	lock3 := "/lock/b"

	register := Register{}
	t.Logf("locking %s", lock1)
	err := register.Lock(lock1, time.Second)
	if err != nil {
		t.Errorf("lock %s failed: %s", lock1, err)
	}

	t.Logf("locking %s again", lock1)
	err = register.Lock(lock1, time.Second*5)
	if err != context.DeadlineExceeded {
		t.Errorf("lock should timeout! %s", err)
	}

	t.Logf("locking %s", lock2)
	err = register.Lock(lock2, time.Second*5)
	if err != context.DeadlineExceeded {
		t.Errorf("lock1 owns lock2, lock should timeout! %s", err)
	}

	done := make(chan struct{})
	go func() {
		err = register.Lock(lock2, time.Hour)
		if err != nil {
			t.Errorf("lock %s failed: %s", lock2, err)
		}
		t.Logf("lock %s is realeased!", lock2)
		close(done)
	}()

	t.Logf("unlocking %s", lock1)
	err = register.Unlock(lock1)
	if err != nil {
		t.Errorf("unlock %s failed: %s", lock1, err)
	}
	<-done

	t.Logf("locking %s", lock3)
	err = register.Lock(lock3, time.Second)
	if err != nil {
		t.Errorf("lock %s failed: %s", lock3, err)
	}

	t.Logf("unlock %s", lock3)
	err = register.Unlock(lock3)
	if err != nil {
		t.Errorf("unlock %s failed: %s", lock3, err)
	}

	// 监听session有效性
	startReconnect := make(chan struct{})
	go func() {
		<-startReconnect
		for {
			<-lockSession.Done()
			retry := 0
			for {
				retry++
				log.Warnf("lock session gone, trying to recreate, take %d", retry)
				lockSession, err = concurrency.NewSession(etcdcli, concurrency.WithTTL(sessionTimeout))
				if err == nil {
					rectifyLocks() // 修复旧锁
					startReconnect <- struct{}{}
					break
				}
				log.Errorf("create lock session failed: %s", err)
				time.Sleep(time.Second)
			}
		}
	}()

	t.Logf("closing session")
	lockSession.Close()
	err = register.Lock(lock1, time.Hour)
	if err == nil {
		t.Errorf("expeting lock error")
	} else {
		t.Logf("got error as expected: %s", err)
	}

	startReconnect <- struct{}{}
	<-startReconnect

	err = register.Lock(lock1, time.Hour)
	if err != nil {
		t.Errorf("lock error: %s", err)
	}
	err = register.Unlock(lock1)
	if err != nil {
		t.Errorf("unlock error: %s", err)
	}
}
