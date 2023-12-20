package reader

import (
	"bytes"
	"net"

	"github.com/ciiim/syncmember/codec"
)

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
