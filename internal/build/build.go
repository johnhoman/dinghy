package build

import (
	"github.com/johnhoman/dinghy/internal/context"
	"github.com/johnhoman/dinghy/internal/path"
	"github.com/johnhoman/dinghy/internal/resource"
	"github.com/johnhoman/dinghy/internal/types"
)

var (
	_ Builder = &dinghy{}
)

type Builder interface {
	Build(ctx *context.Context, path path.Path, opts ...Option) (resource.Tree, error)
	BuildFromConfig(ctx *context.Context, c *types.Config, opts ...Option) (resource.Tree, error)
}

func New() Builder {
	return &dinghy{}
}
