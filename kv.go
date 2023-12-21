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

type KVItem struct {
	key   string
	value []byte
}

func (k *KVItem) Less(than btree.Item) bool {
	return k.key < than.(*KVItem).key
}

func NewKVItem(key string, value []byte) *KVItem {
	return &KVItem{
		key:   key,
		value: value,
	}
}

func (s *SyncMember) setKV(kv *KeyValuePayload) {
	s.kvTreeMu.Lock()
	defer s.kvTreeMu.Unlock()
	s.lazyInit()
	item := NewKVItem(kv.Key, kv.Value)

	if exist := s.kvcopyTree.Has(item); exist {
		return
	}

	s.logger.Info("SetKV", "key", kv.Key)
	s.kvcopyTree.ReplaceOrInsert(item)

	//广播
	s.boardcastQueue.PutMessage(KVSet, kv.Key, kv.Encode().Bytes())
}

func (s *SyncMember) deleteKV(kv *KeyValuePayload) {
	s.kvTreeMu.Lock()
	defer s.kvTreeMu.Unlock()
	s.lazyInit()
	item := NewKVItem(kv.Key, kv.Value)

	s.logger.Info("DeleteKV", "key", kv.Key)
	if deletedItem := s.kvcopyTree.Delete(item); deletedItem == nil {
		return
	}

	//广播
	s.boardcastQueue.PutMessage(KVDelete, kv.Key, kv.Encode().Bytes())
}

func (s *SyncMember) updateKV(kv *KeyValuePayload) {
	s.kvTreeMu.Lock()
	defer s.kvTreeMu.Unlock()
	s.lazyInit()
	item := NewKVItem(kv.Key, kv.Value)

	getItem := s.kvcopyTree.Get(item)
	if getItem == nil {
		return
	}
	getBytes := getItem.(*KVItem).value
	if getBytes == nil {
		return
	}

	//如果值相同，不需要更新
	if bytes.Equal(getBytes, kv.Value) {
		return
	}

	s.logger.Info("UpdateKV", "key", kv.Key)
	s.kvcopyTree.ReplaceOrInsert(item)

	//广播
	s.boardcastQueue.PutMessage(KVUpdate, kv.Key, kv.Encode().Bytes())
}

func (s *SyncMember) SetKV(key string, value []byte) {
	s.setKV(&KeyValuePayload{
		Key:   key,
		Value: value,
	})
}

func (s *SyncMember) DeleteKV(key string) {
	s.deleteKV(&KeyValuePayload{
		Key: key,
	})
}

func (s *SyncMember) UpdateKV(key string, value []byte) {
	s.updateKV(&KeyValuePayload{
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
	res := s.kvcopyTree.Get(NewKVItem(key, nil))
	if res == nil {
		return nil
	}
	return res.(*KVItem).value
}
