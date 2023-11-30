package syncmember

import (
	"bytes"
	"fmt"
	"net"

	"github.com/ciiim/syncmember/codec"
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

func readTCPMessage(conn net.Conn, buf *bytes.Buffer) error {
	header := make([]byte, codec.HeaderLength)
	_, err := conn.Read(header)
	if err != nil {
		return err
	}
	if err := codec.ValidateHeader(header); err != nil {
		return err
	}

	//read body
	bodyLen := codec.GetBodyLength(header)
	b := make([]byte, bodyLen)
	_, err = conn.Read(b)
	if err != nil {
		return err
	}
	buf.Write(b)
	return nil
}

// TCP
// 处理pushPull请求
// 读取远程节点的数据；推送本地节点的数据
// 不要在这里关闭连接
func (s *SyncMember) handlepushPull(conn net.Conn) {
	s.logger.Debug("handlepushPull", "remote addr", conn.RemoteAddr().String())

	//PULL OR GET
	buf := bytes.NewBuffer(nil)
	readTCPMessage(conn, buf)

	//unmarshal body
	var remote []NodeInfo
	if err := codec.TCPUnmarshal(buf.Bytes(), &remote); err != nil {
		s.logger.Error("handlepushPull", "unmarshal error", err)
		return
	}

	if err := s.MergeNodes(remote); err != nil {
		s.logger.Error("handlepushPull", "merge nodes error", err)
		return
	}

	//PUSH or PUT
	nodeinfos := make([]NodeInfo, len(s.nodes)+1)
	for i, n := range s.nodes {
		nodeinfos[i] = n.GetInfo()
	}
	nodeinfos[len(s.nodes)] = s.me.GetInfo()
	bufBytes, err := codec.TCPMarshal(nodeinfos)
	if err != nil {
		s.logger.Error("handlepushPull", "marshal error", err)
		return
	}
	_, err = conn.Write(bufBytes)
	if err != nil {
		s.logger.Error("handlepushPull", "write error", err)
		return
	}
}

func (s *SyncMember) pushPullNode(node *Node, join bool) (remote []NodeInfo, err error) {
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
func (s *SyncMember) pushPullNodeInternal(node *Node, nodes []*Node) (remote []NodeInfo, err error) {
	s.logger.Debug("pushPullNode", "target node", node.Addr())

	//PUSH or PUT
	nodeinfos := make([]NodeInfo, len(nodes))
	for i, n := range nodes {
		nodeinfos[i] = n.GetInfo()
	}
	bufBytes, err := codec.TCPMarshal(nodeinfos)
	if err != nil {
		return
	}
	buf := bytes.NewBuffer(bufBytes)
	conn, err := s.tcpTransport.DialAndSendRawBytes(node.Addr().String(), buf)
	if err != nil {
		return
	}

	//PULL or GET
	buf.Reset()
	readTCPMessage(conn, buf)
	remote = make([]NodeInfo, 0)
	if err = codec.TCPUnmarshal(buf.Bytes(), &remote); err != nil {
		return
	}
	return
}

func (s *SyncMember) MergeNodes(remote []NodeInfo) error {
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
