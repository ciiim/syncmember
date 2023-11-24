package main

import (
	"log/slog"
	"os"

	"github.com/ciiim/syncmember"
	"github.com/ciiim/syncmember/signal"
)

func main1() {
	si := signal.NewManager()
	s1 := syncmember.NewSyncMember("node1", syncmember.DebugConfig.SetLogLevel(slog.LevelDebug))
	si.AddWatcher(os.Interrupt, "shutdown", func() {
		s1.Shutdown()
	})
	go si.Wait()
	s1.Run()

}
