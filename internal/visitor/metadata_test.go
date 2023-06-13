package visitor

import (
	qt "github.com/frankban/quicktest"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"testing"
)

func TestAnnotations(t *testing.T) {
	add := map[string]string{
		"app.kubernetes.io/part-of": "testing",
		"app.kubernetes.io/name":    "annotations",
	}
	obj := &unstructured.Unstructured{Object: make(map[string]any)}
	visitor := AddAnnotations(add)
	qt.Assert(t, visitor.Visit(obj), qt.IsNil)
	qt.Assert(t, obj.GetAnnotations(), qt.DeepEquals, add)
}

func TestAnnotation(t *testing.T) {

}
