package syncmember

import (
	"bytes"

	"github.com/ciiim/syncmember/codec"
)

type MessageType int8

const (
	Ping MessageType = iota
	Pong

	Alive
	Dead

	KVSet
	KVDelete
	KVUpdate
)

type Message struct {
	MsgType MessageType
	Seq     uint64
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

func NewPingMessage(from, to Address) *Message {
	return &Message{
		MsgType: Ping,
		Seq:     randSeq(),
		From:    from,
		To:      to,
		Payload: nil,
	}
}

func NewPongMessage(from, to Address, seq uint64) *Message {
	return &Message{
		MsgType: Pong,
		Seq:     seq + 1,
		From:    from,
		To:      to,
		Payload: nil,
	}
}

func NewAliveMessage(from, to Address) *Message {
	return &Message{
		MsgType: Alive,
		From:    from,
		To:      to,
		Payload: nil,
	}
}

func NewDeadMessage(from, to Address) *Message {
	return &Message{
		MsgType: Dead,
		From:    from,
		To:      to,
		Payload: nil,
	}
}
