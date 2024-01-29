package syncmember

import (
	"fmt"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ciiim/syncmember/codec"
	"github.com/ciiim/syncmember/transport"
	"github.com/google/btree"
)

type SyncMember struct {
	// hot fields
	nodesMap map[string]*Node // ip:port -> *Node

	nMutex *sync.Mutex
	me     *Node
	nodes  []*Node
	//等待Pong的节点
	waitPongMap map[string]*Node //ip:port -> *Node

	config *Config

	nodeName string

	host Address

	logger *slog.Logger

	pingTicker     *time.Ticker
	pushPullTicker *time.Ticker
	gossipTicker   *time.Ticker

	udpTransport *transport.UDPTransport
	tcpTransport *transport.TCPTransport

	boardcastQueue *BoardcastQueue

	//副本
	//存储用户数据
	kvcopyTree *btree.BTree
	kvTreeMu   *sync.RWMutex

	messageHandlers map[MessageType]PacketHandlerFunc

	nodeEvent NodeEventDelegate

	kWatcher *kVWatcher

	stopCh  chan struct{}
	stopVar *atomic.Bool
}

func NewSyncMember(nodeName string, config *Config) *SyncMember {
	s := &SyncMember{
		config:   config,
		nodeName: nodeName,
		stopCh:   make(chan struct{}, 2),
		stopVar:  new(atomic.Bool),

		nMutex:         new(sync.Mutex),
		nodes:          make([]*Node, 0),
		nodesMap:       make(map[string]*Node),
		waitPongMap:    make(map[string]*Node),
		boardcastQueue: newBoardcastQueue(),

		kvTreeMu: new(sync.RWMutex),

		messageHandlers: make(map[MessageType]PacketHandlerFunc),

		kWatcher: newKVWatcher(),
	}
	err := s.init(config)
	if err != nil {
		return nil
	}
	return s
}

func (s *SyncMember) init(config *Config) error {
	if err := s.readConfig(config); err != nil {
		return err
	}
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", s.config.AdvertiseIP.String(), s.config.AdvertisePort))
	if err != nil {
		return err
	}
	listenAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", s.config.BindIP.String(), s.config.BindPort))
	if err != nil {
		return err
	}
	udpConfig := transport.UDPConfig{
		HostAddr:      udpAddr,
		ListenAddr:    listenAddr,
		Logger:        s.logger,
		UDPBuffer:     s.config.UDPBufferSize,
		PacketHandler: s.packetHandler,
	}

	tcpConfig := transport.TCPConfig{
		Logger:     s.logger,
		ListenAddr: listenAddr.String(),
		TCPTimeout: s.config.TCPTimeout,
	}

	s.host = s.host.withName(s.nodeName)
	s.me = newNode(s.host, nil)
	s.me.setAlive()

	s.udpTransport = transport.NewUDPTransport(&udpConfig, s.stopVar)
	s.tcpTransport = transport.NewTCPTransport(&tcpConfig, s.stopVar, s.handlepushPull)

	s.registerMessageHandler(Ping, s.handlePing)
	s.registerMessageHandler(Pong, s.handlePong)

	s.registerMessageHandler(Alive, s.handleGossip)
	s.registerMessageHandler(Dead, s.handleGossip)
	s.registerMessageHandler(KVSet, s.handleGossip)
	s.registerMessageHandler(KVDelete, s.handleGossip)
	s.registerMessageHandler(KVUpdate, s.handleGossip)

	wg := new(sync.WaitGroup)
	wg.Add(2)

	//UDP service
	go s.udpTransport.Handle()
	go s.udpTransport.Listen(wg)

	//TCP service
	go s.tcpTransport.Listen(wg)

	wg.Wait()

	return nil
}

func (s *SyncMember) Run() error {

	go s.ping()
	go s.pushPull()
	go s.gossip()

	s.waitShutdown()

	s.stop()
	return nil
}

func (s *SyncMember) Join(addr string) error {
	s.logger.Info("Join in", "member", addr)

	node := newNode(resolveAddr(addr), nil)

	if node.address.Port == s.config.AdvertisePort && node.address.IP.Equal(s.config.AdvertiseIP) {
		s.logger.Error("Join", "failed", "can't join self")
		return fmt.Errorf("can't join self")
	}

	remote, err := s.pushPullNode(node, true)
	if err != nil {
		s.logger.Error("Push Pull Node", "failed", err)
		return err
	}
	err = s.MergeNodes(remote)
	if err != nil {
		s.logger.Error("MergeNodes", "failed", err)
		return err
	}
	return nil
}

// func (s *SyncMember) joinDebug(addr string) error {
// 	s.logger.Info("JoinDebug", "addr", addr)

// 	node := newNode(resolveAddr(addr), nil)
// 	node.SetAlive()
// 	s.AddNode(node)
// 	return nil
// }

func (s *SyncMember) registerMessageHandler(msgType MessageType, handler PacketHandlerFunc) {
	if handler == nil {
		panic("Nil registerMessageHandler handler")
	}
	if s.messageHandlers == nil {
		s.messageHandlers = make(map[MessageType]PacketHandlerFunc)
	}
	s.messageHandlers[msgType] = handler
}

func (s *SyncMember) packetHandler(p *transport.Packet) {
	start := time.Now()
	var packet Packet
	err := codec.Unmarshal(p.Buffer.Bytes(), &packet)
	if err != nil {
		s.logger.Error("UDPUnmarshal error", "error", err)
		return
	}
	s.logger.Debug("handle packet", "packet message", packet.MessageBody.MsgType)
	handler, ok := s.messageHandlers[packet.MessageBody.MsgType]
	if !ok {
		s.logger.Error("no handler for packet", "message type", packet.MessageBody.MsgType, "from", packet.From)
		return
	}
	handler(&packet)
	s.logger.Debug("handle packet done", "packet message type", packet.MessageBody.MsgType, "cost(ms)", float64(time.Since(start).Microseconds())/1000.0)
}

func (s *SyncMember) Node() Address {
	return s.me.address
}

func (s *SyncMember) Shutdown() {
	if s.stopVar.Load() {
		s.logger.Warn("Already shutdown")
		return
	}
	s.logger.Info("Shutdown...")
	s.stopCh <- struct{}{}
}

func (s *SyncMember) waitShutdown() {
	<-s.stopCh
}

func (s *SyncMember) stop() {

	s.stopVar.Store(true)
	s.pingTicker.Stop()
	s.pushPullTicker.Stop()
	s.gossipTicker.Stop()
}
