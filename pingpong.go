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
	var packet *Packet
	for _, node := range nodes {
		packet = NewPacket(NewPingMessage(), s.host, node.Addr())
		if err := SendPacket(s.udpTransport, packet); err != nil {
			s.logger.Error("SendMsg", "error", err)
		}
		s.waitPongMap[node.address.String()] = node
		s.logger.Debug("Ping", "target node", node.Addr())
	}
}

func (s *SyncMember) clearwaitPongMap() {
	for k, node := range s.waitPongMap {
		if node.nodeLocalInfo.credibility.Load()-1 <= 0 {
			s.logger.Info("[Ping failed]Node Become Dead", "node addr", node.Addr().String())
			node.SetDead()
			delete(s.waitPongMap, k)
			//添加广播
			nodePayload := &NodeInfoPayload{
				Addr:      node.Addr(),
				NodeState: NodeDead,
				Version:   node.GetInfo().Version,
			}
			s.boardcastQueue.PutMessage(Dead, node.Addr().String(), nodePayload.Encode().Bytes())
		} else {
			node.nodeLocalInfo.credibility.Add(-1)
		}
	}
}

func (s *SyncMember) handlePong(packet *Packet) {
	s.logger.Debug("PongPing", "health node", packet.From)
	s.nMutex.Lock()
	defer s.nMutex.Unlock()
	from := packet.From.String()

	//如果一段时间后才收到Pong，且节点为死亡状态，转变为存活节点
	node, ok := s.nodesMap[from]
	if ok && node.nodeLocalInfo.nodeState == NodeDead {
		s.logger.Warn("Pong Find a DeadNode become alive", "this node", node)
		node.SetAlive()
		return
	}
	if !ok {
		s.logger.Warn("Unknown Pong Message", "From", packet.From)
		return
	}

	//如果收到Pong，且节点为存活状态，增加可信度
	node, ok = s.waitPongMap[from]
	if !ok {
		s.logger.Warn("Unknown Pong Message", "From", packet.From)
		return
	}
	node.BecomeCredible()
	delete(s.waitPongMap, from)
}

// 由PongHandler触发
func (s *SyncMember) handlePing(packet *Packet) {
	s.nMutex.Lock()
	_, ok := s.nodesMap[packet.From.String()]
	s.nMutex.Unlock()
	if !ok {
		s.logger.Warn("Received an unknown Ping", "node addr", packet.From)
	} else {
		//创建一个Pong消息
		PongPacket := NewPacket(NewPongMessage(packet.MessageBody.Seq), s.host, packet.From)
		if err := SendPacket(s.udpTransport, PongPacket); err != nil {
			s.logger.Error("SendMsg", "error", err)
		}
	}

}
