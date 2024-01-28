package syncmember

import "sync"

type KVEventType int8

type KVEventFunc func(kv *KeyValuePayload)

const (
	EventKVSet KVEventType = iota
	EventKVDelete
	EventKVUpdate
)

const (
	KVEventNums int8 = 3
)

type kVWatcher struct {
	watcherMu sync.Mutex
	// key -> [set, delete, update]
	watchers map[string][KVEventNums]KVEventFunc
}

func newKVWatcher() *kVWatcher {
	return &kVWatcher{
		watchers: make(map[string][KVEventNums]KVEventFunc),
	}
}

func (k *kVWatcher) setWatcher(key string, kvEventType KVEventType, fn KVEventFunc) {
	k.watcherMu.Lock()
	defer k.watcherMu.Unlock()
	if _, ok := k.watchers[key]; !ok {
		k.watchers[key] = [3]KVEventFunc{}
	}
	watchers := k.watchers[key]
	watchers[kvEventType] = fn
}

func (k *kVWatcher) removeWatcher(key string, kvEventType KVEventType) {
	k.watcherMu.Lock()
	defer k.watcherMu.Unlock()
	watchers, ok := k.watchers[key]
	if !ok {
		return
	}
	watchers[kvEventType] = nil
}

func (s *SyncMember) SetKVWatcher(key string, kvEventType KVEventType, fn KVEventFunc) {
	s.kWatcher.setWatcher(key, kvEventType, fn)
}

func (s *SyncMember) RemoveKVWatcher(key string, kvEventType KVEventType) {
	s.kWatcher.removeWatcher(key, kvEventType)
}

func (s *SyncMember) notifyKVWatcher(eventType KVEventType, kv *KeyValuePayload) {
	s.kWatcher.watcherMu.Lock()
	defer s.kWatcher.watcherMu.Unlock()
	watchers, ok := s.kWatcher.watchers[kv.Key]
	if !ok {
		return
	}
	if watchers[eventType] == nil {
		return
	}
	watchers[eventType](kv)
}
