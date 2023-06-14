package visitor

import (
	qt "github.com/frankban/quicktest"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"testing"
)

func TestVisitorFunc_Visit(t *testing.T) {
	got := &unstructured.Unstructured{Object: make(map[string]any)}
	visitor := Func(func(obj *unstructured.Unstructured) error {
		obj.SetAnnotations(map[string]string{"foo": "bar"})
		return nil
	})
	qt.Assert(t, visitor.Visit(got), qt.IsNil)
	want := &unstructured.Unstructured{Object: map[string]any{
		"metadata": map[string]any{
			"annotations": map[string]any{
				"foo": "bar",
			},
		},
	}}
	qt.Assert(t, got.Object, qt.DeepEquals, want.Object)
}

func TestChain_Visit(t *testing.T) {
	got := &unstructured.Unstructured{Object: make(map[string]any)}
	want := &unstructured.Unstructured{Object: map[string]any{
		"metadata": map[string]any{
			"annotations": map[string]any{
				"foo":                       "bar",
				"app.kubernetes.io/part-of": "kubernetes",
				"app.kubernetes.io/name":    "name",
			},
		},
	}}
	addAnnotation := func(key, value string) Visitor {
		return Func(func(obj *unstructured.Unstructured) error {
			a := obj.GetAnnotations()
			if a == nil {
				a = make(map[string]string)
			}
			a[key] = value
			obj.SetAnnotations(a)
			return nil
		})
	}
	chain := Chain{
		addAnnotation("foo", "bar"),
		addAnnotation("app.kubernetes.io/part-of", "kubernetes"),
		addAnnotation("app.kubernetes.io/name", "name"),
	}
	qt.Assert(t, chain.Visit(got), qt.IsNil)
	qt.Assert(t, got, qt.DeepEquals, want)
}
