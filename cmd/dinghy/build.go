package main

import (
	"io"

	"github.com/johnhoman/dinghy/internal/build"
	"github.com/johnhoman/dinghy/internal/context"
	"github.com/johnhoman/dinghy/internal/path"
	"github.com/johnhoman/dinghy/internal/resource"
	"github.com/johnhoman/dinghy/internal/types"
)

type cmdBuild struct {
	Dir       string `kong:"name=dir,arg"`
	Kustomize bool   `kong:"default=false,short=k"`
}

// Run builds the kustomization package and emits the resources
// to stdout
func (cmd *cmdBuild) Run(stdout io.Writer) error {
	// cmd.Dir could be relative to the current working directory, so it
	// may need to be joined with the working directory
	dir, err := path.Parse(cmd.Dir)
	if err != nil {
		return nil
	}

	c := context.NewContext(true)
	b := build.New()
	if cmd.Kustomize {
		tree, err := b.BuildFromConfig(c, &types.Config{
			Generators: []types.GeneratorSpec{{
				Uses: "builtin.dinghy.dev/kustomize",
				With: map[string]any{
					"source": cmd.Dir,
				},
			}},
		})
		if err != nil {
			return err
		}
		return resource.PrintTree(tree, stdout)
	}
	c.SetRoot(cmd.Dir)

	tree, err := b.Build(c, dir)
	if err != nil {
		return err
	}
	return resource.PrintTree(tree, stdout)
}
