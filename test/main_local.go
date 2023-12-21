package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/ciiim/syncmember"
	"github.com/ciiim/syncmember/signal"
)

const (
	LocalAddr = "172.26.123.188"
)

func main() {

	stopCh := make(chan struct{})

	si := signal.NewManager()
	s1 := syncmember.NewSyncMember("node1", syncmember.DebugConfig().SetPort(9001).SetLogLevel(slog.LevelInfo).OpenLogDetail(false))
	s2 := syncmember.NewSyncMember("node2", syncmember.DebugConfig().SetPort(9002).SetLogLevel(slog.LevelInfo).OpenLogDetail(false))
	s3 := syncmember.NewSyncMember("node3", syncmember.DebugConfig().SetPort(9003).SetLogLevel(slog.LevelInfo).OpenLogDetail(false))
	// s4 := syncmember.NewSyncMember("node4", syncmember.DebugConfig().SetPort(9004).SetLogLevel(slog.LevelInfo).OpenLogDetail(false))
	si.AddWatcher(os.Interrupt, "shutdown", func() {
		s1.Shutdown()
		s2.Shutdown()
		s3.Shutdown()
		func() {
			stopCh <- struct{}{}
			os.Exit(0)
		}()
		// s4.Shutdown()
	})
	si.Wait()
	go s1.Run()
	go s2.Run()
	go s3.Run()
	// go s4.Run()
	s1.Join(LocalAddr + ":9002")
	s2.Join(LocalAddr + ":9003")
	time.Sleep(time.Second * 1)
	s3.SetKV("key1", []byte("value1"))
	s3.SetKV("key2", []byte("value2"))
	time.Sleep(time.Second * 10)
	if res := s2.GetValue("key1"); res == nil {
		println("key1 not found")
	} else {
		println(string(res))
	}
	s2.Shutdown()
	<-stopCh
}
