package syncmember

type NodeStateType uint8

const (
	NodeUnknown NodeStateType = iota
	NodeDead
	NodeAlive
)

// 由远程节点发起的状态变更触发
// 也可以由心跳判断的状态变更触发
func alive(n *Node, m IMessage) {
	n.SetAlive()
}

// 由远程节点发起的状态变更触发
// 也可以由心跳判断的状态变更触发
func dead(n *Node, m IMessage) {
	n.SetDead()
}
