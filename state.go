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
	if equalAddress(remoteNodeInfo.Addr, s.me.address) {
		return
	}
	// 如果节点不存在，添加节点
	if !ok {
		node = newNode(remoteNodeInfo.Addr, remoteNodeInfo)
		node.changeState(NodeAlive)
		node.becomeCredible()
		s.AddNode(node)

		if s.nodeEvent != nil {
			s.nodeEvent.NotifyJoin(node)
		}

		//广播
		s.boardcastQueue.PutMessage(Alive, remoteNodeInfo.Addr.String(), remoteNodeInfo.Encode().Bytes())

		s.logger.Info("New node added", "new node addr", remoteNodeInfo.Addr.String())
		return
	}

	if remoteNodeInfo.Version <= node.GetInfo().Version {
		return
	}

	s.logger.Info("Node Alive", "node", remoteNodeInfo.Addr.String())
	node.increaseVersionTo(remoteNodeInfo.Version)

	// 如果节点存在，但是状态不是存活，设置节点状态为存活
	if node.nodeLocalInfo.nodeState != NodeAlive {
		node.changeState(NodeAlive)
		node.becomeCredible()

		if s.nodeEvent != nil {
			s.nodeEvent.NotifyAlive(node)
		}

		//广播
		s.boardcastQueue.PutMessage(Alive, remoteNodeInfo.Addr.String(), remoteNodeInfo.Encode().Bytes())
	}
}

// 由远程节点发起的状态变更触发
// 也可以由心跳判断的状态变更触发
func (s *SyncMember) dead(remoteNodeInfo *NodeInfoPayload) {
	node, ok := s.nodesMap[remoteNodeInfo.Addr.String()]
	if equalAddress(remoteNodeInfo.Addr, s.me.address) {
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

	s.logger.Info("Node Dead", "node", remoteNodeInfo.Addr.String())
	node.increaseVersionTo(remoteNodeInfo.Version)

	// 如果节点存在，但是状态不是死亡，设置节点状态为死亡
	if node.nodeLocalInfo.nodeState != NodeDead {
		node.changeState(NodeDead)
		node.becomeUnCredible()

		if s.nodeEvent != nil {
			s.nodeEvent.NotifyDead(node)
		}

		//广播
		s.boardcastQueue.PutMessage(Dead, remoteNodeInfo.Addr.String(), remoteNodeInfo.Encode().Bytes())
	}
}

func (s *SyncMember) refute() {
	//广播
	payload := s.me.GetInfo()
	s.boardcastQueue.PutMessage(Alive, s.me.address.Name, payload.Encode().Bytes())
}

func (s *SyncMember) GetNodeState(addr string) NodeStateType {
	s.nMutex.Lock()
	defer s.nMutex.Unlock()
	node, ok := s.nodesMap[addr]
	if !ok {
		return NodeUnknown
	}
	return node.nodeLocalInfo.nodeState
}
