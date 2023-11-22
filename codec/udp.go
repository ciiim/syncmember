package codec

import (
	"github.com/vmihailenco/msgpack/v5"
)

func UDPMarshal(v any) ([]byte, error) {
	return msgpack.Marshal(v)
}

func UDPUnmarshal(data []byte, v any) error {
	return msgpack.Unmarshal(data, v)
}
