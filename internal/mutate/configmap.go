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

type ConfigMapJSONPatchConfig struct {
	// Key is the config map key to patch. The value
	// at the provided key will be decoded into
	Key   string `yaml:"key" dinghy:"required"`
	Patch []any  `yaml:"patch" dinghy:"required"`
}

func ConfigMapJSONPatch(config any) (visitor.Visitor, error) {
	c, ok := config.(*ConfigMapJSONPatchConfig)
	if !ok {
		return nil, ErrTypedConfig
	}
	raw, err := json.Marshal(c.Patch)
	if err != nil {
		return nil, err
	}
	patch := jsonpatch.Patch{}
	if err := json.Unmarshal(raw, &patch); err != nil {
		return nil, err
	}

	return visitor.Func(func(obj *unstructured.Unstructured) error {
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
		return visitor.Visit(visitor.JSONPatch(patch), m)
	}), nil
}
