package mutate

import (
	"encoding/json"
	jsonpatch "github.com/evanphx/json-patch"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/johnhoman/dinghy/internal/visitor"
)

var (
	_ yaml.Unmarshaler = &JSONPatch{}
)

type JSONPatch jsonpatch.Patch

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
	return json.Unmarshal(raw, j)
}

func (j *JSONPatch) Visit(obj *unstructured.Unstructured) error {
	return visitor.JSONPatch(jsonpatch.Patch(*j)).Visit(obj)
}
