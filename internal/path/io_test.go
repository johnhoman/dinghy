package path

import (
	qt "github.com/frankban/quicktest"
	"github.com/johnhoman/dinghy/internal/codec"
	"github.com/spf13/afero"
	"testing"
)

func TestReader_UnmarshalYAML(t *testing.T) {
	type Foo struct {
		Bar string `yaml:"bar" bespoke:"required"`
	}

	t.Run("UnsetRequiredFieldsReturnsError", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		writeYAMLToFS(t, fs, "/foo/bar", `{}`)

		path := NewFSPath(fs, "/foo/bar")
		r := NewReader(path)
		qt.Assert(t, r.UnmarshalYAML(&Foo{}), qt.ErrorIs, codec.ErrRequiredField("bar"))
	})

	t.Run("ParsesValidYAML", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		writeYAMLToFS(t, fs, "/foo/bar", `{"bar": "foo/baz"}`)

		path := NewFSPath(fs, "/foo/bar")
		r := NewReader(path)
		qt.Assert(t, r.UnmarshalYAML(&Foo{}), qt.IsNil)
	})

}

func writeYAMLToFS(t *testing.T, fs afero.Fs, path string, doc string) {
	f, err := fs.Create(path)
	qt.Assert(t, err, qt.IsNil)
	ret, err := f.WriteString(doc)
	qt.Assert(t, err, qt.IsNil)
	qt.Assert(t, ret, qt.Equals, len(doc))
}
