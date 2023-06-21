package path

import (
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
)

func NewMemory() Memory {
	return Memory{}
}

type Memory map[string]any

func (m Memory) toString(root string, segments ...string) string {
	return m.join(root, segments...)
}

func (m Memory) ReadFile(path string) ([]byte, error) {
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return nil, os.ErrNotExist
	}
	zero, path := parts[0], filepath.Join(parts[1:]...)
	if len(parts) == 1 {
		f, ok := m[zero]
		if !ok {
			return nil, os.ErrNotExist
		}
		if data, ok := f.([]byte); ok {
			return data, nil
		}
		if _, ok := f.(Memory); ok {
			return nil, errors.Wrapf(os.ErrInvalid, "%q is a directory", path)
		}
		panic("BUG: this should never be any other type")
	}

	if sub, ok := m[zero]; ok {
		return sub.(Memory).ReadFile(path)
	}
	return nil, os.ErrNotExist
}

func (m Memory) IsDir(path string) (bool, error) {
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return false, os.ErrNotExist
	}
	zero, path := parts[0], filepath.Join(parts[1:]...)
	if len(parts) == 1 {
		_, ok := m[zero]
		if ok {
			switch m[zero].(type) {
			case Memory, map[string]any:
				return true, nil
			case []byte:
				return false, nil
			default:
				panic("BUG: there should be no other types")
			}
		}
		return false, os.ErrNotExist
	}
	dir, ok := m[zero]
	if ok {
		return dir.(Memory).IsDir(path)
	}
	return false, os.ErrNotExist
}

func (m Memory) WriteFile(path string, data []byte) error {
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		panic("BUG: the content should have been written by now")
	}

	zero, path := parts[0], filepath.Join(parts[1:]...)
	if len(parts) == 1 {
		m[zero] = data
		return nil
	}

	sub, ok := m[zero].(Memory)
	if !ok {
		sub = make(Memory)
		m[zero] = sub
	}

	return sub.WriteFile(path, data)
}

func (m Memory) join(root string, segments ...string) string {
	return filepath.Join(root, filepath.Join(segments...))
}
