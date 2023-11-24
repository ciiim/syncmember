package main

import (
	"log/slog"
	"os"

	"github.com/ciiim/syncmember"
	"github.com/ciiim/syncmember/signal"
)

func main() {
	si := signal.NewManager()
	s1 := syncmember.NewSyncMember("node1", syncmember.DebugConfig.SetPort(9000).SetLogLevel(slog.LevelInfo).OpenLogDetail(false))
	s2 := syncmember.NewSyncMember("node2", syncmember.DebugConfig.SetPort(9001).SetLogLevel(slog.LevelInfo).OpenLogDetail(false))
	si.AddWatcher(os.Interrupt, "shutdown", func() {
		s2.Shutdown()
		s1.Shutdown()
	})
	go si.Wait()
	go s1.Run()
	s2.JoinDebug("172.26.123.188:9000")
	s2.Run()
}
