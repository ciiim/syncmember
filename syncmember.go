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
)

type SyncMember struct {
	config *Config

	nodeName string

	host Address

	logger *slog.Logger

	heartBeatTicker *time.Ticker
	pullPushTicker  *time.Ticker

	udpTransport *transport.UDPTransport

	nMutex     *sync.Mutex
	me         *Node
	nodes      []*Node
	nodesMap   map[string]*Node // key: ip:port
	waitAckMap map[string]*Node //key ip:port -> *Node

	messageHandlers map[MessageType]func(*Message)

	stopCh  chan struct{}
	stopVar *atomic.Bool
}

func NewSyncMember(nodeName string, config *Config) *SyncMember {
	s := &SyncMember{
		config:   config,
		nodeName: nodeName,
		stopCh:   make(chan struct{}, 2),
		stopVar:  new(atomic.Bool),

		nMutex:     new(sync.Mutex),
		nodes:      make([]*Node, 0),
		nodesMap:   make(map[string]*Node),
		waitAckMap: make(map[string]*Node),

		messageHandlers: make(map[MessageType]func(*Message)),
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

	s.host = s.host.WithName(s.nodeName)
	s.me = NewNode(s.host)

	s.udpTransport = transport.NewUDPTransport(&udpConfig, s.stopVar)

	s.RegisterMessageHandler(HeartBeat, s.handleHeartbeat)
	s.RegisterMessageHandler(HeartBeatAck, s.handleAckHeartbeat)

	return nil
}

func (s *SyncMember) Join(addr string) error {
	s.logger.Info("Join", "addr", addr)

	node := NewNode(ResolveAddr(addr))
	s.AddNode(node)
	return nil
}

func (s *SyncMember) JoinDebug(addr string) error {
	s.logger.Info("JoinDebug", "addr", addr)

	node := NewNode(ResolveAddr(addr))
	node.SetAlive()
	s.AddNode(node)
	return nil
}

func (s *SyncMember) RegisterMessageHandler(msgType MessageType, handler func(*Message)) {
	if handler == nil {
		s.logger.Error("RegisterMessageHandler handler is nil")
		panic("RegisterMessageHandler handler is nil")
	}
	s.messageHandlers[msgType] = handler
}

func (s *SyncMember) PacketHandler(p *transport.Packet) {
	start := time.Now()
	var msg Message
	err := codec.UDPUnmarshal(p.Buffer.Bytes(), &msg)
	if err != nil {
		s.logger.Error("UDPUnmarshal error", "error", err)
		return
	}
	s.logger.Debug("handle packet", "packet message", msg)
	handler, ok := s.messageHandlers[msg.MsgType]
	if !ok {
		s.logger.Error("no handler for message", "message type", msg.MsgType, "from", msg.From)
		return
	}
	handler(&msg)
	s.logger.Debug("handle packet done", "packet message type", msg.MsgType, "cost(ms)", float64(time.Since(start).Microseconds())/1000.0)
}

func (s *SyncMember) Run() error {

	s.me.SetAlive()
	//UDP service
	go s.udpTransport.UDPHandler()
	go s.udpTransport.UDPListen()

	//HeartBeat
	go s.heartBeat()

	s.waitShutdown()

	s.stop()
	return nil
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
	s.heartBeatTicker.Stop()
	s.pullPushTicker.Stop()
}
