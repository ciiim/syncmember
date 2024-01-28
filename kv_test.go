package syncmember_test

import (
	"testing"
	"time"

	"github.com/ciiim/syncmember"
	"github.com/stretchr/testify/assert"
)

func TestKVWatcher(t *testing.T) {
	s1 := syncmember.NewSyncMember("node1", syncmember.DefaultConfig())

	defer s1.Shutdown()

	setCh := s1.WaitKVSet("key1")
	updateCh := s1.WaitKVUpdate("key1")
	deleteCh := s1.WaitKVDelete("key1")

	s1.SetKV("key1", []byte("value1"))

	select {
	case v := <-setCh:
		assert.Equal(t, "value1", string(v))
	case <-time.After(time.Second * 5):
		t.Fatal("key1 set timeout")
	}

	s1.UpdateKV("key1", []byte("value1update"))

	select {
	case v := <-updateCh:
		assert.Equal(t, "value1update", string(v))
	case <-time.After(time.Second * 5):
		t.Fatal("key1 update timeout")
	}

	s1.DeleteKV("key1")

	select {
	case v := <-deleteCh:
		assert.Equal(t, "value1update", string(v))
	case <-time.After(time.Second * 5):
		t.Fatal("key1 delete timeout")
	}
}
