package reader

import (
	"bytes"
	"io"
	"net"

	"github.com/ciiim/syncmember/codec"
)

type AReader struct {
	conn  net.Conn
	coder codec.Coder
	head  [codec.AHeadLength]byte
}

func NewAReader(conn net.Conn) *AReader {
	return &AReader{
		conn:  conn,
		coder: codec.AACoder,
		head:  [codec.AHeadLength]byte{},
	}
}

var _ io.ReadCloser = &AReader{}

func (r *AReader) Close() error {
	return r.conn.Close()
}

// if head is not read, read head first
func (r *AReader) Read(p []byte) (n int, err error) {
	if r.head[0] == 0 {
		_, err = io.ReadFull(r.conn, r.head[:])
		if err != nil {
			return 0, err
		}
	}
	bodyLen := r.coder.GetBodyLength(r.head[:])
	if bodyLen > len(p) {
		return 0, io.ErrShortBuffer
	}
	n, err = io.ReadFull(r.conn, p[:bodyLen])
	if err != nil {
		return 0, err
	}
	r.head[0] = 0
	return n, nil
}

func ReadTCPMessage(conn net.Conn, buf *bytes.Buffer, coder codec.Coder) error {
	header := make([]byte, codec.HeaderLength)
	_, err := conn.Read(header)
	if err != nil {
		return err
	}
	if err := coder.ValidateHeader(header); err != nil {
		return err
	}

	//read body
	bodyLen := coder.GetBodyLength(header)
	b := make([]byte, bodyLen)
	_, err = conn.Read(b)
	if err != nil {
		return err
	}
	buf.Write(b)
	return nil
}
