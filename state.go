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

	// 版本小于等于当前节点版本，若本地副本状态为死亡，设置为存活；若本地副本状态为存活，返回
	if remoteNodeInfo.Version <= node.GetInfo().Version && node.nodeLocalInfo.nodeState == NodeAlive {
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
	//如果收到的死亡节点是自己，需要反驳
	if equalAddress(remoteNodeInfo.Addr, s.me.address) {

		//	当集群内的一个节点误认为自己死亡时，会发送Gossip消息通知其他节点
		//	其他节点也会发送Gossip消息通知其他节点
		//	该节点会多次收到同样节点版本的死亡通知
		//	如果该节点收到的死亡通知版本和自己的版本一致，说明该节点已经反驳过了，不需要再次反驳
		if remoteNodeInfo.Version == s.me.GetInfo().Version {
			return
		}
		//同步版本
		s.me.increaseVersionTo(remoteNodeInfo.Version)

		//反驳
		s.refute()
		return
	}

	node, ok := s.nodesMap[remoteNodeInfo.Addr.String()]

	// 如果该死亡节点不存在，不需要处理
	if !ok {
		return
	}

	if remoteNodeInfo.Version <= node.GetInfo().Version {
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

	s.logger.Info("[Refute] I'm alive", "node", s.me.address.Name)
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
