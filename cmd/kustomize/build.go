package main

import (
	"github.com/johnhoman/kustomize"
	"github.com/pkg/errors"
)

type cmdBuild struct {
	Dir string `kong:"name=dir,arg"`
}

// Run builds the kustomization package and emits the resources
// to stdout
func (cmd *cmdBuild) Run() error {
	// the path might not be a file system path
	dir := kustomize.NewOsPath(cmd.Dir)
	switch kustomize.ParseReferenceType(cmd.Dir) {
	case kustomize.ReferenceTypeRemoteGitHub:
	}

	c := &kustomize.Config{}
	path := dir.Join("kustomization.yaml")
	if err := kustomize.NewReader(path).UnmarshalYAML(c); err != nil {
		return errors.Wrap(err, "an error occurred read the kustomization.yaml")
	}
	r := kustomize.NewRenderer(c, dir)
	return errors.Wrap(r.Print(), "failed to render manifests")
}
