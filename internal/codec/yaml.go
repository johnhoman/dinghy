package codec

import (
	"bytes"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io"
	"strings"
)

var (
	_ Decoder = &yamlDecoder{}
)

func YAMLCopyTo(to, from any) error {
	buf := new(bytes.Buffer)
	if err := YAMLEncoder(buf).Encode(from); err != nil {
		return err
	}
	return YAMLDecoder(buf).Decode(to)
}

// YAMLDecoder parses only known YAML and checks for required fields.
// Required fields are required when they are marked with the
// tag `dinghy:"required"`.
func YAMLDecoder(r io.Reader) Decoder {
	d := yaml.NewDecoder(r)
	// d.KnownFields(true)
	return &yamlDecoder{d: d}
}

func YAMLEncoder(w io.Writer) Encoder {
	e := yaml.NewEncoder(w)
	return e
}

type yamlDecoder struct{ d *yaml.Decoder }

func (y *yamlDecoder) Decode(obj any) error {
	if err := y.d.Decode(obj); err != nil {
		return err
	}
	if v, ok := obj.(Validator); ok {
		if resp := v.Validate(); len(resp) > 0 {
			return errors.Errorf("decoding yaml failed with the following validation errors: \n\t%s", strings.Join(resp, "\n\t"))
		}
	}
	return nil
}
