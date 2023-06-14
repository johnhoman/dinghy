package generate

import (
	"github.com/pkg/errors"
)

var (
	ErrTypedConfig = errors.New("failed convert to typed config")
	ErrNotFound    = errors.New("mutator not found")
)

func Get(name string) (Generator, NewConfigFunc, error) {
	e, ok := r.store[name]
	if !ok {
		return nil, nil, ErrNotFound
	}
	return e.g, e.newConfig, nil
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
func Register(name string, f Generator, newConfig NewConfigFunc) error {
	if newConfig == nil {
		newConfig = func() any {
			return make(map[string]any)
		}
	}
	r.store[name] = entry{g: f, newConfig: newConfig}
	return nil
}

// MustRegister is just like Register, but it panics it can't register the
// mutator.
func MustRegister(name string, f Generator, newConfig func() any) {
	if err := Register(name, f, newConfig); err != nil {
		panic(err)
	}
}

type entry struct {
	g         Generator
	newConfig func() any
}

type registry struct {
	store map[string]entry
}

var r = &registry{store: make(map[string]entry)}

func init() {
	// only visitors can traverse the Tree, so the mutator registry should reference
	// visitors, and all the visitors can live in the visitor.
	MustRegister("builtin.dinghy.dev/service", Service(), newConfig[ServiceConfig]())
}
