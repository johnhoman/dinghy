package path

import (
	"github.com/johnhoman/dinghy/internal/codec"
)

func NewReader(path Path) Reader {
	return &reader{path: path}
}

type reader struct {
	path Path
}

func (r *reader) UnmarshalYAML(obj any) error {
	f, err := r.path.Open()
	if err != nil {
		return err
	}
	return codec.YAMLDecoder(f).Decode(obj)
}
