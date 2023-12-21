package syncmember

import (
	"github.com/ciiim/syncmember/codec"
)

func (s *SyncMember) gossip() {
	for {
		select {
		case <-s.gossipTicker.C:
			s.doGossip()
		case <-s.stopCh:
			return
		}
	}
}

func (s *SyncMember) doGossip() {
	//取出Boardcast
	availableBytes := s.config.UDPBufferSize
	messages := s.boardcastQueue.GetGossipBoardcast(availableBytes)
	if len(messages) == 0 {
		return
	}
	//广播
	s.nMutex.Lock()
	defer s.nMutex.Unlock()
	nodes := kRamdonNodes(s.config.Fanout, s.nodes, func(n *Node) bool {
		return !n.IsCredible()
	})
	for _, node := range nodes {
		for _, msgBytes := range messages { //XXX: 可以组装成一个包，无需多次发送
			packet := buildPacketMessageBytes(msgBytes, s.host, node.Addr())
			if err := SendPacket(s.udpTransport, packet); err != nil {
				s.logger.Error("SendMsg", "error", err)
				continue
			}
		}
	}

}

func buildPacketMessageBytes(msg *Message, from, to Address) *Packet {
	return NewPacket(msg, from, to)
}

func (s *SyncMember) handleGossip(packet *Packet) {

	switch packet.MessageBody.MsgType {
	case Alive:
		fallthrough
	case Dead:
		s.handleStateChange(packet.MessageBody)
	case KVSet:
		fallthrough
	case KVDelete:
		fallthrough
	case KVUpdate:
		s.handleKV(packet.MessageBody)
	}
}

func (s *SyncMember) handleStateChange(msg *Message) {
	nodeinfo := NodeInfoPayload{}
	if err := codec.Unmarshal(msg.Payload, &nodeinfo); err != nil {
		s.logger.Error("handleStateChange", "UDPUnmarshal error", err)
		return
	}

	switch msg.MsgType {
	case Alive:
		s.alive(&nodeinfo)
	case Dead:
		s.dead(&nodeinfo)
	}
}

func (s *SyncMember) handleKV(msg *Message) {
	kv := KeyValuePayload{}
	if err := codec.Unmarshal(msg.Payload, &kv); err != nil {
		s.logger.Error("handleKV", "UDPUnmarshal error", err)
		return
	}

	switch msg.MsgType {
	case KVSet:
		s.setKV(&kv)
	case KVDelete:
		s.deleteKV(&kv)
	case KVUpdate:
		s.updateKV(&kv)
	}
}
