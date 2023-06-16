package generate

import (
	"github.com/pkg/errors"
	"reflect"
)

var (
	ErrNotFound = errors.New("mutator not found")
)

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

// Register registers a new Mutator under the provided name. newConfig should
// be a function that returns the required data structure for the Mutator. It
// will be up to the Mutator to convert the config to the correct type. The registry
// will handle deserializing YAML into the returned config when the Mutator plugin
// is invoked
func Register(name string, f Generator) {
	r.store[name] = func() any {
		t := reflect.TypeOf(f).Elem()
		return reflect.New(t).Interface()
	}
}

type registry struct {
	store map[string]func() any
}

var r = &registry{store: make(map[string]func() any)}

func init() {
	// only visitors can traverse the Tree, so the mutator registry should reference
	// visitors, and all the visitors can live in the visitor.
	Register("builtin.dinghy.dev/service", &Service{})
}
