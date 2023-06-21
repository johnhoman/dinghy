package path

import (
	"io"
)

func NewPath(path impl, root string) Path {
	return Path{
		path: path,
		root: root,
	}
}

type Path struct {
	path impl
	root string
}

func (bp Path) ReadFile(path ...string) ([]byte, error) {
	return bp.path.ReadFile(bp.path.join(bp.root, path...))
}

func (bp Path) IsDir(path ...string) (bool, error) {
	return bp.path.IsDir(bp.path.join(bp.root, path...))
}

func (bp Path) Join(segments ...string) Path {
	return Path{
		path: bp.path,
		root: bp.path.join(bp.root, segments...),
	}
}

func (bp Path) ReadText(path ...string) (string, error) {
	return ReadText(bp.path, bp.path.join(bp.root, path...))
}

func (bp Path) ReadBytes(path ...string) ([]byte, error) {
	return ReadBytes(bp.path, bp.path.join(bp.root, path...))
}

func (bp Path) Exists(path ...string) (bool, error) {
	return Exists(bp.path, bp.path.join(bp.root, path...))
}

func (bp Path) Reader(path ...string) (io.Reader, error) {
	return Reader(bp.path, bp.path.join(bp.root, path...))
}

func (bp Path) String(path ...string) string {
	return bp.path.toString(bp.root, path...)
}

func (bp Path) Relative() bool {
	return IsRelative(bp.root)
}
