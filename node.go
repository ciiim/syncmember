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
	Addr      Address
	NodeState NodeStateType
	Version   int64
	TTL       int8
}

type Node struct {
	address       Address
	nodeLocalInfo NodeLocalInfo
}

func (s *SyncMember) AddNode(node *Node) {
	s.nMutex.Lock()
	defer s.nMutex.Unlock()
	if _, ok := s.nodesMap[node.Addr().String()]; ok {
		s.logger.Warn("node already exist", "node", node.Addr())
		return
	}
	s.nodes = append(s.nodes, node)
	s.nodesMap[node.Addr().String()] = node
}

func NewNode(addr Address, nodeInfo *NodeInfo) *Node {
	n := &Node{
		address: addr,
		nodeLocalInfo: NodeLocalInfo{
			nodeState:   NodeUnknown, //初始默认未知
			version:     atomic.Int64{},
			credibility: atomic.Int32{},
		},
	}
	if nodeInfo == nil {
		return n
	} else {
		n = &Node{
			address: addr,
			nodeLocalInfo: NodeLocalInfo{
				nodeState:   nodeInfo.NodeState,
				version:     atomic.Int64{},
				credibility: atomic.Int32{},
			},
		}
		n.nodeLocalInfo.version.Store(nodeInfo.Version)
	}
	return n
}

// 改变节点状态，重置节点可信度，增加版本号
func (n *Node) SetAlive() {
	if n.nodeLocalInfo.nodeState == NodeAlive {
		return
	}
	n.ChangeState(NodeAlive)
	n.IncreaseVersionTo(n.GetInfo().Version + 1)
	n.BecomeCredible()
}

// 改变节点状态，重置节点可信度，增加版本号
func (n *Node) SetDead() {
	if n.nodeLocalInfo.nodeState == NodeDead {
		return
	}
	n.ChangeState(NodeDead)
	n.IncreaseVersionTo(n.GetInfo().Version + 1)
	n.BecomeUnCredible()
}

func (n *Node) BecomeUnCredible() {
	n.nodeLocalInfo.credibility.Store(0)
}

func (n *Node) BecomeCredible() {
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
		Addr:      n.address,
		NodeState: n.nodeLocalInfo.nodeState,
		Version:   n.nodeLocalInfo.version.Load(),
	}
}

func (ni *NodeInfo) SetTTL(ttl int8) {
	ni.TTL = ttl
}
