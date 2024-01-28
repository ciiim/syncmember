package transport

import (
	"bytes"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
)

type Packet struct {
	From   *net.UDPAddr
	Buffer bytes.Buffer
}

type UDPConfig struct {
	HostAddr   *net.UDPAddr
	ListenAddr *net.UDPAddr

	Logger *slog.Logger

	UDPBuffer int

	PacketHandler func(*Packet)
}

type UDPTransport struct {
	conn *net.UDPConn

	logger *slog.Logger

	config *UDPConfig

	packetCh chan *Packet

	stopVar *atomic.Bool

	packetPool *sync.Pool
}

func NewUDPTransport(config *UDPConfig, stop *atomic.Bool) *UDPTransport {
	if config.PacketHandler == nil {
		return nil
	}
	return &UDPTransport{
		config:   config,
		packetCh: make(chan *Packet, 1024),
		stopVar:  stop,
		logger: func() *slog.Logger {
			if config.Logger == nil {
				return slog.Default()
			}
			return config.Logger
		}(),
		packetPool: &sync.Pool{
			New: func() interface{} {
				return &Packet{
					Buffer: bytes.Buffer{},
				}
			},
		},
	}
}

func (u *UDPTransport) GetPacket() *Packet {
	return u.packetPool.Get().(*Packet)
}

func (u *UDPTransport) PutPacket(p *Packet) {
	p.Buffer.Reset()
	u.packetPool.Put(p)
}

func (u *UDPTransport) packet() <-chan *Packet {
	return u.packetCh
}

func ResolveUDPAddr(addr string) (*net.UDPAddr, error) {
	return net.ResolveUDPAddr("udp", addr)
}

func listenUDP(addr *net.UDPAddr) (*net.UDPConn, error) {
	return net.ListenUDP("udp", addr)
}

func (u *UDPTransport) buildPacket(from *net.UDPAddr, buffer []byte) *Packet {
	packet := u.GetPacket()
	packet.From = from
	packet.Buffer.Write(buffer)
	return packet
}

func (u *UDPTransport) Listen() {
	var err error
	u.conn, err = listenUDP(u.config.ListenAddr)
	if err != nil {
		u.logger.Error("UDPTransport listen error:", err)
		return
	}
	u.logger.Info("UDPTransport listening", "listen addr", u.config.ListenAddr)
	for {
		buf := make([]byte, u.config.UDPBuffer)
		n, addr, err := u.conn.ReadFromUDP(buf)
		if u.stopVar.Load() {
			u.stopAll()
			return
		}
		if err != nil {
			u.logger.Error("UDPTransport read error:", err)
			continue
		}
		packet := u.buildPacket(addr, buf[:n])
		if len(u.packetCh) == cap(u.packetCh) {
			u.logger.Warn("UDPTransport packetCh full, packet dropped")
			continue
		}
		u.packetCh <- packet
	}
}

// Handle 接受packetCh中的packet，调用PacketHandler处理
func (u *UDPTransport) Handle() {
	for {
		packet, ok := <-u.packet()
		if !ok {
			u.logger.Debug("UDPTransport packet channel closed")
			return
		}
		if u.stopVar.Load() {
			u.PutPacket(packet)
			u.stopAll()
			return
		}
		u.config.PacketHandler(packet)
		u.PutPacket(packet)
	}
}

func (u *UDPTransport) stopAll() {
	u.conn.Close()
	close(u.packetCh)
	u.logger.Info("UDPTransport stoped")
}

func (u *UDPTransport) SendRaw(buf *bytes.Buffer, to *net.UDPAddr) error {
	n, err := u.conn.WriteToUDP(buf.Bytes(), to)
	if n != buf.Len() {
		errmsg := fmt.Errorf("[WARN] sent %d bytes, expected %d bytes", n, buf.Len())
		u.logger.Warn("UDPTransport WARN ", errmsg)
		return errmsg
	}
	return err
}
