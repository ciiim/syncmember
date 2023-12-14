package syncmember

import (
	"sync"

	"github.com/google/btree"
)

type Boardcast interface {
	Less(than btree.Item) bool
}

type BoardcastQueue struct {
	mu sync.Mutex
	tq *btree.BTree
}

type GossipBoardcast struct {
	name string
	msg  []byte
	life int
}

func NewBoardcastQueue() *BoardcastQueue {
	return &BoardcastQueue{}
}

func (b *BoardcastQueue) lazyInit() {
	if b.tq == nil {
		b.tq = btree.New(32)
	}
}

func NewGossipBoardcast(name string, msg []byte) *GossipBoardcast {
	return &GossipBoardcast{
		name: name,
		msg:  msg,
		life: 3,
	}
}

func (g *GossipBoardcast) Less(than btree.Item) bool {
	//比较消息长度，若相同则比较生命值
	if len(g.msg) < len(than.(*GossipBoardcast).msg) {
		return true
	} else if len(g.msg) == len(than.(*GossipBoardcast).msg) {
		if g.life < than.(*GossipBoardcast).life {
			return true
		}
	}

	return false
}

func (b *BoardcastQueue) PutGossipBoardcast(item Boardcast) btree.Item {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.lazyInit()
	gb, ok := item.(*GossipBoardcast)
	if ok {
		has := b.tq.Has(gb)
		if has {
			b.tq.Delete(gb)
			return nil
		}
	}
	return b.tq.ReplaceOrInsert(item)

}

func (b *BoardcastQueue) GetGossipBoardcast(availableBytes int) [][]byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	//每次取出最小的，如果小于limitBytes，就删除
	msgs := make([][]byte, 0)
	reinsert := make([]*GossipBoardcast, 0)
	b.tq.Ascend(func(i btree.Item) bool {
		gb := i.(*GossipBoardcast)
		if availableBytes < len(gb.msg) {
			return false
		}
		msgs = append(msgs, gb.msg)
		availableBytes -= len(gb.msg)
		if gb.life > 0 {
			gb.life--
			reinsert = append(reinsert, gb)
		}
		b.tq.Delete(i)
		return true
	})
	for _, gb := range reinsert {
		b.tq.ReplaceOrInsert(gb)
	}
	return msgs
}
