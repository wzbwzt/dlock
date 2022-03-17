# 分布式锁的实现
基于etcd/数据库/zookeeper/redis实现的分布式锁


>- 基于etcd
>- 基于数据库: 待更新...
>- 基于zooKeeper: 待更新...
>- 基于redis: 待更新...


### 使用姿势

```go
go get -u  github.com/wzbwzt/dlock

```

### 使用事例
```go
	//初始化
	dlock.Init(dlock.ETCD,[]string{"127.0.0.1:2379"})

	//实例化对象
	lock:=dlock.NewDlock(dlock.WithPath("/lock/a"),dlock.WithTiemOut(time.Secoud*5))
	err:=lock.Lock()
	if err!=nil{
		return 
	}
	defer lock.Unlock()

```


