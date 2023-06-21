package mutate

import (
	"encoding/json"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/johnhoman/dinghy/internal/resource"
	"gopkg.in/yaml.v3"
)

var (
	_ yaml.Unmarshaler = &JSONPatch{}
	_ Mutator          = &JSONPatch{}
)

type JSONPatch struct {
	patch jsonpatch.Patch
}

func (j *JSONPatch) Name() string {
	return "builtin.dinghy.dev/jsonpatch"
}

func (j *JSONPatch) UnmarshalYAML(value *yaml.Node) error {
	// the json patch type has a json.RawMessage field that yaml
	// can decode, so it has to be encoded from a normal object to json,
	// then json decoded into the jsonpatch.Patch

	var patch []any
	if err := value.Decode(&patch); err != nil {
		return err
	}
	raw, err := json.Marshal(patch)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, &j.patch)
}

func (j *JSONPatch) Visit(obj *resource.Object) error {
	return obj.JSONPatch(j.patch)
}
