package mutate

import (
	"github.com/johnhoman/dinghy/internal/fieldpath"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Patch struct {
	FieldPaths []string `yaml:"fieldPaths"`
	Value      any      `yaml:"value"`
}

func (p *Patch) Visit(obj *unstructured.Unstructured) error {
	m := obj.UnstructuredContent()
	for _, fieldPath := range p.FieldPaths {
		fp, err := fieldpath.Parse(fieldPath)
		if err != nil {
			return err
		}
		if err := fp.SetValue(m, p.Value); err != nil {
			return err
		}
	}
	return nil
}
