package reader

import (
	"bytes"
	"io"
	"net"

	"github.com/ciiim/syncmember/codec"
)

type AReader struct {
	conn          net.Conn
	contentLength int
	offset        int
	coder         codec.Coder
	head          [codec.AHeadLength]byte
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
func (r *AReader) Read(b []byte) (n int, err error) {
	if r.head[0] == 0 {
		_, err = r.conn.Read(r.head[:])
		if err != nil {
			return 0, err
		}
		r.contentLength = r.coder.GetBodyLength(r.head[:])
	}

	n, err = r.conn.Read(b)
	if err != nil {
		return n, err
	}
	r.offset += n
	if r.offset >= r.contentLength {
		r.offset = 0
		r.contentLength = 0
		r.head[0] = 0
		return n, io.EOF
	}
	return n, nil
}

func ReadTCPMessage(conn net.Conn, buf *bytes.Buffer, coder codec.Coder) error {
	header := make([]byte, codec.HeaderLength)
	_, err := io.ReadFull(conn, header)
	if err != nil {
		return err
	}
	if err := coder.ValidateHeader(header); err != nil {
		return err
	}

	//read body
	bodyLen := coder.GetBodyLength(header)
	b := make([]byte, bodyLen)
	_, err = io.ReadFull(conn, b)
	if err != nil {
		return err
	}
	buf.Write(b)
	return nil
}
