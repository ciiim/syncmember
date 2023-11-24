package syncmember

import (
	"sync/atomic"
)

type NodeLocalInfo struct {
	nodeState   NodeStateType
	version     atomic.Int64
	credibility atomic.Int32
}

type NodeInfo struct {
	NodeState NodeStateType
	Version   int64
}

type Node struct {
	address       Address
	nodeLocalInfo NodeLocalInfo
}

func (s *SyncMember) AddNode(node *Node) {
	s.nMutex.Lock()
	defer s.nMutex.Unlock()
	s.nodes = append(s.nodes, node)
	s.nodesMap[node.Addr().String()] = node
}

func NewNode(addr Address) *Node {
	n := &Node{
		address: addr,
		nodeLocalInfo: NodeLocalInfo{
			nodeState:   NodeUnknown, //初始默认未知
			version:     atomic.Int64{},
			credibility: atomic.Int32{},
		},
	}
	n.nodeLocalInfo.version.Store(0)
	n.BecomeCredible()
	return n
}

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
	n.BecomeUnCredible()
}

func (n *Node) BecomeUnCredible() {
	if n.nodeLocalInfo.nodeState == NodeDead {
		return
	}
	n.nodeLocalInfo.credibility.Store(0)
}

func (n *Node) BecomeCredible() {
	if n.nodeLocalInfo.nodeState == NodeDead {
		return
	}
	n.nodeLocalInfo.credibility.Store(3)
}

func (n *Node) IsCredible() bool {
	return n.nodeLocalInfo.credibility.Load() > 0
}

func (n *Node) IncreaseVersionTo(d int64) bool {
	if d < n.nodeLocalInfo.version.Load() {
		return false
	}
	n.nodeLocalInfo.version.Store(d)
	return true
}

func (n *Node) NodeState() NodeStateType {
	return n.nodeLocalInfo.nodeState
}

func (n *Node) ChangeState(newState NodeStateType) {
	n.nodeLocalInfo.nodeState = newState
}

func (n *Node) Addr() Address {
	return n.address
}

func (n *Node) GetInfo() NodeInfo {
	return NodeInfo{
		NodeState: n.nodeLocalInfo.nodeState,
		Version:   n.nodeLocalInfo.version.Load(),
	}
}
