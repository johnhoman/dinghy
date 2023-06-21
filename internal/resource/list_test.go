package resource

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestList_Insert(t *testing.T) {
	obj := &unstructured.Unstructured{}
	obj.SetName("foo")
	obj.SetNamespace("bar")
	obj.SetGroupVersionKind(schema.GroupVersionKind{Version: "v1", Kind: "Pod"})
	n := Unstructured(obj.UnstructuredContent())

	l := NewList()
	qt.Assert(t, l.Insert(n), qt.IsNil)
	qt.Assert(t, l.objs, qt.HasLen, 1)
}
