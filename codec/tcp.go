package codec

import (
	"fmt"
	"math"

	"github.com/vmihailenco/msgpack/v5"
)

/*
Begin 2Bytes
ContentLength 2Bytes
ExtraFlag 1Byte
Body ContentLength Bytes
End 2Bytes
*/
const (
	Begin          uint16 = 0xfc0f
	End            uint16 = Begin
	BeginBytes     int    = 2
	ExtraFlagBytes int    = 1
	LengthBytes    int    = 2
	HeaderLength   int    = BeginBytes + LengthBytes + ExtraFlagBytes
)

func TCPMarshal(v any) ([]byte, error) {
	body, err := msgpack.Marshal(v)
	if err != nil {
		return nil, err
	}
	return buildMessage(body)

}

func buildMessage(body []byte) ([]byte, error) {
	bodyLen := len(body)
	if bodyLen > math.MaxInt16 {
		return nil, fmt.Errorf("body length %d is too long", bodyLen)
	}
	msg := make([]byte, len(body)+HeaderLength)
	bodyLen16 := uint16(bodyLen)
	begin := Begin

	//[0:1] begin
	msg[0] = byte(begin >> 8)
	msg[1] = byte(uint8(begin))

	//[2:3] content length
	contentLengthSlice := msg[2:4]
	contentLengthSlice[0] = byte(bodyLen16 >> 8)
	contentLengthSlice[1] = byte(uint8(bodyLen16))

	//[4] extra flag
	msg[4] = 0

	//[5:len-3] body
	copy(msg[5:], body)

	return msg, nil
}

func TCPUnmarshal(data []byte, v any) error {
	return msgpack.Unmarshal(data, v)
}

func ValidateHeader(header []byte) error {
	if len(header) != HeaderLength {
		return fmt.Errorf("invalid header length %d", len(header))
	}
	begin := uint16(header[0])<<8 | uint16(header[1])
	if begin != Begin {
		return fmt.Errorf("invalid begin %d", begin)
	}
	return nil
}

func GetBodyLength(header []byte) int {
	if len(header) != HeaderLength {
		return 0
	}
	return int(uint16(header[2])<<8 | uint16(header[3]))
}
