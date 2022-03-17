# 分布式锁的实现

基于 etcd/数据库/zookeeper/redis 实现的分布式锁

> - 基于 etcd
> - 基于数据库: 待更新...
> - 基于 zooKeeper: 待更新...
> - 基于 redis: 待更新...

### 使用姿势

```go
go get -u  github.com/wzbwzt/dlock

```

### 使用事例

```go
etcd := etcd.NewRegister(etcd.WithTimeOut(time.Second * 5))
lock := dlock.NewDlock(dlock.WithRegister(etcd))
//获取etcd的客户端
// etcdclient:=etcd.GetEtcdClient()
path := "/lock/a"
lock.Lock(path, time.Second*10)
defer lock.UnLock(path)

```
