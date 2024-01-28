package syncmember

import "sync"

type KVEventType int8

type KV kVItem

/*
KVEventFunc

EventKVSet kv <- after set

EventKVDelete kv <- before delete

EventKVUpdate kv <- after update
*/
type KVEventFunc func(kv *KV)

const (
	EventKVSet KVEventType = iota
	EventKVDelete
	EventKVUpdate
)

const (
	KVEventNums int8 = 3
)

var kvWaitExecuteFn = func(s *SyncMember, kvEventType KVEventType, key string, c chan<- []byte) func(kv *KV) {
	return func(kv *KV) {
		clone := make([]byte, len(kv.value))
		copy(clone, kv.value)

		// 删除watcher
		s.kWatcher.removeWatcher(key, kvEventType)

		c <- clone

		// 关闭channel
		close(c)
	}
}

type kVWatcher struct {
	// key -> [set, delete, update]
	// 置于结构体第一位
	watchers  map[string][]KVEventFunc
	watcherMu sync.Mutex
}

func newKVWatcher() *kVWatcher {
	return &kVWatcher{
		watchers: make(map[string][]KVEventFunc),
	}
}

func (k *kVWatcher) setWatcher(key string, kvEventType KVEventType, fn KVEventFunc) {
	k.watcherMu.Lock()
	defer k.watcherMu.Unlock()
	_, ok := k.watchers[key]
	if !ok {
		k.watchers[key] = make([]KVEventFunc, KVEventNums)
	}

	k.watchers[key][kvEventType] = fn

}

func (k *kVWatcher) removeWatcher(key string, kvEventType KVEventType) {
	k.watcherMu.Lock()
	defer k.watcherMu.Unlock()
	watchers, ok := k.watchers[key]
	if !ok {
		return
	}

	watchers[kvEventType] = nil

	if watchers[EventKVSet] == nil && watchers[EventKVDelete] == nil && watchers[EventKVUpdate] == nil {
		delete(k.watchers, key)
	}
}

func (s *SyncMember) SetKVWatcher(key string, kvEventType KVEventType, fn KVEventFunc) {
	s.kWatcher.setWatcher(key, kvEventType, fn)
}

func (s *SyncMember) RemoveKVWatcher(key string, kvEventType KVEventType) {
	s.kWatcher.removeWatcher(key, kvEventType)
}

func (s *SyncMember) notifyKVWatcher(kvEventType KVEventType, item *kVItem) {
	fn := s.getKVWatcherFunc(item.key, kvEventType)

	if fn == nil {
		return
	}
	fn((*KV)(item))
}

func (s *SyncMember) getKVWatcherFunc(key string, kvEventType KVEventType) KVEventFunc {
	s.kWatcher.watcherMu.Lock()
	defer s.kWatcher.watcherMu.Unlock()
	watchers, ok := s.kWatcher.watchers[key]

	// 该key没有被监控
	if !ok {
		return nil
	}

	return watchers[kvEventType]
}

/*
WaitKVSet 用于等待某个key被设置

return 设置的值

会覆盖之前的watcher
*/
func (s *SyncMember) WaitKVSet(key string) <-chan []byte {
	c := make(chan []byte)
	s.kWatcher.setWatcher(key, EventKVSet, kvWaitExecuteFn(s, EventKVSet, key, c))
	return c
}

/*
WaitKVDelete 用于等待某个key被删除

return 被删除的值

会覆盖之前的watcher
*/
func (s *SyncMember) WaitKVDelete(key string) <-chan []byte {
	c := make(chan []byte)
	s.kWatcher.setWatcher(key, EventKVDelete, kvWaitExecuteFn(s, EventKVDelete, key, c))
	return c
}

/*
WaitKVUpdate 用于等待某个key被更新

return 更新后的值

会覆盖之前的watcher
*/
func (s *SyncMember) WaitKVUpdate(key string) <-chan []byte {
	c := make(chan []byte)
	s.kWatcher.setWatcher(key, EventKVUpdate, kvWaitExecuteFn(s, EventKVUpdate, key, c))
	return c
}
