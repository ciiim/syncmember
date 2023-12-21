package syncmember

import (
	"bytes"

	"github.com/ciiim/syncmember/codec"
)

type MessageType int8

func (m MessageType) String() string {
	switch m {
	case Ping:
		return "Ping"
	case Pong:
		return "Pong"
	case Alive:
		return "Alive"
	case Dead:
		return "Dead"
	case KVSet:
		return "KVSet"
	case KVDelete:
		return "KVDelete"
	case KVUpdate:
		return "KVUpdate"
	default:
		return "Unknown"
	}
}

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
	Seq     uint64 //FIXME: 暂时不起作用
	Payload []byte
}

func NewMessage(msgType MessageType, payload []byte) *Message {

	return &Message{
		MsgType: msgType,
		Payload: payload,
	}
}

type NodeInfoPayload struct {
	Addr      Address
	NodeState NodeStateType
	Version   int64
}

func (p *NodeInfoPayload) Encode() *bytes.Buffer {
	b, err := codec.Marshal(p)
	if err != nil {
		return nil
	}
	return bytes.NewBuffer(b)
}

func (p *NodeInfoPayload) Decode(b []byte) error {
	return codec.Unmarshal(b, p)
}

type KeyValuePayload struct {
	Key   string
	Value []byte
}

func (p *KeyValuePayload) Encode() *bytes.Buffer {
	b, err := codec.Marshal(p)
	if err != nil {
		return nil
	}
	return bytes.NewBuffer(b)
}

func (p *KeyValuePayload) Decode(b []byte) error {
	return codec.Unmarshal(b, p)
}

func (m *Message) GetPayload() []byte {
	return m.Payload
}

func NewPingMessage() *Message {
	return &Message{
		MsgType: Ping,
		Seq:     randSeq(),
		Payload: nil,
	}
}

func NewPongMessage(seq uint64) *Message {
	return &Message{
		MsgType: Pong,
		Seq:     seq + 1,

		Payload: nil,
	}
}

func NewAliveMessage(payload []byte) *Message {
	return &Message{
		MsgType: Alive,
		Payload: payload,
	}
}

func NewDeadMessage(payload []byte) *Message {
	return &Message{
		MsgType: Dead,
		Payload: payload,
	}
}
