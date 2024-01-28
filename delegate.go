package syncmember

type NodeEventType int8

type NodeEventDelegate interface {
	NotifyJoin(n *Node)

	NotifyAlive(n *Node)

	NotifyDead(n *Node)
}

const (
	DelegateNodeAlive NodeEventType = iota
	DelegateNodeDead
)

func (s *SyncMember) SetNodeDelegate(delegate NodeEventDelegate) {
	s.nodeEvent = delegate
}
