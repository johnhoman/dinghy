package mutate

import (
	_ "embed"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"strings"

	"github.com/johnhoman/dinghy/internal/resource"
)

var (
	_ Mutator          = &Labels{}
	_ Mutator          = &MatchLabels{}
	_ yaml.Unmarshaler = &MatchLabels{}
)

// Labels mutates the labels on a resource and any selectors that may
// reference that label
type Labels map[string]string

func (l *Labels) Name() string {
	return "builtin.dinghy.dev/metadata/labels"
}

// Visit sets the prefix and suffix on the metadata.name attribute
// of provided object.
func (l *Labels) Visit(obj *resource.Object) error {
	obj.AddLabels(*l)
	return nil
}

// MatchLabels get applied to all resources, but in some cases also
// make changes to the resource spec, such as matchLabels in a deployment,
// or service labels in a selector.
type MatchLabels struct {
	m map[string]string
}

func (l *MatchLabels) Name() string {
	return "builtin.dinghy.dev/matchLabels"
}

func (l *MatchLabels) UnmarshalYAML(value *yaml.Node) error {
	return value.Decode(&l.m)
}

func (l *MatchLabels) Visit(obj *resource.Object) error {
	obj.AddLabels(l.m)
	key := obj.GroupVersionKind().GroupKind().String()
	paths, ok := labelRefs[key]
	if !ok {
		return nil
	}
	for lk, lv := range l.m {
		for _, path := range paths {
			current := obj.Object
			parts := strings.Split(path, "/")
			for k, part := range parts {
				next, ok := current[part]
				if !ok {
					current[part] = make(map[string]any)
					next = current[part]
				}
				current, ok = next.(map[string]any)
				if !ok {
					return errors.Errorf("expected field to be a mapping, but got %T: %s",
						next, strings.Join(parts[:k+1], "."))
				}
			}
			current[lk] = lv
		}
	}
	return nil
}

type deepSetNameRef struct {
	to      string
	from    string
	refSpec map[string][]string
}

func (ds *deepSetNameRef) setValue(obj map[string]any, path ...string) {
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

func (ds *deepSetNameRef) Visit(obj *resource.Object) error {
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

func newDeepSet(to, from string, refSpec map[string][]string) *deepSetNameRef {
	return &deepSetNameRef{from: from, to: to, refSpec: refSpec}
}

var (
	//go:embed labels.yaml
	labelRefContent []byte
	labelRefs       map[string][]string
)

func init() {
	labelRefs = make(map[string][]string)
	err := yaml.Unmarshal(labelRefContent, &labelRefs)
	if err != nil {
		panic("failed to decode label refs")
	}
}
