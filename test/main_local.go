package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"sync"

	"github.com/ciiim/syncmember"
	"github.com/ciiim/syncmember/signal"

	_ "net/http/pprof"
)

const (
	LocalAddr   = "127.0.0.1"
	ClusterSize = 2
)

func main() {
	runtime.SetCPUProfileRate(1000)
	runtime.SetMutexProfileFraction(1)
	go func() {

		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	var startWg, stopWg sync.WaitGroup
	stopCh := make(chan struct{})

	si := signal.NewManager()

	nodes := make([]*syncmember.SyncMember, ClusterSize)

	startWg.Add(ClusterSize)
	stopWg.Add(ClusterSize)
	for i := 0; i < ClusterSize; i++ {
		nodes[i] = syncmember.NewSyncMember(
			fmt.Sprintf("node%d", i+1),
			syncmember.DebugConfig().SetPort(9000+i+1).SetLogLevel(slog.LevelInfo).OpenLogDetail(false),
		)

		go func(node *syncmember.SyncMember) {
			defer stopWg.Done()

			// 运行节点
			go node.Run()

			startWg.Done()
			// 等待停止信号
			<-stopCh
			// 关闭节点
			node.Shutdown()
		}(nodes[i])
	}

	// 添加信号处理器以优雅地关闭所有节点
	si.AddWatcher(os.Interrupt, "shutdown", func() {
		close(stopCh)
	})

	si.Wait()

	// 等待所有节点启动
	startWg.Wait()

	// 让每个节点加入集群
	for i := 1; i < ClusterSize; i++ {
		_ = nodes[i].Join(fmt.Sprintf("%s:%d", LocalAddr, 9001))
	}

	stopWg.Wait()
}
