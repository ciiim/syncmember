package main

import (
	"os"

	"github.com/ciiim/syncmember"
	"github.com/ciiim/syncmember/signal"
)

func main() {
	si := signal.NewManager()
	s1 := syncmember.NewSyncMember(syncmember.DebugConfig.SetPort(9000))
	s2 := syncmember.NewSyncMember(syncmember.DebugConfig.SetPort(9001))
	si.AddWatcher(os.Interrupt, "shutdown", func() {
		s2.Shutdown()
		s1.Shutdown()
	})
	go si.Wait()
	s2.Join("127.0.0.1:9000")
	go s1.Run()
	s2.Run()
}
