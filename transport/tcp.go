package transport

import (
	"bytes"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type TCPConfig struct {
	ListenAddr string

	TCPTimeout time.Duration

	Logger *slog.Logger
}

type TCPTransport struct {
	config *TCPConfig

	listener net.Listener

	logger *slog.Logger

	connHandler func(net.Conn)

	stopVar *atomic.Bool
}

func NewTCPTransport(conf *TCPConfig, stopVar *atomic.Bool, handler func(net.Conn)) *TCPTransport {
	return &TCPTransport{
		config:      conf,
		logger:      conf.Logger,
		stopVar:     stopVar,
		connHandler: handler,
	}
}

func (t *TCPTransport) Listen(wg *sync.WaitGroup) {
	l, err := net.Listen("tcp", t.config.ListenAddr)
	if err != nil {
		t.logger.Error("TCPTransport listen error", err)
	}
	t.listener = l
	t.logger.Info("TCPTransport listening", "addr", t.config.ListenAddr)
	wg.Done()
	for {
		if err = t.accept(); err != nil {
			t.logger.Error("TCPTransport listen error", err)
		}
	}
}

// func (t *TCPTransport) SendRawBytes(conn net.Conn, to string, buf *bytes.Buffer) error {
// 	if t.stopVar.Load() {
// 		return fmt.Errorf("TCPTransport is stopped")
// 	}
// }

func (t *TCPTransport) DialAndSendRawBytes(to string, buf *bytes.Buffer) (net.Conn, error) {
	if t.stopVar.Load() {
		return nil, fmt.Errorf("TCPTransport is stopped")
	}
	conn, err := net.Dial("tcp", to)
	if err != nil {
		return nil, err
	}
	if err = conn.SetDeadline(time.Now().Add(t.config.TCPTimeout)); err != nil {
		return nil, err
	}
	sent, err := conn.Write(buf.Bytes())
	if err != nil {
		return nil, err
	}
	if sent != buf.Len() {
		return nil, fmt.Errorf("TCPTransport error: sent %d bytes, expected %d", sent, buf.Len())
	}
	return conn, nil
}

func (t *TCPTransport) accept() error {
	conn, err := t.listener.Accept()
	if err != nil {
		if t.stopVar.Load() {
			t.stopAll()
			return nil
		}
		return err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			t.logger.Error("TCPTransport conn close error", err)
		}
	}()
	t.connHandler(conn)
	return nil
}

func (t *TCPTransport) stopAll() {
	t.listener.Close()
	t.logger.Info("TCPTransport stoped")
}
