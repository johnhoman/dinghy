package resource

import (
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"strings"
)

func newResourceKey(obj *unstructured.Unstructured) Key {
	gvk := obj.GroupVersionKind()
	name := obj.GetName()
	key := Key{
		Kind:         gvk.Kind,
		GroupVersion: obj.GroupVersionKind().GroupVersion().String(),
		Name:         name,
		Namespace:    obj.GetNamespace(),
	}
	return key
}

// Key is a unique identifier for a resource.
type Key struct {
	GroupVersion string
	Kind         string
	Name         string
	Namespace    string
}

func (r Key) String() string {
	s := strings.ReplaceAll(r.GroupVersion, "/", ".")
	s = fmt.Sprintf("%s.%s", s, r.Kind)
	if r.Namespace != "" {
		s = s + "/" + r.Namespace
	}
	return s + "/" + r.Name
}

func Wrap(m map[string]any) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: m}
}
