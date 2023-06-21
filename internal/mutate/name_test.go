package mutate

import (
	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/johnhoman/dinghy/internal/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"
)

func TestName_Visit_SetsNamePrefixAndSuffix(t *testing.T) {
	c := qt.New(t)

	name := &Name{
		Prefix: "pre-",
		Suffix: "-suf",
	}

	obj := &resource.Object{
		Unstructured: &unstructured.Unstructured{},
	}
	obj.SetName("example")
	obj.SetLabels(map[string]string{
		"app.kubernetes.io/name": "example",
		"app":                    "example",
	})

	err := name.Visit(obj)
	c.Assert(err, qt.IsNil)

	c.Assert(obj.GetName(), qt.Equals, "pre-example-suf")

	expectedLabels := map[string]string{
		"app.kubernetes.io/name": "pre-example-suf",
		"app":                    "pre-example-suf",
	}
	c.Assert(obj.GetLabels(), qt.DeepEquals, expectedLabels)
}

func TestName_Visit_SkipsIfLabelsNotPresent(t *testing.T) {
	c := qt.New(t)

	name := &Name{
		Prefix: "pre-",
		Suffix: "-suf",
	}

	obj := &resource.Object{
		Unstructured: &unstructured.Unstructured{},
	}
	obj.SetName("example")

	err := name.Visit(obj)
	c.Assert(err, qt.IsNil)

	c.Assert(obj.GetName(), qt.Equals, "pre-example-suf")

	expectedLabels := map[string]string{}
	c.Assert(obj.GetLabels(), qt.CmpEquals(cmpopts.EquateEmpty()), expectedLabels)
}

func TestName_SideEffect_ChangesReferencingResources(t *testing.T) {
	c := qt.New(t)

	name := &Name{
		Prefix: "pre-",
		Suffix: "-suf",
	}

	oldObj := &resource.Object{
		Unstructured: &unstructured.Unstructured{},
	}
	oldObj.SetName("configmap-example")
	oldObj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "ConfigMap",
	})

	tree := resource.NewTree()

	err := name.SideEffect(oldObj, tree)
	c.Assert(err, qt.IsNil)

}

func TestName_SideEffect_SkipsIfNoReferencingResources(t *testing.T) {
	c := qt.New(t)

	name := &Name{
		Prefix: "pre-",
		Suffix: "-suf",
	}

	oldObj := &resource.Object{
		Unstructured: &unstructured.Unstructured{},
	}
	oldObj.SetName("configmap-example")
	oldObj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "ConfigMap",
	})

	tree := resource.NewTree()

	err := name.SideEffect(oldObj, tree)
	c.Assert(err, qt.IsNil)
}
