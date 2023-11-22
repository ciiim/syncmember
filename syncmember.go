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

	host Address

	logger *slog.Logger

	heartBeatTicker *time.Ticker
	pullPushTicker  *time.Ticker

	udpTransport *transport.UDPTransport

	nMutex   *sync.Mutex
	me       *Node
	nodes    []*Node
	nodesMap map[string]*Node // key: ip:port

	messageHandlers map[MessageType]func(*Message)
	ackMap          map[string]*Node //key ip:port

	stopCh  chan struct{}
	stopVar *atomic.Bool
}

func NewSyncMember(config *Config) *SyncMember {
	s := &SyncMember{
		config:  config,
		stopCh:  make(chan struct{}, 2),
		stopVar: new(atomic.Bool),

		me:       NewNode(ResolveAddr(fmt.Sprintf("%s:%d", config.AdvertiseIP.String(), config.AdvertisePort))),
		nodes:    make([]*Node, 0),
		nodesMap: make(map[string]*Node),

		nMutex: new(sync.Mutex),
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

	s.udpTransport = transport.NewUDPTransport(&udpConfig, s.stopVar)
	return nil
}

func (s *SyncMember) Join(addr string) error {
	s.logger.Info("Join", "addr", addr)

	node := NewNode(ResolveAddr(addr))
	s.AddNode(node)
	return nil
}

func (s *SyncMember) PacketHandler(p *transport.Packet) {
	var msg Message
	err := codec.UDPUnmarshal(p.Buffer.Bytes(), &msg)
	if err != nil {
		s.logger.Error("UDPUnmarshal error", "error", err)
		return
	}
	s.logger.Debug("handle packet", "packet message", msg)
	switch msg.MsgType {
	case HeartBeat:
		s.handleHeartbeat(&msg)
	default:
		s.logger.Error("unknown message type", "type", msg.MsgType)
		return
	}
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
