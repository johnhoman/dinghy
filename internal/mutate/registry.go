package mutate

import (
	"github.com/johnhoman/dinghy/internal/visitor"
	"github.com/pkg/errors"
)

var (
	ErrTypedConfig = errors.New("failed convert to typed config")
)

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
	// only visitors can traverse the tree, so the mutator registry should reference
	// visitors, and all the visitors can live in the visitor.
	MustRegister("builtin.dinghy.dev/strategicMergePatch", StrategicMergePatch, newConfig[map[string]any]())
	MustRegister("builtin.dinghy.dev/jsonpatch", JSONPatch, newConfig[[]any]())
	MustRegister("builtin.dinghy.dev/metadata", Metadata, newConfig[map[string]any]())
	MustRegister("builtin.dinghy.dev/metadata/name", Name, newConfig[visitor.NameConfig]())
	MustRegister("builtin.dinghy.dev/metadata/namespace", Namespace, newConfig[visitor.NamespaceConfig]())
	MustRegister("builtin.dinghy.dev/metadata/annotations", AddAnnotations, newConfig[map[string]string]())
	MustRegister("builtin.dinghy.dev/metadata/annotations/set", SetAnnotations, newConfig[map[string]string]())

	// TODO: mapping of what mutators work on which resources, e.g. name cannot be used on
	//   CRDs.
}
