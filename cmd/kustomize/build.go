package main

import (
	"os"

	"github.com/pkg/errors"

	"github.com/johnhoman/kustomize"
)

type cmdBuild struct {
	Dir string `kong:"name=dir,arg"`

	Namespace  string `kong:"name=namespace,help=Set the output namespace"`
	NamePrefix string `kong:"help=set a name prefix on all emitted resources"`
	NameSuffix string `kong:"help=set a name suffix on all emitted resources"`
}

// Run builds the kustomization package and emits the resources
// to stdout
func (cmd *cmdBuild) Run() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	dir, err := kustomize.Factory(kustomize.NewOsPath(wd), cmd.Dir)
	if err != nil {
		return err
	}

	c := &kustomize.Config{}
	path := dir.Join("kustomization.yaml")
	if err := kustomize.NewReader(path).UnmarshalYAML(c); err != nil {
		return errors.Wrap(err, "an error occurred read the kustomization.yaml")
	}
	if cmd.Namespace != "" {
		c.Mutations = append(c.Mutations, kustomize.MutationSpec{
			Selector: kustomize.ResourceSelector{},
			Uses:     "builtin.kustomize.k8s.io/replaceNamespace",
			With: map[string]any{
				"namespace": cmd.Namespace,
			},
		})
	}
	if cmd.NamePrefix != "" {
		c.Mutations = append(c.Mutations, kustomize.MutationSpec{
			Selector: kustomize.ResourceSelector{},
			Uses:     "builtin.kustomize.k8s.io/prependNamePrefix",
			With: map[string]any{
				"namePrefix": cmd.NamePrefix,
			},
		})
	}
	if cmd.NameSuffix != "" {
		c.Mutations = append(c.Mutations, kustomize.MutationSpec{
			Selector: kustomize.ResourceSelector{},
			Uses:     "builtin.kustomize.k8s.io/appendNameSuffix",
			With: map[string]any{
				"nameSuffix": cmd.NameSuffix,
			},
		})
	}
	return errors.Wrap(kustomize.NewRenderer(c, dir).Print(), "failed to render manifests")
}
