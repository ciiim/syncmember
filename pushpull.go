package syncmember

import (
	"bytes"
	"fmt"
	"net"

	"github.com/ciiim/syncmember/codec"
	"github.com/ciiim/syncmember/reader"
)

func (s *SyncMember) pushPull() {
	for {
		select {
		case <-s.pushPullTicker.C:
			s.doPushPull()
		case <-s.stopCh:
			return
		}
	}
}

func (s *SyncMember) doPushPull() {
	target := kRamdonNodes(s.config.PushPullNums, s.nodes, func(n *Node) bool {
		return !n.IsCredible()
	})
	if len(target) == 0 {
		s.logger.Debug("no nodes can pushPull")
		return
	}
	for _, node := range target {
		remoteNodes, err := s.pushPullNodeInternal(node, s.nodes)
		if err != nil {
			s.logger.Error("pushPullNode", "error", err)
			continue
		}
		if err = s.MergeNodes(remoteNodes); err != nil {
			s.logger.Error("MergeNode", "error", err)
			continue
		}
	}
}

// TCP
// 处理pushPull请求
// 读取远程节点的数据；推送本地节点的数据
// 不要在这里关闭连接
func (s *SyncMember) handlepushPull(conn net.Conn) {
	s.logger.Debug("handlepushPull", "remote addr", conn.RemoteAddr().String())

	//PULL OR GET
	buf := bytes.NewBuffer(nil)
	if err := reader.ReadTCPMessage(conn, buf, codec.AACoder); err != nil {
		s.logger.Error("handlepushPull", "read error", err)
		return
	}

	//unmarshal body
	var remote []NodeInfoPayload
	if err := codec.Unmarshal(buf.Bytes(), &remote); err != nil {
		s.logger.Error("handlepushPull", "unmarshal error", err)
		return
	}

	if err := s.MergeNodes(remote); err != nil {
		s.logger.Error("handlepushPull", "merge nodes error", err)
		return
	}

	//PUSH or PUT
	nodeinfos := make([]NodeInfoPayload, len(s.nodes)+1)
	for i, n := range s.nodes {
		nodeinfos[i] = n.GetInfo()
	}
	nodeinfos[len(s.nodes)] = s.me.GetInfo()
	bufBytes, err := codec.Marshal(nodeinfos)
	if err != nil {
		s.logger.Error("handlepushPull", "marshal error", err)
		return
	}
	messageBytes, err := codec.AACoder.Encode(bufBytes)
	if err != nil {
		s.logger.Error("handlepushPull", "encode error", err)
		return
	}
	_, err = conn.Write(messageBytes)
	if err != nil {
		s.logger.Error("handlepushPull", "write error", err)
		return
	}
}

func (s *SyncMember) pushPullNode(node *Node, join bool) (remote []NodeInfoPayload, err error) {
	s.nMutex.Lock()
	defer s.nMutex.Unlock()
	if join {
		nodesCopy := make([]*Node, len(s.nodes)+1)
		copy(nodesCopy, s.nodes)
		nodesCopy[len(s.nodes)] = s.me
		return s.pushPullNodeInternal(node, nodesCopy)
	}
	return s.pushPullNodeInternal(node, s.nodes)
}

// TCP
// 发起pushPull请求
// 推送本地节点的数据；读取远程节点的数据
func (s *SyncMember) pushPullNodeInternal(node *Node, nodes []*Node) (remote []NodeInfoPayload, err error) {
	s.logger.Debug("pushPullNode", "target node", node.Addr())

	//PUSH or PUT
	nodeinfos := make([]NodeInfoPayload, len(nodes))
	for i, n := range nodes {
		nodeinfos[i] = n.GetInfo()
	}
	bufBytes, err := codec.Marshal(nodeinfos)
	if err != nil {
		return
	}
	messageBytes, err := codec.AACoder.Encode(bufBytes)
	if err != nil {
		return
	}
	buf := bytes.NewBuffer(messageBytes)
	conn, err := s.tcpTransport.DialAndSendRawBytes(node.Addr().String(), buf)
	if err != nil {
		return
	}

	//PULL or GET
	buf.Reset()
	if err = reader.ReadTCPMessage(conn, buf, codec.AACoder); err != nil {
		return
	}

	remote = make([]NodeInfoPayload, 0)
	if err = codec.Unmarshal(buf.Bytes(), &remote); err != nil {
		return
	}
	return
}

func (s *SyncMember) MergeNodes(remote []NodeInfoPayload) error {
	if len(remote) == 0 {
		return nil
	}
	for _, nodeinfo := range remote {
		switch nodeinfo.NodeState {
		case NodeAlive:
			s.alive(&nodeinfo)
		case NodeDead:
			s.dead(&nodeinfo)
		default:
			return fmt.Errorf("MergeNodes Unknown NodeState %d", nodeinfo.NodeState)
		}
	}
	return nil
}
