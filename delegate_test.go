package syncmember_test

import (
	"log/slog"
	"testing"
	"time"

	"github.com/ciiim/syncmember"
)

type MyDelegate struct {
	Joined chan struct{}
	Dead   chan struct{}
	Alive  chan struct{}
}

func NewMyDelegate() *MyDelegate {
	return &MyDelegate{
		Joined: make(chan struct{}),
		Dead:   make(chan struct{}),
		Alive:  make(chan struct{}),
	}
}

var _ syncmember.NodeEventDelegate = &MyDelegate{}

func (m *MyDelegate) NotifyJoin(n *syncmember.Node) {
	m.Joined <- struct{}{}
}

func (m *MyDelegate) NotifyDead(n *syncmember.Node) {
	m.Dead <- struct{}{}
}
func (m *MyDelegate) NotifyAlive(n *syncmember.Node) {
	m.Alive <- struct{}{}
}

func TestDelegate(t *testing.T) {
	s1 := syncmember.NewSyncMember("node1", syncmember.DefaultConfig().
		SetPort(9001).SetLogLevel(slog.LevelInfo))
	s2 := syncmember.NewSyncMember("node2", syncmember.DefaultConfig().
		SetPort(9002).SetLogLevel(slog.LevelError))

	defer s1.Shutdown()
	defer s2.Shutdown()

	delegate := NewMyDelegate()

	s1.SetNodeDelegate(delegate)

	go func() {
		_ = s1.Run()
	}()
	go func() {
		_ = s2.Run()
	}()

	//test join
	go func() {
		err := s2.Join("127.0.0.1:9001")
		if err != nil {
			t.Error(err)
		}
	}()
	select {
	case <-delegate.Joined:
		break
	case <-time.After(500 * time.Millisecond):
		t.Errorf("Joined expected true but found false")
		t.Fail()
	}

	//test dead
	s2.Shutdown()
	select {
	case <-delegate.Dead:
		break
	case <-time.After(3000 * time.Millisecond):
		t.Errorf("Dead expected true but found false")
		t.Fail()
	}

}
