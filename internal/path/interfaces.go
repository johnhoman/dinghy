package path

import (
	"fmt"
	"io"
	"net/http"
)

var (
	_ Path   = &fsPath{}
	_ Path   = &github{}
	_ Reader = &reader{}
	_ Writer = &fsPath{}
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Path interface {
	fmt.Stringer
	Join(path ...string) Path
	Open() (io.Reader, error)
	IsDir() (bool, error)
	Exists() (bool, error)

	init() error
}

type Reader interface {
	UnmarshalYAML(obj any) error
}

type Writer interface {
	WriteJSON(obj any) error
	WriteYAML(obj any) error
	Create() (io.Writer, error)
	WriteString(content string) error
	Copy(r io.Reader) error
	WriteBytes(b []byte) error
}
