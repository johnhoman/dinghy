package mutate

import (
	"github.com/dop251/goja"
	"github.com/johnhoman/dinghy/internal/resource"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

var (
	_ Mutator          = &Script{}
	_ yaml.Unmarshaler = &Script{}
)

type Script struct {
	vm     *goja.Runtime
	mutate resource.Visitor
}

func (s *Script) Name() string {
	return "builtin.dinghy.dev/script/js"
}

func (s *Script) UnmarshalYAML(value *yaml.Node) error {
	var in struct {
		Script string         `yaml:"script"`
		Config map[string]any `yaml:"config"`
	}
	if err := value.Decode(&in); err != nil {
		return err
	}

	vm := goja.New()
	if _, err := vm.RunString(in.Script); err != nil {
		return err
	}

	mutate, ok := goja.AssertFunction(vm.Get("mutate"))
	if !ok {
		return errors.New("mutate function not found")
	}
	s.mutate = resource.VisitorFunc(func(obj *resource.Object) error {
		_, err := mutate(goja.Undefined(), vm.ToValue(obj.Object), vm.ToValue(in.Config))
		return err
	})
	return nil
}

// Visit runs a javascript snippet, passing is the current resource
// for mutation. The script MUST define a single function `mutate`
// with the signature mutate(obj, config), which wil be called with the
// resource being visited as well as the config provided to the mutator
func (s *Script) Visit(obj *resource.Object) error { return s.mutate.Visit(obj) }
