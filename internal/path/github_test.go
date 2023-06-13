package path

import (
	"fmt"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestGithubPath_Join(t *testing.T) {

	const sha = "76b46f2ecc5c896217f5cb5a0bfdf3346365050b"

	path := NewGitHub("johnhoman", "nop", WithGitHubRepoRef(sha))
	path = path.Join("foo")
	s := path.String()
	qt.Assert(t, s, qt.Equals, "https://github.com/johnhoman/nop.git/foo?ref=76b46f2ecc5c896217f5cb5a0bfdf3346365050b")

	path = NewGitHub("johnhoman", "nop", WithGitHubRepoRef(sha))
	qt.Assert(t, path.Join("..").String(), qt.Equals, "https://github.com/johnhoman/nop.git?ref=76b46f2ecc5c896217f5cb5a0bfdf3346365050b")

	path = NewGitHub("johnhoman", "nop", WithGitHubRepoRef(sha))
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
			path := NewGitHub("johnhoman", "nop", WithGitHubRepoRef(sha))
			ok, err := path.Join(subtest.path).Exists()
			qt.Assert(t, err, qt.IsNil)
			qt.Assert(t, ok, qt.Equals, subtest.want)
		})
	}
}
