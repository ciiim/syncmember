package syncmember

import (
	"log"
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
	msg  *Message
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

func NewGossipBoardcast(name string, msg *Message) *GossipBoardcast {
	return &GossipBoardcast{
		name: name,
		msg:  msg,
		life: 3,
	}
}

func (g *GossipBoardcast) Less(than btree.Item) bool {
	//比较消息长度，若相同则比较生命值
	if len(g.msg.GetPayload()) < len(than.(*GossipBoardcast).msg.GetPayload()) {
		return true
	} else if len(g.msg.GetPayload()) == len(than.(*GossipBoardcast).msg.GetPayload()) {
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

func (b *BoardcastQueue) GetGossipBoardcast(availableBytes int) []*Message {
	b.mu.Lock()
	defer b.mu.Unlock()
	//每次取出最小的，如果小于limitBytes，就删除

	if b.tq == nil {
		return nil
	}

	msgs := make([]*Message, 0)
	reinsert := make([]*GossipBoardcast, 0)
	b.tq.Ascend(func(i btree.Item) bool {
		gb := i.(*GossipBoardcast)
		if gb.msg == nil {
			log.Fatalf("gb.msg is nil")
		}
		if gb.msg.GetPayload() == nil {
			log.Fatalf("gb.msg.GetPayload() is nil")
		}
		if availableBytes < len(gb.msg.GetPayload())+1+8 { // + 1 MessageType in8 + 8 Seq int64
			return false
		}
		msgs = append(msgs, gb.msg)
		availableBytes -= len(gb.msg.GetPayload()) + 1 + 8
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
