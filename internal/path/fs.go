package path

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

func NewFSPath(fs afero.Fs, cur string) Path {
	return &fsPath{fs: fs, cur: cur}
}

type fsPath struct {
	fs  afero.Fs
	cur string
}

func (f *fsPath) init() error {
	return nil
}

func (f *fsPath) Copy(r io.Reader) error {
	fp, err := f.Create()
	if err != nil {
		return err
	}
	_, err = io.Copy(fp, r)
	return err
}

func (f *fsPath) WriteBytes(b []byte) error {
	fp, err := f.Create()
	if err != nil {
		return err
	}
	_, err = fp.Write(b)
	return err
}

func (f *fsPath) WriteJSON(obj any) error {
	fp, err := f.Create()
	if err != nil {
		return err
	}
	return json.NewEncoder(fp).Encode(obj)
}

func (f *fsPath) WriteYAML(obj any) error {
	fp, err := f.Create()
	if err != nil {
		return err
	}
	return yaml.NewEncoder(fp).Encode(obj)
}

func (f *fsPath) WriteString(content string) error {
	w, err := f.Create()
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, content)
	return err
}

// Create opens a new file (or truncates an existing one) for writing and
// creates any intermediate directories fieldPath the path.
func (f *fsPath) Create() (io.Writer, error) {

	parent := f.Join("..").(*fsPath)
	if err := f.fs.MkdirAll(parent.cur, 0700); err != nil {
		return nil, err
	}
	fp, err := f.fs.Create(f.cur)
	if err != nil {
		return nil, err
	}
	return fp, nil
}

func (f *fsPath) IsDir() (bool, error) {
	ok, err := f.Exists()
	if err != nil {
		return false, err
	}
	if !ok {
		return false, os.ErrNotExist
	}

	info, err := f.fs.Stat(f.cur)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

func (f *fsPath) Exists() (bool, error) {
	fs := afero.Afero{Fs: f.fs}
	if fs.Fs == nil {
		fs.Fs = afero.NewOsFs()
	}
	return fs.Exists(f.cur)
}

func (f *fsPath) String() string { return f.cur }
func (f *fsPath) Join(path ...string) Path {
	fs := f.fs
	if fs == nil {
		fs = afero.NewOsFs()
	}
	joined := filepath.Join(path...)
	if filepath.IsAbs(joined) {
		return NewFSPath(f.fs, joined)
	}
	return NewFSPath(f.fs, filepath.Join(f.cur, joined))
}

func (f *fsPath) Open() (io.Reader, error) {
	fs := f.fs
	if fs == nil {
		fs = afero.NewOsFs()
	}
	fp, err := fs.Open(f.cur)
	if err != nil {
		return nil, err
	}
	return fp, nil
}
