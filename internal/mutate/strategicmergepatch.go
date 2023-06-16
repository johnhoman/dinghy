package mutate

import (
	"github.com/johnhoman/dinghy/internal/visitor"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type StrategicMergePatch map[string]any

func (s *StrategicMergePatch) Visit(obj *unstructured.Unstructured) error {
	return visitor.StrategicMergePatch(*s).Visit(obj)
}
