package mutate

import (
	qt "github.com/frankban/quicktest"
	"github.com/johnhoman/dinghy/internal/resource"
	"testing"
)

func TestAnnotations_Visit_AddsAnnotationsToObject(t *testing.T) {
	c := qt.New(t)

	annotations := Annotations{
		"key1": "value1",
		"key2": "value2",
	}

	obj := resource.Unstructured(map[string]any{
		"apiVersion": "example.com/v1",
		"kind":       "Example",
	})

	err := annotations.Visit(obj)
	c.Assert(err, qt.IsNil)

	expectedAnnotations := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	c.Assert(obj.GetAnnotations(), qt.DeepEquals, expectedAnnotations)
}

func TestAnnotations_Visit_ReturnsErrorIfObjectIsNil(t *testing.T) {
	c := qt.New(t)

	annotations := Annotations{
		"key": "value",
	}

	err := annotations.Visit(nil)
	c.Assert(err, qt.Not(qt.IsNil))
}

func TestAnnotations_Visit_OverwritesExistingAnnotations(t *testing.T) {
	c := qt.New(t)

	annotations := Annotations{
		"key1": "new_value1",
		"key3": "value3",
	}

	obj := resource.Unstructured(map[string]any{
		"apiVersion": "example.com/v1",
		"kind":       "Example",
	})

	obj.SetAnnotations(map[string]string{
		"key1": "value1",
		"key2": "value2",
	})

	err := annotations.Visit(obj)
	c.Assert(err, qt.IsNil)

	expectedAnnotations := map[string]string{
		"key1": "new_value1",
		"key3": "value3",
		"key2": "value2",
	}

	c.Assert(obj.GetAnnotations(), qt.DeepEquals, expectedAnnotations)
}

func TestAnnotations_Visit_ReturnsNilError(t *testing.T) {
	c := qt.New(t)

	annotations := Annotations{
		"key": "value",
	}

	obj := resource.Unstructured(map[string]any{
		"apiVersion": "example.com/v1",
		"kind":       "Example",
	})

	err := annotations.Visit(obj)
	c.Assert(err, qt.IsNil)
}

func TestAnnotations_Visit_SkipsIfEmptyAnnotations(t *testing.T) {
	c := qt.New(t)

	annotations := Annotations{}

	obj := resource.Unstructured(map[string]any{
		"apiVersion": "example.com/v1",
		"kind":       "Example",
	})

	obj.SetAnnotations(map[string]string{
		"key1": "value1",
		"key2": "value2",
	})

	err := annotations.Visit(obj)
	c.Assert(err, qt.IsNil)

	expectedAnnotations := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	c.Assert(obj.GetAnnotations(), qt.DeepEquals, expectedAnnotations)
}

func TestNamespace_Visit_SetsNamespaceOnObject(t *testing.T) {
	c := qt.New(t)

	namespace := &Namespace{
		Namespace: "my-namespace",
	}

	obj := resource.Unstructured(make(map[string]any))

	err := namespace.Visit(obj)
	c.Assert(err, qt.IsNil)

	c.Assert(obj.GetNamespace(), qt.Equals, "my-namespace")
}

func TestNamespace_Visit_ReturnsErrorIfObjectIsNil(t *testing.T) {
	c := qt.New(t)

	namespace := &Namespace{
		Namespace: "my-namespace",
	}

	err := namespace.Visit(nil)
	c.Assert(err, qt.Not(qt.IsNil))
}

func TestNamespace_Visit_SetsEmptyNamespaceIfNameIsEmpty(t *testing.T) {
	c := qt.New(t)

	namespace := &Namespace{
		Namespace: "",
	}

	obj := resource.Unstructured(make(map[string]any))

	err := namespace.Visit(obj)
	c.Assert(err, qt.IsNil)

	c.Assert(obj.GetNamespace(), qt.Equals, "")
}
