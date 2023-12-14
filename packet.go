package syncmember

type PacketHandlerFunc func(packet *Packet)

type Packet struct {
	MessageBody *Message
	From        Address
	To          Address
}

func NewPacket(msg *Message, from, to Address) *Packet {
	return &Packet{
		MessageBody: msg,
		From:        from,
		To:          to,
	}
}
