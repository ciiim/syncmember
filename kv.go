package syncmember

import (
	"bytes"

	"github.com/google/btree"
)

func (s *SyncMember) lazyInit() {
	if s.kvcopyTree == nil {
		s.kvcopyTree = btree.New(32)
	}
}

type kVItem struct {
	key   string
	value []byte
}

func (k *kVItem) Less(than btree.Item) bool {
	return k.key < than.(*kVItem).key
}

func newKVItem(key string, value []byte) *kVItem {
	return &kVItem{
		key:   key,
		value: value,
	}
}

func (s *SyncMember) kvOperation(op MessageType, kv *KeyValuePayload) {
	s.kvTreeMu.Lock()
	defer s.kvTreeMu.Unlock()
	s.lazyInit()
	item := newKVItem(kv.Key, kv.Value)

	shouldBoardcast := false
	var oldItem *kVItem

	switch op {
	case KVSet:
		shouldBoardcast = s.setKV(item)
	case KVDelete:
		shouldBoardcast, oldItem = s.deleteKV(item)
	case KVUpdate:
		shouldBoardcast, _ = s.updateKV(item)
	default:
		return
	}

	if shouldBoardcast {
		switch op {
		case KVSet:
			// 使用协程通知，防止阻塞
			go s.notifyKVWatcher(EventKVSet, item)
		case KVDelete:
			go s.notifyKVWatcher(EventKVDelete, oldItem)
		case KVUpdate:
			go s.notifyKVWatcher(EventKVUpdate, item)
		}
		//广播
		s.boardcastQueue.PutMessage(op, kv.Key, kv.Encode().Bytes())
	}
}

func (s *SyncMember) setKV(item *kVItem) bool {
	if exist := s.kvcopyTree.Has(item); exist {
		return false
	}

	s.kvcopyTree.ReplaceOrInsert(item)
	s.logger.Info("SetKV", "key", item.key)
	return true
}

func (s *SyncMember) deleteKV(item *kVItem) (bool, *kVItem) {
	deletedItem := s.kvcopyTree.Delete(item)
	if deletedItem == nil {
		return false, nil
	}
	s.logger.Info("DeleteKV", "key", item.key)
	return true, deletedItem.(*kVItem)
}

func (s *SyncMember) updateKV(item *kVItem) (bool, *kVItem) {
	getItem := s.kvcopyTree.Get(item)
	if getItem == nil {
		return false, nil
	}
	getBytes := getItem.(*kVItem).value
	if getBytes == nil {
		return false, nil
	}

	//如果值相同，不需要更新
	if bytes.Equal(getBytes, item.value) {
		return false, nil
	}

	old := s.kvcopyTree.ReplaceOrInsert(item)
	s.logger.Info("UpdateKV", "key", item.key)
	if old == nil {
		return false, nil
	}
	return true, old.(*kVItem)
}

func (s *SyncMember) SetKV(key string, value []byte) {
	s.kvOperation(KVSet, &KeyValuePayload{
		Key:   key,
		Value: value,
	})
}

func (s *SyncMember) DeleteKV(key string) {
	s.kvOperation(KVDelete, &KeyValuePayload{
		Key:   key,
		Value: nil,
	})
}

func (s *SyncMember) UpdateKV(key string, value []byte) {
	s.kvOperation(KVUpdate, &KeyValuePayload{
		Key:   key,
		Value: value,
	})
}

func (s *SyncMember) GetValue(key string) []byte {
	s.kvTreeMu.RLock()
	defer s.kvTreeMu.RUnlock()
	if s.kvcopyTree == nil {
		return nil
	}
	res := s.kvcopyTree.Get(newKVItem(key, nil))
	if res == nil {
		return nil
	}
	return res.(*kVItem).value
}
