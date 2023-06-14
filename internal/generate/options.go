package generate

import (
	"github.com/johnhoman/dinghy/internal/path"
)

type Option func(o *Options)

func WithDirectoryRoot(root path.Path) Option {
	return func(o *Options) {
		o.Root = root
	}
}

type Options struct {
	Root path.Path
}
