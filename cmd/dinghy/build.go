package main

import (
	"github.com/spf13/afero"
	"io"

	"github.com/johnhoman/dinghy/internal/build"
	"github.com/johnhoman/dinghy/internal/path"
	"github.com/johnhoman/dinghy/internal/resource"
)

type cmdBuild struct {
	Dir string `kong:"name=dir,arg"`
}

// Run builds the kustomization package and emits the resources
// to stdout
func (cmd *cmdBuild) Run(fs afero.Fs, wd path.Path, stdout io.Writer) error {
	// cmd.Dir could be relative to the current working directory, so it
	// may need to be joined with the working directory

	p, err := path.Parse(cmd.Dir, path.WithRelativeRoot(wd))
	if err != nil {
		return err
	}

	b := build.New()
	tree, err := b.Build(p)
	if err != nil {
		return err
	}
	return resource.PrintTree(tree, stdout)
}
