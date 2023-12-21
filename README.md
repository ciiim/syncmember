# SyncMember

基于Gossip协议的成员同步组件

支持同步Key-Value数据

Member synchronization component based on Gossip protocol

Support synchronization of Key-Value data

### 安装 Install
```shell
go get github.com/ciiim/syncmember
```

### 例子 Example

```go
func main() {
    // create two nodes
    // once created, nodes can be joined
    s1 := syncmember.NewSyncMember("node1",syncmember.DefaultConfig().SetPort(9001))
    s2 := syncmember.NewSyncMember("node2",syncmember.DefaultConfig().SetPort(9002))

    // add signal watcher
    si := signal.NewManager()
    si.AddWatcher(os.Interrupt, "shutdown", func() {
		s1.Shutdown()
		s2.Shutdown()
	})
    si.Wait()

    go s1.Run()

    //join node1
    s1.Join("127.0.0.1:9002")
    s2.Run()
}
```
#### Key-Value 操作 Key-Value operation
```go
func main() {
    s1 := ...
    s2 := ...

    s1.SetKV("key1", []byte("value1"))
    // ...
    s2.GetValue("key1") // return []byte("value1")
    // ...
    s2.DeleteKV("key1")
    // ...
    s1.GetValue("key1") // return nil
}
```