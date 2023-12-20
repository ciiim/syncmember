package syncmember

type NodeStateType int8

const (
	NodeUnknown NodeStateType = iota
	NodeDead
	NodeAlive
)

// 由远程节点发起的状态变更触发
// 也可以由心跳判断的状态变更触发
func (s *SyncMember) alive(remoteNodeInfo *NodeInfoPayload) {
	node, ok := s.nodesMap[remoteNodeInfo.Addr.String()]
	if EqualAddress(remoteNodeInfo.Addr, s.me.address) {
		return
	}
	// 如果节点不存在，添加节点
	if !ok {
		node = NewNode(remoteNodeInfo.Addr, remoteNodeInfo)
		node.ChangeState(NodeAlive)
		node.BecomeCredible()
		s.AddNode(node)

		//广播
		msg := NewMessage(Alive, remoteNodeInfo.Encode().Bytes())
		b := NewGossipBoardcast(s.me.address.Name, msg)
		s.boardcastQueue.PutGossipBoardcast(b)

		s.logger.Info("add remote node", "me", s.me.address.Name, "remote node", remoteNodeInfo)
		return
	}

	if remoteNodeInfo.Version <= node.GetInfo().Version {
		return
	}

	s.logger.Info("node become alive", "me", s.me.address.Name, "node", remoteNodeInfo.Addr, "remote version", remoteNodeInfo.Version, "local version", node.GetInfo().Version)
	node.IncreaseVersionTo(remoteNodeInfo.Version)

	// 如果节点存在，但是状态不是存活，设置节点状态为存活
	if node.nodeLocalInfo.nodeState != NodeAlive {
		node.ChangeState(NodeAlive)
		node.BecomeCredible()

		//广播
		msg := NewMessage(Alive, remoteNodeInfo.Encode().Bytes())
		b := NewGossipBoardcast(s.me.address.Name, msg)
		s.boardcastQueue.PutGossipBoardcast(b)
	}
}

// 由远程节点发起的状态变更触发
// 也可以由心跳判断的状态变更触发
func (s *SyncMember) dead(remoteNodeInfo *NodeInfoPayload) {
	node, ok := s.nodesMap[remoteNodeInfo.Addr.String()]
	if EqualAddress(remoteNodeInfo.Addr, s.me.address) {
		return
	}
	// 如果该死亡节点不存在，不需要处理
	if !ok {
		return
	}

	if remoteNodeInfo.Version <= node.GetInfo().Version {
		return
	}

	//如果收到的死亡节点是自己，需要反驳
	if remoteNodeInfo.Addr.String() == s.me.address.String() {
		//反驳
		s.refute()
		return
	}

	s.logger.Info("node become dead", "me", s.me.address.Name, "node", remoteNodeInfo.Addr, "remote version", remoteNodeInfo.Version, "local version", node.GetInfo().Version)
	node.IncreaseVersionTo(remoteNodeInfo.Version)

	// 如果节点存在，但是状态不是死亡，设置节点状态为死亡
	if node.nodeLocalInfo.nodeState != NodeDead {
		node.ChangeState(NodeDead)
		node.BecomeUnCredible()

		//广播
		msg := NewMessage(Dead, remoteNodeInfo.Encode().Bytes())
		b := NewGossipBoardcast(s.me.address.Name, msg)
		s.boardcastQueue.PutGossipBoardcast(b)
	}
}

func (s *SyncMember) refute() {
	//广播
	payload := s.me.GetInfo()
	buf := payload.Encode()
	msg := NewMessage(MessageType(s.me.GetInfo().NodeState), buf.Bytes())
	b := NewGossipBoardcast(s.me.address.Name, msg)
	s.boardcastQueue.PutGossipBoardcast(b)
}
