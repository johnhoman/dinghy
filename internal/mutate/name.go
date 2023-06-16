package mutate

import (
	_ "embed"
	"github.com/johnhoman/dinghy/internal/resource"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Name mutates the name on a resource and any resources that may
// reference that name
type Name struct {
	Prefix string `yaml:"prefix"`
	Suffix string `yaml:"suffix"`
}

// Visit sets the prefix and suffix on the metadata.name attribute
// of provided object.
func (n *Name) Visit(obj *unstructured.Unstructured) error {
	l := obj.GetLabels()
	if name, ok := l["app.kubernetes.io/name"]; ok && name == obj.GetName() {
		l["app.kubernetes.io/name"] = n.newName(obj)
	}
	if name, ok := l["app"]; ok && name == obj.GetName() {
		l["app"] = n.newName(obj)
	}
	obj.SetLabels(l)
	obj.SetName(n.newName(obj))
	return nil
}

// SideEffect searches for resource in the tree that might reference the mutated
// resource by name, such as a deployment referencing a configmap. If a ConfigMap
// name is changed, any Deployments, StatefulSets, pods, DaemonSets, ..., etc that
// reference it will also need to change the reference name
func (n *Name) SideEffect(old *unstructured.Unstructured, tree resource.Tree) error {
	key := old.GroupVersionKind().GroupKind().String()
	return tree.Visit(newDeepSet(n.newName(old), old.GetName(), nameRefs[key]))
}

func (n *Name) newName(obj *unstructured.Unstructured) string {
	return n.Prefix + obj.GetName() + n.Suffix
}

var (
	//go:embed name.yaml
	nameRefContent []byte
	nameRefs       map[string]map[string][]string
)

func init() {
	nameRefs = make(map[string]map[string][]string)
	err := yaml.Unmarshal(nameRefContent, &nameRefs)
	if err != nil {
		panic(errors.Wrap(err, "failed to unmarshal configmap refs"))
	}
}
