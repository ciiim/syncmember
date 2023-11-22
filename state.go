package syncmember

type NodeStateType uint8

const (
	NodeUnknown NodeStateType = iota
	NodeDead
	NodeAlive
	NodeTimeout
)

func (n *Node) SetAlive() {
	if n.nodeLocalInfo.nodeState == NodeAlive {
		return
	}
	n.ChangeState(NodeAlive)
	n.IncreaseVersionTo(n.GetInfo().Version + 1)
	n.BecomeCredible()
}

func (n *Node) SetDead() {
	if n.nodeLocalInfo.nodeState == NodeDead {
		return
	}
	n.ChangeState(NodeDead)
	n.IncreaseVersionTo(n.GetInfo().Version + 1)
	n.nodeLocalInfo.credibility.Store(0)
}
