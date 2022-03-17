package etcdclient

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
)

type lock struct {
	llock int32              // 本地锁，用于本地保护（1表示上锁，0表示释放）
	rlock *concurrency.Mutex // 远程锁，用于分布式保护
}

var (
	lockMap   sync.Map
	lockMutex sync.RWMutex
)

// 分布互斥锁添加及锁定，lockPath代表锁路径
// 为避免锁错误，必须设置超时等待时间
func Lock(lockPath string, duration time.Duration) error {
	for strings.HasSuffix(lockPath, "/") {
		lockPath = lockPath[:len(lockPath)-1]
	}
	if len(lockPath) == 0 || !strings.HasPrefix(lockPath, "/") {
		panic("invalid lock path")
	}

	// 获取/初始化锁
	lockMutex.RLock()
	defer lockMutex.RUnlock()
	lk := &lock{llock: 0, rlock: concurrency.NewMutex(lockSession, lockPath)}
	v, _ := lockMap.LoadOrStore(lockPath, lk)
	realLock := v.(*lock)

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	// 先加本地锁，模拟自旋锁
	tc := time.NewTicker(time.Millisecond * 10)
	defer tc.Stop()
	for !atomic.CompareAndSwapInt32(&realLock.llock, 0, 1) {
		select {
		case <-tc.C:
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// 再加远程锁
	err := realLock.rlock.Lock(ctx)
	if err != nil {
		atomic.StoreInt32(&realLock.llock, 0)
		return err
	}

	return nil
}

// 分布互斥锁添加及解锁
func Unlock(lockPath string) error {
	for strings.HasSuffix(lockPath, "/") {
		lockPath = lockPath[:len(lockPath)-1]
	}
	if len(lockPath) == 0 || !strings.HasPrefix(lockPath, "/") {
		panic("invalid lock path")
	}

	// 获取
	lockMutex.RLock()
	defer lockMutex.RUnlock()
	v, ok := lockMap.Load(lockPath)
	if !ok {
		return fmt.Errorf("锁路径(%s)不存在", lockPath)
	}
	realLock := v.(*lock)

	// 释放远程锁
	err := realLock.rlock.Unlock(context.Background())
	if err != nil {
		return err
	}

	// 释放本地锁
	atomic.StoreInt32(&realLock.llock, 0)
	return nil
}

func rectifyLocks() {
	lockMutex.Lock()
	defer lockMutex.Unlock()

	lockMap.Range(func(k, v interface{}) bool {
		lockPath := k.(string)
		realLock := v.(*lock)
		realLock.rlock = concurrency.NewMutex(lockSession, lockPath)
		lockMap.Store(lockPath, realLock)
		return true
	})
}
