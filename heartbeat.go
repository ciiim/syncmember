package syncmember

import (
	"bytes"

	"github.com/ciiim/syncmember/codec"
)

func (s *SyncMember) heartBeat() {
	for {
		select {
		case <-s.heartBeatTicker.C:
			s.doHeartBeat()
		case <-s.stopCh:
			return
		}
	}
}

func (s *SyncMember) doHeartBeat() {
	s.nMutex.Lock()
	s.logger.Debug("HeartBeat", "node number", len(s.nodes))
	if len(s.nodes) == 0 {
		s.nMutex.Unlock()
		return
	}
	nodes := kRamdonNodes(s.config.Fanout, s.nodes)
	s.nMutex.Unlock()
	for _, node := range nodes {
		msg := NewHeartBeatMessage(s.host, node.Addr())
		b, err := codec.UDPMarshal(msg)
		if err != nil {
			s.logger.Error("UDPMarshal error", "error", err)
			continue
		}
		buf := bytes.NewBuffer(b)
		s.udpTransport.SendRawMsg(buf, node.Addr().UDPAddr())
		s.logger.Info("doHeartBeat", "node", node.Addr())
	}
}

func (s *SyncMember) handleAckHeartbeat() {

}

// 由PacketHandler触发
func (s *SyncMember) handleHeartbeat(msg *Message) {
	_, ok := s.nodesMap[msg.From.String()]
	if !ok {
		s.logger.Warn("recived an unknown Heartbeat", "node addr", msg.From)
	} else {
		//创建一个Ack消息

	}

}
