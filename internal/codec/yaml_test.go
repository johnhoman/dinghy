package codec

import (
	qt "github.com/frankban/quicktest"
	"strings"
	"testing"
)

func TestYamlDecoder_DecodeErrMissingRequiredField(t *testing.T) {

	type Bar struct {
		Name string `yaml:"name" dinghy:"required"`
		Port int    `yaml:"port"`
	}

	type Foo struct {
		Bar   int   `yaml:"bar" dinghy:"required"`
		Baz   int   `yaml:"baz" dinghy:"required"`
		Foo   *Foo  `yaml:"foo"`
		Items []Bar `yaml:"items"`
	}

	cases := map[string]struct {
		name string
	}{
		`{"bar": 2}`: {
			name: "baz",
		},
		`{"bar": 2, "baz": 1, "foo": {"bar": 10}}`: {
			name: "foo.baz",
		},
		`{"bar": 2, "baz": 1, "foo": {"bar": 10, "baz": 5}, "items": [{"port": 80}]}`: {
			name: "items[0].name",
		},
	}
	for doc, subtest := range cases {
		t.Run(doc, func(t *testing.T) {
			d := YAMLDecoder(strings.NewReader(doc))
			qt.Assert(t, d.Decode(&Foo{}), qt.ErrorIs, ErrRequiredField(subtest.name))
		})
	}
}
