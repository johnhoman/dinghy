package path

import (
	"os"
	"path/filepath"
)

type Local struct{}

func (l Local) toString(root string, segments ...string) string {
	return filepath.Join(root, filepath.Join(segments...))
}

func NewLocal() Local {
	return Local{}
}

func (l Local) IsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.Mode().IsDir(), nil
}

func (l Local) ReadFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (l Local) join(root string, segments ...string) string {
	return filepath.Join(root, filepath.Join(segments...))
}
