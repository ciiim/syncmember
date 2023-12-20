package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/ciiim/syncmember"
	"github.com/ciiim/syncmember/signal"
)

func main() {
	si := signal.NewManager()
	s1 := syncmember.NewSyncMember("node1", syncmember.DebugConfig().SetPort(9000).SetLogLevel(slog.LevelInfo).OpenLogDetail(false))
	s2 := syncmember.NewSyncMember("node2", syncmember.DebugConfig().SetPort(9001).SetLogLevel(slog.LevelInfo).OpenLogDetail(false))
	s3 := syncmember.NewSyncMember("node3", syncmember.DebugConfig().SetPort(9002).SetLogLevel(slog.LevelInfo).OpenLogDetail(false))
	si.AddWatcher(os.Interrupt, "shutdown", func() {
		s2.Shutdown()
		s1.Shutdown()
		s3.Shutdown()
	})
	go s1.Run()
	go s2.Run()
	go s3.Run()
	time.Sleep(time.Second * 1)
	s1.Join("172.26.123.188:9001")
	time.Sleep(time.Second * 1)
	s2.Join("172.26.123.188:9002")
	time.Sleep(time.Second * 10)
	s3.Shutdown()
	si.Wait()
}
