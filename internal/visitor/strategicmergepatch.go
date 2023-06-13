package visitor

import (
	"github.com/johnhoman/dinghy/internal/scheme"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

// StrategicMergePatch is a visitor that performs a strategic merge patch
// on a resource with the provided patch
func StrategicMergePatch(patch map[string]any) Visitor {
	return Func(func(obj *unstructured.Unstructured) error {
		d, err := scheme.Scheme.New(obj.GroupVersionKind())
		if err != nil {
			d = &unstructured.Unstructured{}
		}

		o := obj.UnstructuredContent()
		o, err = strategicpatch.StrategicMergeMapPatch(o, patch, d)
		if err != nil {
			return err
		}
		obj.SetUnstructuredContent(o)
		return nil
	})
}
