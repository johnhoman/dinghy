package visitor

import (
	"github.com/dop251/goja"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Script runs a piece of Javascript in a VM
func Script(script string, config map[string]any) (Visitor, error) {
	vm := goja.New()
	_, err := vm.RunString(script)
	if err != nil {
		panic(err)
	}

	mutate, ok := goja.AssertFunction(vm.Get("mutate"))
	if !ok {
		// mutate function not defined, ignore
		return nopVisitor(), nil
	}

	return Func(func(obj *unstructured.Unstructured) error {
		_, err = mutate(goja.Undefined(), vm.ToValue(obj.Object), vm.ToValue(config))
		return err
	}), nil
}

func nopVisitor() Visitor {
	return Func(func(obj *unstructured.Unstructured) error {
		return nil
	})
}
