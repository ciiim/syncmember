package syncmember

type NodeEventType int8

type NodeEventDelegate interface {

	// 当新节点加入时被调用
	NotifyJoin(n *Node)

	// 当已存在的节点变为存活时被调用
	NotifyAlive(n *Node)

	// 当已存在的节点变为死亡时被调用
	NotifyDead(n *Node)
}

const (
	DelegateNodeAlive NodeEventType = iota
	DelegateNodeDead
)

func (s *SyncMember) SetNodeDelegate(delegate NodeEventDelegate) {
	s.nodeEvent = delegate
}
