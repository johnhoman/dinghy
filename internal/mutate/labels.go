package mutate

import (
	"strings"

	"github.com/johnhoman/dinghy/internal/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Labels mutates the labels on a resource and any selectors that may
// reference that label
type Labels map[string]string

// Visit sets the prefix and suffix on the metadata.name attribute
// of provided object.
func (l *Labels) Visit(obj *unstructured.Unstructured) error {
	labels := obj.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	for key, value := range *l {
		labels[key] = value
	}
	obj.SetLabels(labels)
	return nil
}

// SideEffect searches for resource in the tree that might reference the mutated
// resource by name, such as a deployment referencing a configmap. If a ConfigMap
// name is changed, any Deployments, StatefulSets, pods, DaemonSets, ..., etc that
// reference it will also need to change the reference name
func (l *Labels) SideEffect(old *unstructured.Unstructured, tree resource.Tree) error {
	return nil
}

type deepSet struct {
	to      string
	from    string
	refSpec map[string][]string
}

func (ds *deepSet) setValue(obj map[string]any, path ...string) {
	if len(path) == 0 {
		panic("BUG: this should never get here")
	}
	zero := path[0]
	if len(path) == 1 {
		if obj[zero].(string) == ds.from {
			obj[zero] = ds.to
		}
		return
	}
	switch o := obj[zero].(type) {
	case map[string]any:
		ds.setValue(o, path[1:]...)
	case []any:
		for k := 0; k < len(o); k++ {
			ds.setValue(o[k].(map[string]any), path[1:]...)
		}
	}
}

func (ds *deepSet) Visit(obj *unstructured.Unstructured) error {
	gk := obj.GroupVersionKind().GroupKind().String()
	paths, ok := ds.refSpec[gk]
	if !ok {
		return nil
	}
	o := obj.Object
	for _, path := range paths {
		parts := strings.Split(path, "/")
		ds.setValue(o, parts...)
	}
	return nil
}

func newDeepSet(to, from string, refSpec map[string][]string) *deepSet {
	return &deepSet{from: from, to: to, refSpec: refSpec}
}
