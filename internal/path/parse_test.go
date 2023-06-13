package path

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestParse_GitHub(t *testing.T) {
	cases := []string{
		"https://github.com/john-homan/nop.git",
		"https://github.com/john-homan/nop.git/path/to/file",
		"https://github.com/john-homan/nop.git/path/to/file?ref=asdfasdfasdf",
		"https://github.com/john.homan/nop.git",
		"https://github.com/john.homan/nop.git/path/to/file",
		"https://github.com/john.homan/nop.git/path/to/file?ref=asdfasdfasdf",
		"https://github.com/johnhoman/nop.git",
		"https://github.com/johnhoman/nop.git/path/to/file",
		"https://github.com/johnhoman/nop.git/path/to/file?ref=asdfasdfasdf",
		"https://github.com/johnhoman/nop",
		"https://github.com/johnhoman/nop/path/to/file",
		"https://github.com/johnhoman/nop/path/to/file?ref=asdfasdfasdf",
		"github.com/johnhoman/nop",
		"github.com/johnhoman/nop/path/to/file",
		"github.com/johnhoman/nop/path/to/file?ref=asdfasdfasdf",
		"github.com/johnhoman/nop.git",
		"github.com/johnhoman/nop.git/path/to/file",
		"github.com/johnhoman/nop.git/path/to/file?ref=asdfasdfasdf",
	}

	for _, u := range cases {
		t.Run(u, func(t *testing.T) {
			path, err := Parse(u)
			qt.Assert(t, err, qt.IsNil)
			gh, ok := path.(*github)
			qt.Assert(t, ok, qt.IsTrue)
			qt.Assert(t, gh, qt.IsNotNil)
		})
	}
}
