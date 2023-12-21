package syncmember

import (
	"fmt"
	"log"
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
	config *Config

	nodeName string

	host Address

	logger *slog.Logger

	pingTicker     *time.Ticker
	pushPullTicker *time.Ticker
	gossipTicker   *time.Ticker

	udpTransport *transport.UDPTransport
	tcpTransport *transport.TCPTransport

	nMutex   *sync.Mutex
	me       *Node
	nodes    []*Node
	nodesMap map[string]*Node // ip:port -> *Node
	//等待Pong的节点
	waitPongMap map[string]*Node //ip:port -> *Node

	boardcastQueue *BoardcastQueue

	//副本
	//存储用户数据
	kvcopyTree *btree.BTree
	kvTreeMu   *sync.RWMutex

	messageHandlers map[MessageType]PacketHandlerFunc

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
		boardcastQueue: NewBoardcastQueue(),

		kvTreeMu: new(sync.RWMutex),

		messageHandlers: make(map[MessageType]PacketHandlerFunc),
	}
	err := s.init(config)
	if err != nil {
		log.Fatalln(err)
		// return nil
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
		PacketHandler: s.PacketHandler,
	}

	tcpConfig := transport.TCPConfig{
		Logger:     s.logger,
		ListenAddr: listenAddr.String(),
		TCPTimeout: s.config.TCPTimeout,
	}

	s.host = s.host.WithName(s.nodeName)
	s.me = NewNode(s.host, nil)
	s.me.SetAlive()

	s.udpTransport = transport.NewUDPTransport(&udpConfig, s.stopVar)
	s.tcpTransport = transport.NewTCPTransport(&tcpConfig, s.stopVar, s.handlepushPull)

	s.RegisterMessageHandler(Ping, s.handlePing)
	s.RegisterMessageHandler(Pong, s.handlePong)

	s.RegisterMessageHandler(Alive, s.handleGossip)
	s.RegisterMessageHandler(Dead, s.handleGossip)
	s.RegisterMessageHandler(KVSet, s.handleGossip)
	s.RegisterMessageHandler(KVDelete, s.handleGossip)
	s.RegisterMessageHandler(KVUpdate, s.handleGossip)

	//UDP service
	go s.udpTransport.Handle()
	go s.udpTransport.Listen()

	//TCP service
	go s.tcpTransport.Listen()

	// go monitor.ReportMemoryUsagePer(time.Second * 10)

	time.Sleep(100 * time.Millisecond) //FIXME: wait for udp service start

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

	node := NewNode(ResolveAddr(addr), nil)

	if node.address.Port == s.config.AdvertisePort && node.address.IP.Equal(s.config.AdvertiseIP) {
		s.logger.Error("Join", "failed", "can't join self")
		return fmt.Errorf("can't join self")
	}

	remote, err := s.pushPullNode(node, true)
	if err != nil {
		s.logger.Error("Join", "failed", err)
		return err
	}
	s.logger.Debug("Join to", "remote", remote)
	s.MergeNodes(remote)
	return nil
}

func (s *SyncMember) JoinDebug(addr string) error {
	s.logger.Info("JoinDebug", "addr", addr)

	node := NewNode(ResolveAddr(addr), nil)
	node.SetAlive()
	s.AddNode(node)
	return nil
}

func (s *SyncMember) RegisterMessageHandler(msgType MessageType, handler PacketHandlerFunc) {
	if handler == nil {
		panic("Nil RegisterMessageHandler handler")
	}
	s.messageHandlers[msgType] = handler
}

func (s *SyncMember) PacketHandler(p *transport.Packet) {
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

func (s *SyncMember) Shutdown() {
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
