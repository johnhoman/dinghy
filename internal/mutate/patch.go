package mutate

import (
	"github.com/johnhoman/dinghy/internal/fieldpath"
	"github.com/johnhoman/dinghy/internal/resource"
	"gopkg.in/yaml.v3"
)

var (
	_ Mutator          = &Patch{}
	_ yaml.Unmarshaler = &Patch{}
)

type Patch struct {
	FieldPaths []*fieldpath.FieldPath
	Value      any
}

func (p *Patch) Name() string {
	return "builtin.dinghy.dev/patch"
}

func (p *Patch) UnmarshalYAML(value *yaml.Node) error {
	var in struct {
		FieldPaths []string `yaml:"fieldPaths"`
		Value      any      `yaml:"value"`
	}
	if err := value.Decode(&in); err != nil {
		return err
	}
	p.Value = in.Value
	p.FieldPaths = make([]*fieldpath.FieldPath, len(in.FieldPaths))
	for k, fp := range in.FieldPaths {
		parsed, err := fieldpath.Parse(fp)
		if err != nil {
			return err
		}
		p.FieldPaths[k] = parsed
	}
	return nil
}

func (p *Patch) Visit(obj *resource.Object) error {
	for _, fp := range p.FieldPaths {
		if err := obj.FieldPatch(fp, p.Value); err != nil {
			return err
		}
	}
	return nil
}
