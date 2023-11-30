package syncmember

type NodeStateType int8

const (
	NodeUnknown NodeStateType = iota
	NodeDead
	NodeAlive
)

// 由远程节点发起的状态变更触发
// 也可以由心跳判断的状态变更触发
func (s *SyncMember) alive(remoteNodeInfo *NodeInfo) {
	node, ok := s.nodesMap[remoteNodeInfo.Addr.String()]
	if EqualAddress(remoteNodeInfo.Addr, s.me.address) {
		return
	}
	// 如果节点不存在，添加节点
	if !ok {
		node = NewNode(remoteNodeInfo.Addr, remoteNodeInfo)
		s.AddNode(node)
		node.ChangeState(NodeAlive)
		node.BecomeCredible()
		s.logger.Info("add remote node", "me", s.me.address.Name, "remote node", remoteNodeInfo)
		return
	}

	if node.GetInfo().Version == remoteNodeInfo.Version {
		s.logger.Debug("alive remote node version is same", "remote node", remoteNodeInfo, "local node", node.GetInfo())
		return
	}

	//判断远程节点版本
	if node.GetInfo().Version > remoteNodeInfo.Version {
		s.logger.Warn("alive remote node version is old", "me", s.me.address.Name, "remote node", remoteNodeInfo, "local node", node.GetInfo())
		return
	}
	s.logger.Info("alive remote node version is new", "me", s.me.address.Name, "remote node", remoteNodeInfo, "local node", node.GetInfo())
	node.IncreaseVersionTo(remoteNodeInfo.Version)

	// 如果节点存在，但是状态不是存活，设置节点状态为存活
	if node.nodeLocalInfo.nodeState != NodeAlive {
		node.ChangeState(NodeAlive)
		node.BecomeCredible()
	}
}

// 由远程节点发起的状态变更触发
// 也可以由心跳判断的状态变更触发
func (s *SyncMember) dead(remoteNodeInfo *NodeInfo) {
	node, ok := s.nodesMap[remoteNodeInfo.Addr.String()]
	if EqualAddress(remoteNodeInfo.Addr, s.me.address) {
		return
	}
	// 如果该死亡节点不存在，不需要处理
	if !ok {
		return
	}

	if node.GetInfo().Version == remoteNodeInfo.Version {
		s.logger.Debug("alive remote node version is same", "remote node", remoteNodeInfo, "local node", node.GetInfo())
		return
	}

	//判断远程节点版本
	if node.GetInfo().Version > remoteNodeInfo.Version {
		s.logger.Warn("dead remote node version is old", "me", s.me.address.Name, "remote node", remoteNodeInfo, "local node", node.GetInfo())
		return
	}
	s.logger.Info("dead remote node version is new", "me", s.me.address.Name, "remote node", remoteNodeInfo, "local node", node.GetInfo())
	node.IncreaseVersionTo(remoteNodeInfo.Version)

	// 如果节点存在，但是状态不是死亡，设置节点状态为死亡
	if node.nodeLocalInfo.nodeState != NodeDead {
		node.ChangeState(NodeDead)
		node.BecomeUnCredible()
	}
}
