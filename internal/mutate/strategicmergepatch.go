package mutate

import (
	"github.com/imdario/mergo"
	"github.com/johnhoman/dinghy/internal/resource"
	"gopkg.in/yaml.v3"
)

var (
	_ yaml.Unmarshaler = &StrategicMergePatch{}
	_ Mutator          = &StrategicMergePatch{}
	_ resource.Visitor = &StrategicMergePatch{}
	_ yaml.Unmarshaler = &MergePatch{}
	_ Mutator          = &MergePatch{}
	_ resource.Visitor = &MergePatch{}
)

type MergePatch struct {
	patch map[string]any
}

func (m *MergePatch) Name() string {
	return "builtin.dinghy.dev/mergePatch"
}

func (m *MergePatch) UnmarshalYAML(value *yaml.Node) error {
	m.patch = make(map[string]any)
	return value.Decode(&m.patch)
}

func (m *MergePatch) Visit(obj *resource.Object) error {
	return mergo.Merge(obj.Object, m.patch, mergo.WithOverride)
}

type StrategicMergePatch struct {
	patch map[string]any
}

func (s *StrategicMergePatch) Name() string {
	return "builtin.dinghy.dev/strategicMergePatch"
}

func (s *StrategicMergePatch) UnmarshalYAML(value *yaml.Node) error {
	var m map[string]any
	if err := value.Decode(&m); err != nil {
		return err
	}
	s.patch = m
	return nil
}

func (s *StrategicMergePatch) Visit(obj *resource.Object) error {
	return obj.StrategicMergePatch(s.patch)
}
