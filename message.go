package syncmember

import (
	"bytes"

	"github.com/ciiim/syncmember/codec"
)

type MessageType int8

const (
	HeartBeat MessageType = iota
	HeartBeatAck

	Join

	Alive
	Dead

	KVSet
	KVDelete
	KVUpdate
)

type IMessage interface {
	BMessage() *Message
	Payload() *bytes.Buffer
}

type Message struct {
	MsgType MessageType
	From    Address
	To      Address
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

func (m *Message) BMessage() *Message {
	return m
}

func (m *Message) Payload() *bytes.Buffer {
	return nil
}

type AliveMessage struct {
	Message
	Node NodeInfo
}

type DeadMessage struct {
	Message
	Node NodeInfo
}

type HeartBeatMessage struct {
	*Message
}

func NewHeartBeatMessage(from, to Address) *HeartBeatMessage {
	return &HeartBeatMessage{
		Message: &Message{
			MsgType: HeartBeat,
			From:    from,
			To:      to,
		},
	}
}

func (h *HeartBeatMessage) BMessage() *Message {
	return h.Message
}

func (h *HeartBeatMessage) Payload() *bytes.Buffer {
	return nil
}

type HeartBeatAckMessage struct {
	*Message
}

func NewHeartBeatAckMessage(from, to Address) *HeartBeatAckMessage {
	return &HeartBeatAckMessage{
		Message: &Message{
			MsgType: HeartBeatAck,
			From:    from,
			To:      to,
		},
	}
}

func (a *HeartBeatAckMessage) BMessage() *Message {
	return a.Message
}

func (a *HeartBeatAckMessage) Payload() *bytes.Buffer {
	return nil
}

type JoinMessage struct {
	*Message
}

func NewJoinMessage(from, to Address) *JoinMessage {
	return &JoinMessage{
		Message: &Message{
			MsgType: Join,
			From:    from,
			To:      to,
		},
	}
}

func (j *JoinMessage) BMessage() *Message {
	return j.Message
}

func (j *JoinMessage) Payload() *bytes.Buffer {
	return nil
}
