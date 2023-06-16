package mutate

import (
	"github.com/dop251/goja"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Script struct {
	Script string         `yaml:"script"`
	Config map[string]any `yaml:"config"`
}

// Visit runs a javascript snippet, passing is the current resource
// for mutation. The script MUST define a single function `mutate`
// with the signature mutate(obj, config), which wil be called with the
// resource being visited as well as the config provided to the mutator
func (s *Script) Visit(obj *unstructured.Unstructured) error {
	vm := goja.New()
	_, err := vm.RunString(s.Script)
	if err != nil {
		panic(err)
	}

	mutate, ok := goja.AssertFunction(vm.Get("mutate"))
	if !ok {
		return nil
	}

	_, err = mutate(goja.Undefined(), vm.ToValue(obj.Object), vm.ToValue(s.Config))
	return err
}
