package main

import (
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
func (cmd *cmdBuild) Run(stdout io.Writer) error {
	// cmd.Dir could be relative to the current working directory, so it
	// may need to be joined with the working directory
	dir, err := path.Parse(cmd.Dir)
	if err != nil {
		return nil
	}

	tree, err := build.New().Build(dir)
	if err != nil {
		return err
	}
	return resource.PrintTree(tree, stdout)
}
