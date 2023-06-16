package mutate

import (
	"encoding/json"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/johnhoman/dinghy/internal/visitor"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	ErrKeyError = "the requested key could not be found"
)

type ConfigMapJSONPatch struct {
	// Key is the config map key to patch. The value
	// at the provided key will be decoded into
	Key   string `yaml:"key" dinghy:"required"`
	Patch []any  `yaml:"patch" dinghy:"required"`
}

func (c *ConfigMapJSONPatch) Visit(obj *unstructured.Unstructured) error {
	raw, err := json.Marshal(c.Patch)
	if err != nil {
		return err
	}
	patch := jsonpatch.Patch{}
	if err := json.Unmarshal(raw, &patch); err != nil {
		return err
	}
	o := obj.UnstructuredContent()
	v, ok, err := unstructured.NestedString(o, "data", c.Key)
	if err != nil {
		return err
	}
	if !ok {
		return errors.Errorf("%s: %q", ErrKeyError, c.Key)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(v), &m); err != nil {
		return err
	}
	if err := visitor.Visit(visitor.JSONPatch(patch), m); err != nil {
		return err
	}

	if err := unstructured.SetNestedField(o, "data", c.Key); err != nil {
		return err
	}
	obj.SetUnstructuredContent(o)
	return nil
}
