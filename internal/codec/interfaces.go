package codec

type Decoder interface {
	Decode(obj any) error
}

type Encoder interface {
	Encode(obj any) error
}
