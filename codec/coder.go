package codec

type Coder interface {
	Encode(data []byte) ([]byte, error)
	ValidateHeader(header []byte) error
	GetBodyLength(header []byte) int
	GetExtra(header []byte) any
}
