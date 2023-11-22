package syncmember

type MessageType int8

const (
	HeartBeat MessageType = iota
	HeartBeatAck
	Join

	Alive
	Dead
)

type Message struct {
	MsgType MessageType
	From    Address
	To      Address
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
	Message
}

func NewHeartBeatMessage(from, to Address) *HeartBeatMessage {
	return &HeartBeatMessage{
		Message: Message{
			MsgType: HeartBeat,
			From:    from,
			To:      to,
		},
	}
}

type HeartBeatAckMessage struct {
	Message
}

func NewHeartBeatAckMessage(from, to Address) *HeartBeatAckMessage {
	return &HeartBeatAckMessage{
		Message: Message{
			MsgType: HeartBeatAck,
			From:    from,
			To:      to,
		},
	}
}
