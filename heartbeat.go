package syncmember

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
	s.clearwaitAckMap()
	s.nMutex.Lock()
	defer s.nMutex.Unlock()
	s.logger.Debug("HeartBeat", "node number", len(s.nodes))
	if len(s.nodes) == 0 {
		return
	}
	nodes := kRamdonNodes(s.config.Fanout, s.nodes, func(n *Node) bool {
		return !n.IsCredible()
	})
	for _, node := range nodes {
		msg := NewHeartBeatMessage(s.host, node.Addr())
		if err := SendMsg(s.udpTransport, msg); err != nil {
			s.logger.Error("SendMsg", "error", err)
		}
		s.waitAckMap[node.address.String()] = node
		s.logger.Info("HeartBeat", "target node", node.Addr())
	}
}

func (s *SyncMember) clearwaitAckMap() {
	for k, node := range s.waitAckMap {
		if node.nodeLocalInfo.credibility.Load()-1 == 0 {
			s.logger.Debug("[Heartbeat]credibility is zero", "dead node", node.Addr())

			node.SetDead()
			delete(s.waitAckMap, k)
		} else {
			node.nodeLocalInfo.credibility.Add(-1)
		}
	}
}

func (s *SyncMember) handleAckHeartbeat(msg IMessage) {
	s.logger.Info("AckHeartbeat", "health node", msg.BMessage().From)
	s.nMutex.Lock()
	defer s.nMutex.Unlock()
	from := msg.BMessage().From.String()

	//如果一段时间后才收到ACK，且节点为死亡状态，转变为存活节点
	node, ok := s.nodesMap[from]
	if ok && node.nodeLocalInfo.nodeState == NodeDead {
		s.logger.Warn("AckHeartbeat Find a DeadNode become alive", "this node", node)
		node.SetAlive()
		return
	}
	if !ok {
		s.logger.Warn("Unknown AckHeartbeat Message", "From", msg.BMessage().From)
		return
	}

	//如果收到ACK，且节点为存活状态，增加可信度
	node, ok = s.waitAckMap[from]
	if !ok {
		// s.logger.Warn("Unknown AckHeartbeat Message", "From", msg.From)
		return
	}
	node.BecomeCredible()
	delete(s.waitAckMap, from)
}

// 由PacketHandler触发
func (s *SyncMember) handleHeartbeat(msg IMessage) {
	s.nMutex.Lock()
	_, ok := s.nodesMap[msg.BMessage().From.String()]
	s.nMutex.Unlock()
	if !ok {
		s.logger.Warn("Received an unknown Heartbeat", "node addr", msg.BMessage().From)
		node := NewNode(msg.BMessage().From)
		s.AddNode(node)
		alive(node, msg)
	} else {
		//创建一个Ack消息
		ackMsg := NewHeartBeatAckMessage(s.host, msg.BMessage().From)
		if err := SendMsg(s.udpTransport, ackMsg); err != nil {
			s.logger.Error("SendMsg", "error", err)
		}
	}

}
