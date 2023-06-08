package kustomize

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/afero"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestFsPath_Join(t *testing.T) {
	path := NewPath(afero.NewMemMapFs(), "/")
	qt.Assert(t, path.Join("foo").String(), qt.Equals, "/foo")

	path = NewPath(afero.NewMemMapFs(), "/foo")
	qt.Assert(t, path.Join("..").String(), qt.Equals, "/")

	path = NewPath(afero.NewMemMapFs(), "/foo")
	qt.Assert(t, path.Join("../..").String(), qt.Equals, "/")
}

func TestFsPath_Open(t *testing.T) {
	fs := afero.NewMemMapFs()
	f, err := fs.Create("/foo/bar.json")
	qt.Assert(t, err, qt.IsNil)
	want := map[string]string{"foo": "bar"}
	qt.Assert(t, json.NewEncoder(f).Encode(want), qt.IsNil)

	path := NewPath(fs, "/foo")
	fp, err := path.Join("bar.json").Open()
	qt.Assert(t, err, qt.IsNil)
	var got map[string]string
	qt.Assert(t, json.NewDecoder(fp).Decode(&got), qt.IsNil)
	qt.Assert(t, got, qt.DeepEquals, want)
}

func TestGithubPath_Join(t *testing.T) {

	const sha = "76b46f2ecc5c896217f5cb5a0bfdf3346365050b"

	repo := NewGitHub("").Repo("johnhoman", "nop")
	path := NewPathGitHub(repo, "", sha)
	path = path.Join("foo")
	s := path.String()
	qt.Assert(t, s, qt.Equals, "https://github.com/johnhoman/nop.git/foo?ref=76b46f2ecc5c896217f5cb5a0bfdf3346365050b")

	path = NewPathGitHub(repo, "/foo", sha)
	qt.Assert(t, path.Join("..").String(), qt.Equals, "https://github.com/johnhoman/nop.git?ref=76b46f2ecc5c896217f5cb5a0bfdf3346365050b")

	path = NewPathGitHub(repo, "/foo", sha)
	qt.Assert(t, path.Join("../..").String(), qt.Equals, "https://github.com/johnhoman/nop.git?ref=76b46f2ecc5c896217f5cb5a0bfdf3346365050b")
}

func TestGithubPath_Exists(t *testing.T) {

	cases := map[string]struct {
		path string
		want bool
	}{
		"RootExists": {
			path: "",
			want: true,
		},
		"NotExists": {
			path: "foo",
			want: false,
		},
		"DirectoryExists": {
			path: "1",
			want: true,
		},
		"FileExists": {
			path: "1/kustomization.yaml",
			want: true,
		},
	}

	const sha = "76b46f2ecc5c896217f5cb5a0bfdf3346365050b"

	for name, subtest := range cases {
		t.Run(fmt.Sprintf("%s: %s", name, subtest.path), func(t *testing.T) {
			repo := NewGitHub("").Repo("johnhoman", "nop")
			path := NewPathGitHub(repo, "", sha)
			ok, err := path.Join(subtest.path).Exists()
			qt.Assert(t, err, qt.IsNil)
			qt.Assert(t, ok, qt.Equals, subtest.want)
		})
	}
}
