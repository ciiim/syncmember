package codec

import (
	"github.com/vmihailenco/msgpack/v5"
)

func Marshal(v any) ([]byte, error) {
	return msgpack.Marshal(v)
}

func Unmarshal(data []byte, v any) error {
	return msgpack.Unmarshal(data, v)
}
