package syncmember

import (
	"bytes"

	"github.com/ciiim/syncmember/codec"
)

type MessageType int8

const (
	HeartBeat MessageType = iota
	HeartBeatAck

	Alive
	Dead

	KVSet
	KVDelete
	KVUpdate
)

type Message struct {
	MsgType MessageType
	From    Address
	To      Address
	Payload *bytes.Buffer
}

type KeyValuePayload struct {
	Key   string
	Value []byte
}

func (p *KeyValuePayload) Encode() *bytes.Buffer {
	b, err := codec.UDPMarshal(p)
	if err != nil {
		return nil
	}
	return bytes.NewBuffer(b)
}

func (p *KeyValuePayload) Decode(b []byte) error {
	return codec.UDPUnmarshal(b, p)
}

func (m *Message) GetPayload() *bytes.Buffer {
	return nil
}

func NewHeartBeatMessage(from, to Address) *Message {
	return &Message{
		MsgType: HeartBeat,
		From:    from,
		To:      to,
		Payload: nil,
	}
}

func NewHeartBeatAckMessage(from, to Address) *Message {
	return &Message{
		MsgType: HeartBeatAck,
		From:    from,
		To:      to,
		Payload: nil,
	}
}
