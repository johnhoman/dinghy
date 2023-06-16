package mutate

import (
	"github.com/johnhoman/dinghy/internal/visitor"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Annotations map[string]string

func (m *Annotations) Visit(obj *unstructured.Unstructured) error {
	return visitor.AddAnnotations(*m).Visit(obj)
}

type Namespace struct {
	Name string `yaml:"name"`
}

func (n *Namespace) Visit(obj *unstructured.Unstructured) error {
	obj.SetNamespace(n.Name)
	return nil
}

type Metadata map[string]any

func (m *Metadata) Visit(obj *unstructured.Unstructured) error {
	return visitor.StrategicMergePatch(*m).Visit(obj)
}
