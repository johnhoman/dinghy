package generate

import (
	"github.com/johnhoman/dinghy/internal/resource"
)

var (
	_ Generator = Func(nil)
)

type Generator interface {
	Emit(config any, opts ...Option) (resource.Tree, error)
}

type Func func(config any, opts ...Option) (resource.Tree, error)

func (f Func) Emit(config any, opts ...Option) (resource.Tree, error) {
	return f(config, opts...)
}
