package codec

import (
	"fmt"
	"math"
)

/*
Begin 2Bytes
ContentLength 2Bytes
Body ContentLength Bytes
*/
const (

	//消息开始边界
	Begin uint16 = 0xfc0f

	//消息开始大小
	BeginBytes int = 2

	//内容长度
	LengthBytes int = 2

	HeaderLength int = BeginBytes + LengthBytes
)

var AACoder ACoder

type ACoder struct {
}

func (a ACoder) Encode(body []byte) ([]byte, error) {
	return a.buildMessage(body)
}

func (a ACoder) buildMessage(body []byte) ([]byte, error) {
	bodyLen := len(body)
	if bodyLen > math.MaxInt16 {
		return nil, fmt.Errorf("body length %d is too long", bodyLen)
	}
	msg := make([]byte, len(body)+HeaderLength)
	bodyLen16 := uint16(bodyLen)
	begin := Begin

	//[0:2) begin
	beginSlice := msg[0:2]
	beginSlice[0] = byte(begin >> 8)
	beginSlice[1] = byte(uint8(begin))

	//[2:4) content length
	contentLengthSlice := msg[2:4]
	contentLengthSlice[0] = byte(bodyLen16 >> 8)
	contentLengthSlice[1] = byte(uint8(bodyLen16))

	//[4:len) body
	copy(msg[4:], body)

	return msg, nil
}

func (a ACoder) ValidateHeader(header []byte) error {
	if len(header) != HeaderLength {
		return fmt.Errorf("invalid header length %d", len(header))
	}
	//Check begin
	begin := uint16(header[0])<<8 | uint16(header[1])
	if begin != Begin {
		return fmt.Errorf("invalid begin %d", begin)
	}
	return nil
}

func (a ACoder) GetBodyLength(header []byte) int {
	if len(header) != HeaderLength {
		return 0
	}
	return int(uint16(header[2])<<8 | uint16(header[3]))
}

func (a ACoder) GetExtra(header []byte) any {
	return nil
}
