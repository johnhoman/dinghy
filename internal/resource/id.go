package resource

import (
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"strings"
)

func newResourceKey(obj *unstructured.Unstructured) resourceKey {
	key := resourceKey{
		kind:       obj.GroupVersionKind().Kind,
		apiVersion: obj.GroupVersionKind().GroupVersion().String(),
		name:       obj.GetName(),
		namespace:  obj.GetNamespace(),
	}
	return key
}

// resourceKey is a unique identifier for a resource.
type resourceKey struct {
	apiVersion string
	kind       string
	name       string
	namespace  string
}

func (r resourceKey) String() string {
	s := strings.ReplaceAll(r.apiVersion, "/", ".")
	s = fmt.Sprintf("%s.%s", s, r.kind)
	if r.namespace != "" {
		s = s + "/" + r.namespace
	}
	return s + "/" + r.name
}

func (r resourceKey) GetName() string {
	return r.name
}

func (r resourceKey) GetNamespace() string {
	return r.namespace
}

func (r resourceKey) Kind() string {
	return r.kind
}

func (r resourceKey) GroupVersion() string {
	return r.apiVersion
}
