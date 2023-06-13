package path

import (
	"encoding/json"
	qt "github.com/frankban/quicktest"
	"github.com/spf13/afero"
	"testing"
)

func TestFsPath_Join(t *testing.T) {
	path := NewFSPath(afero.NewMemMapFs(), "/")
	qt.Assert(t, path.Join("foo").String(), qt.Equals, "/foo")

	path = NewFSPath(afero.NewMemMapFs(), "/foo")
	qt.Assert(t, path.Join("..").String(), qt.Equals, "/")

	path = NewFSPath(afero.NewMemMapFs(), "/foo")
	qt.Assert(t, path.Join("../..").String(), qt.Equals, "/")
}

func TestFsPath_Open(t *testing.T) {
	fs := afero.NewMemMapFs()
	f, err := fs.Create("/foo/bar.json")
	qt.Assert(t, err, qt.IsNil)
	want := map[string]string{"foo": "bar"}
	qt.Assert(t, json.NewEncoder(f).Encode(want), qt.IsNil)

	path := NewFSPath(fs, "/foo")
	fp, err := path.Join("bar.json").Open()
	qt.Assert(t, err, qt.IsNil)
	var got map[string]string
	qt.Assert(t, json.NewDecoder(fp).Decode(&got), qt.IsNil)
	qt.Assert(t, got, qt.DeepEquals, want)
}
