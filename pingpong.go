package syncmember

func (s *SyncMember) ping() {
	for {
		select {
		case <-s.pingTicker.C:
			s.doPing()
		case <-s.stopCh:
			return
		}
	}
}

func (s *SyncMember) doPing() {
	s.clearwaitPongMap()
	s.nMutex.Lock()
	defer s.nMutex.Unlock()
	s.logger.Debug("Ping", "node list length", len(s.nodes))
	if len(s.nodes) == 0 {
		return
	}
	nodes := kRamdonNodes(s.config.Fanout, s.nodes, func(n *Node) bool {
		return !n.IsCredible()
	})
	for _, node := range nodes {
		msg := NewPingMessage(s.host, node.Addr())
		if err := SendMsg(s.udpTransport, msg); err != nil {
			s.logger.Error("SendMsg", "error", err)
		}
		s.waitPongMap[node.address.String()] = node
		s.logger.Debug("Ping", "target node", node.Addr())
	}
}

func (s *SyncMember) clearwaitPongMap() {
	for k, node := range s.waitPongMap {
		if node.nodeLocalInfo.credibility.Load()-1 <= 0 {
			s.logger.Debug("[Ping]credibility is zero", "dead node", node.Addr())
			s.logger.Info("Dead Node", "node", node.Addr())
			node.SetDead()
			delete(s.waitPongMap, k)
		} else {
			node.nodeLocalInfo.credibility.Add(-1)
		}
	}
}

func (s *SyncMember) handlePong(msg *Message) {
	s.logger.Debug("PongPing", "health node", msg.From)
	s.nMutex.Lock()
	defer s.nMutex.Unlock()
	from := msg.From.String()

	//如果一段时间后才收到Pong，且节点为死亡状态，转变为存活节点
	node, ok := s.nodesMap[from]
	if ok && node.nodeLocalInfo.nodeState == NodeDead {
		s.logger.Warn("Pong Find a DeadNode become alive", "this node", node)
		node.SetAlive()
		return
	}
	if !ok {
		s.logger.Warn("Unknown Pong Message", "From", msg.From)
		return
	}

	//如果收到Pong，且节点为存活状态，增加可信度
	node, ok = s.waitPongMap[from]
	if !ok {
		s.logger.Warn("Unknown Pong Message", "From", msg.From)
		return
	}
	node.BecomeCredible()
	delete(s.waitPongMap, from)
}

// 由PongHandler触发
func (s *SyncMember) handlePing(msg *Message) {
	s.nMutex.Lock()
	_, ok := s.nodesMap[msg.From.String()]
	s.nMutex.Unlock()
	if !ok {
		s.logger.Warn("Received an unknown Ping", "node addr", msg.From)
	} else {
		//创建一个Pong消息
		PongMsg := NewPongMessage(s.host, msg.From, msg.Seq)
		if err := SendMsg(s.udpTransport, PongMsg); err != nil {
			s.logger.Error("SendMsg", "error", err)
		}
	}

}
