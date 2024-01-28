package main

import (
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"time"

	"github.com/ciiim/syncmember"
	"github.com/ciiim/syncmember/signal"
)

const (
	LocalAddr = "127.0.0.1"
)

func main() {

	stopCh := make(chan struct{})

	si := signal.NewManager()
	s1 := syncmember.NewSyncMember("node1", syncmember.DebugConfig().SetPort(9001).SetLogLevel(slog.LevelInfo).OpenLogDetail(false))
	s2 := syncmember.NewSyncMember("node2", syncmember.DebugConfig().SetPort(9002).SetLogLevel(slog.LevelInfo).OpenLogDetail(false))
	s3 := syncmember.NewSyncMember("node3", syncmember.DebugConfig().SetPort(9003).SetLogLevel(slog.LevelInfo).OpenLogDetail(false))
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
	s2.Join(LocalAddr + ":9003")
	s1.Join(LocalAddr + ":9002")
	time.Sleep(time.Second * 2)
	for i := 0; i < 10; i++ {
		// time.Sleep(time.Millisecond * 50)
		v := fmt.Sprintf("key%d", i)
		s1.SetKV(v, []byte(v))
	}
	time.Sleep(time.Millisecond * 3000)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 5; i++ {
		v := fmt.Sprintf("key%d", r.Intn(10))
		if res := s2.GetValue(v); res == nil {
			println(v + "not found")
		} else {
			println(string(res))
		}
	}
	s2.Shutdown()
	<-stopCh
}
