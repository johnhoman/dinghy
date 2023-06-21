package mutate

import (
	"github.com/johnhoman/dinghy/internal/resource"
	"github.com/pkg/errors"
	"reflect"
)

var (
	ErrNotFound = errors.New("mutator not found")
)

type Mutator interface {
	resource.Visitor
	Name() string
}

func Get(name string) (any, error) {
	f, ok := r.store[name]
	if !ok {
		return nil, ErrNotFound
	}
	return f(), nil
}

func Has(name string) bool {
	_, ok := r.store[name]
	return ok
}

// MustRegister registers a new Mutator under the provided name. newConfig should
// be a function that returns the required data structure for the Mutator. It
// will be up to the Mutator to convert the config to the correct type. The registry
// will handle deserializing YAML into the returned config when the Mutator plugin
// is invoked
func MustRegister(vis Mutator) {
	r.store[vis.Name()] = func() any {
		t := reflect.TypeOf(vis).Elem()
		return reflect.New(t).Interface()
	}
}

type registry struct {
	store map[string]func() any
}

var r = &registry{store: make(map[string]func() any)}

func init() {
	MustRegister(&StrategicMergePatch{})
	MustRegister(&MergePatch{})
	MustRegister(&JSONPatch{})
	MustRegister(&ConfigMapJSONPatch{})
	MustRegister(&Patch{})
	MustRegister(&Metadata{})
	MustRegister(&Name{})
	MustRegister(&Namespace{})
	MustRegister(&Annotations{})
	MustRegister(&Labels{})
	MustRegister(&MatchLabels{})
	MustRegister(&Script{})
}
