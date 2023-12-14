package syncmember

import "github.com/ciiim/syncmember/codec"

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

}

func (s *SyncMember) handleGossip(msg *Message) {

	switch msg.MsgType {
	case Alive:
		fallthrough
	case Dead:
		s.handleStateChange(msg)
	case KVSet:
		fallthrough
	case KVDelete:
		fallthrough
	case KVUpdate:
		s.handleKV(msg)
	}
}

func (s *SyncMember) handleStateChange(msg *Message) {
	nodeinfo := NodeInfoPayload{}
	if err := codec.UDPUnmarshal(msg.Payload, &nodeinfo); err != nil {
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
	if err := codec.UDPUnmarshal(msg.Payload, &kv); err != nil {
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
