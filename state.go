package syncmember

type NodeStateType uint8

const (
	NodeDead NodeStateType = iota
	NodeAlive
	NodeTimeout
)

func (n *Node) SetAlive() {
	if n.nodeLocalInfo.nodeState == NodeAlive {
		return
	}
	n.ChangeState(NodeAlive)
	n.IncreaseVersionTo(1)
}
