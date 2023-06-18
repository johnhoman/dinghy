package mutate

import (
	"github.com/johnhoman/dinghy/internal/visitor"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var (
	_ yaml.Unmarshaler = &StrategicMergePatch{}
)

type StrategicMergePatch map[string]any

func (s *StrategicMergePatch) UnmarshalYAML(value *yaml.Node) error {
	var m map[string]any
	if err := value.Decode(&m); err != nil {
		return err
	}
	*s = m
	return nil
}

func (s *StrategicMergePatch) Visit(obj *unstructured.Unstructured) error {
	return visitor.StrategicMergePatch(*s).Visit(obj)
}
