package generate

import (
	"github.com/johnhoman/dinghy/internal/resource"
)

type Generator interface {
	Emit() (resource.Tree, error)
}

type Func func() (resource.Tree, error)

func (f Func) Emit() (resource.Tree, error) {
	return f()
}
