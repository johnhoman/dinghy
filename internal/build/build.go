package build

import (
	"gopkg.in/yaml.v3"
	"io"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/johnhoman/dinghy/internal/path"
	"github.com/johnhoman/dinghy/internal/resource"
)

var (
	_ Builder = &dinghy{}
	_ Builder = &kustomize{}
)

type Builder interface {
	Build(path path.Path, opts ...Option) (resource.Tree, error)
}

func New() Builder {
	return &dinghy{}
}

func NewKustomize() Builder {
	return &kustomize{}
}

func buildResource(r string, root path.Path, tree resource.Tree, newBuilderFunc func() Builder) error {
	target := root.Join(r)
	if !path.IsRelative(r) {
		parsed, err := path.Parse(r)
		if err != nil {
			return err
		}
		target = parsed
	}

	isDir, err := target.IsDir()
	if err != nil {
		return err
	}
	if isDir {
		b := newBuilderFunc()
		sub, err := b.Build(target)
		if err != nil {
			return err
		}
		return resource.CopyTree(tree, sub)
	}

	f, err := target.Reader()
	if err != nil {
		return err
	}
	dec := yaml.NewDecoder(f)
	for {
		var m map[string]any
		if err := dec.Decode(&m); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if m == nil {
			continue
		}
		obj := &unstructured.Unstructured{Object: m}
		if err := tree.Insert(obj); err != nil {
			return err
		}
	}
	return nil
}
