package mutate

import (
	"github.com/johnhoman/dinghy/internal/resource"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

var (
	_ Mutator = &Annotations{}
	_ Mutator = &Namespace{}
	_ Mutator = &Metadata{}

	_ yaml.Unmarshaler = &Namespace{}
	_ yaml.Unmarshaler = &Metadata{}
)

type Annotations map[string]string

func (m *Annotations) Name() string {
	return "builtin.dinghy.dev/metadata/annotations"
}

func (m *Annotations) Visit(obj *resource.Object) error {
	if obj == nil {
		return errors.Errorf("resource cannot be nil")
	}
	obj.AddAnnotations(*m)
	return nil
}

type Namespace struct {
	Namespace string `yaml:"name"`
}

func (n *Namespace) UnmarshalYAML(value *yaml.Node) error {
	var in struct {
		Name string `yaml:"name"`
	}
	if err := value.Decode(&in); err != nil {
		return err
	}
	n.Namespace = in.Name
	return nil
}

func (n *Namespace) Name() string {
	return "builtin.dinghy.dev/metadata/namespace"
}

func (n *Namespace) Visit(obj *resource.Object) error {
	if obj == nil {
		return errors.Errorf("resource cannot be nil")
	}
	obj.SetNamespace(n.Namespace)
	return nil
}

type Metadata struct {
	patch map[string]any
}

func (m *Metadata) UnmarshalYAML(value *yaml.Node) error {
	var in map[string]any
	if err := value.Decode(&in); err != nil {
		return err
	}
	m.patch = in
	return nil
}

func (m *Metadata) Name() string {
	return "builtin.dinghy.dev/metadata"
}

func (m *Metadata) Visit(obj *resource.Object) error {
	return obj.MergePatch(map[string]any{"metadata": m.patch})
}
