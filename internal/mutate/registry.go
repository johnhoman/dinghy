package mutate

import (
	"github.com/johnhoman/dinghy/internal/visitor"
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

// MustRegister registers a new Mutator under the provided name. newConfig should
// be a function that returns the required data structure for the Mutator. It
// will be up to the Mutator to convert the config to the correct type. The registry
// will handle deserializing YAML into the returned config when the Mutator plugin
// is invoked
func MustRegister(name string, vis any) {
	_, ok := vis.(visitor.Visitor)
	if !ok {
		panic("why why why")
	}
	r.store[name] = func() any {
		t := reflect.TypeOf(vis).Elem()
		return reflect.New(t).Interface()
	}
}

type registry struct {
	store map[string]func() any
}

var r = &registry{store: make(map[string]func() any)}

func init() {
	MustRegister("builtin.dinghy.dev/strategicMergePatch", &StrategicMergePatch{})
	MustRegister("builtin.dinghy.dev/jsonpatch", &JSONPatch{})
	MustRegister("builtin.dinghy.dev/jsonpatch/configmap", &ConfigMapJSONPatch{})
	MustRegister("builtin.dinghy.dev/patch", &Patch{})
	MustRegister("builtin.dinghy.dev/metadata", &Metadata{})
	MustRegister("builtin.dinghy.dev/metadata/name", &Name{})
	MustRegister("builtin.dinghy.dev/metadata/namespace", &Namespace{})
	MustRegister("builtin.dinghy.dev/metadata/annotations", &Annotations{})
	MustRegister("builtin.dinghy.dev/metadata/labels", &Labels{})
	MustRegister("builtin.dinghy.dev/script/js", &Script{})
}
