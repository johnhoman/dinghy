package mutate

import (
	"bytes"
	"encoding/json"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/johnhoman/dinghy/internal/resource"
	"github.com/johnhoman/dinghy/internal/visitor"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	ErrKeyError = "the requested key could not be found"
)

var (
	_ Mutator          = &ConfigMapJSONPatch{}
	_ resource.Visitor = &ConfigMapJSONPatch{}
)

type ConfigMapJSONPatch struct {
	// Key is the config map key to patch. The value
	// at the provided key will be decoded into
	Key   string          `yaml:"key"`
	Patch jsonpatch.Patch `yaml:"patch"`
}

func (c *ConfigMapJSONPatch) Name() string {
	return "builtin.dinghy.dev/jsonpatch/configmap"
}

func (c *ConfigMapJSONPatch) UnmarshalYAML(value *yaml.Node) error {
	var in struct {
		Key   string `yaml:"key"`
		Patch []any  `yaml:"patch"`
	}

	var m map[string]any
	if err := value.Decode(&m); err != nil {
		return err
	}
	data, _ := yaml.Marshal(m)
	d := yaml.NewDecoder(bytes.NewReader(data))
	d.KnownFields(true)
	if err := d.Decode(&in); err != nil {
		return err
	}

	var patch jsonpatch.Patch
	data, err := json.Marshal(in.Patch)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, &patch); err != nil {
		return err
	}
	c.Key = in.Key
	c.Patch = patch
	return nil
}

func (c *ConfigMapJSONPatch) Visit(obj *resource.Object) error {
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
