package mutate

import (
	"github.com/johnhoman/dinghy/internal/visitor"
	"github.com/pkg/errors"
)

var (
	ErrTypedConfig = errors.New("failed convert to typed config")
	ErrNotFound    = errors.New("mutator not found")
)

func Get(name string) (Func, NewConfigFunc, error) {
	m, ok := r.store[name]
	if !ok {
		return nil, nil, ErrNotFound
	}
	return m.m, m.newConfig, nil
}

func Has(name string) bool {
	_, ok := r.store[name]
	return ok
}

type Func func(config any) (visitor.Visitor, error)

// Register registers a new Mutator under the provided name. newConfig should
// be a function that returns the required data structure for the Mutator. It
// will be up to the Mutator to convert the config to the correct type. The registry
// will handle deserializing YAML into the returned config when the Mutator plugin
// is invoked
func Register(name string, f Func, newConfig NewConfigFunc) error {
	if newConfig == nil {
		newConfig = func() any {
			return make(map[string]any)
		}
	}
	r.store[name] = entry{m: f, newConfig: newConfig}
	return nil
}

// MustRegister is just like Register, but it panics it can't register the
// mutator.
func MustRegister(name string, f Func, newConfig func() any) {
	if err := Register(name, f, newConfig); err != nil {
		panic(err)
	}
}

type entry struct {
	m         Func
	newConfig func() any
}

type registry struct {
	store map[string]entry
}

var r = &registry{store: make(map[string]entry)}

func init() {
	type scriptConfig struct {
		Script string         `yaml:"script" dinghy:"required"`
		Config map[string]any `yaml:"config" dinghy:"required"`
	}
	// only visitors can traverse the tree, so the mutator registry should reference
	// visitors, and all the visitors can live in the visitor.
	MustRegister("builtin.dinghy.dev/strategicMergePatch", StrategicMergePatch, newAnyMap)
	MustRegister("builtin.dinghy.dev/jsonpatch", JSONPatch, newAnySlice)
	MustRegister("builtin.dinghy.dev/metadata", Metadata, newAnyMap)
	MustRegister("builtin.dinghy.dev/metadata/name", Name, newNameConfig)
	MustRegister("builtin.dinghy.dev/metadata/namespace", Namespace, newNamespaceConfig)
	MustRegister("builtin.dinghy.dev/metadata/annotations", AddAnnotations, newConfig[map[string]string]())
	MustRegister("builtin.dinghy.dev/metadata/annotations/set", SetAnnotations, newStringMap)
	MustRegister("builtin.dinghy.dev/script/js", func(config any) (visitor.Visitor, error) {
		c, ok := config.(*scriptConfig)
		if !ok {
			return nil, ErrTypedConfig
		}
		return visitor.Script(c.Script, c.Config)
	}, func() any {
		return &scriptConfig{}
	})

}
