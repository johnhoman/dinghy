package mutate

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"gopkg.in/yaml.v3"

	"github.com/johnhoman/dinghy/internal/resource"
)

func TestStrategicMergePatch_UnmarshalYAML_ValidPatch(t *testing.T) {
	c := qt.New(t)

	yamlData := `
name: example
labels:
  app: example-app
  version: v1
`
	var patch StrategicMergePatch
	err := yaml.Unmarshal([]byte(yamlData), &patch)
	c.Assert(err, qt.IsNil)
	c.Assert(patch.patch, qt.Not(qt.IsNil))
	c.Assert(patch.patch["name"], qt.Equals, "example")
	c.Assert(patch.patch["labels"].(map[string]interface{})["app"], qt.Equals, "example-app")
	c.Assert(patch.patch["labels"].(map[string]interface{})["version"], qt.Equals, "v1")
}

func TestStrategicMergePatch_UnmarshalYAML_InvalidPatch(t *testing.T) {
	c := qt.New(t)

	yamlData := `
- invalid
- patch
`

	var patch StrategicMergePatch
	err := yaml.Unmarshal([]byte(yamlData), &patch)
	c.Assert(err, qt.Not(qt.IsNil))
}

func TestStrategicMergePatch_Visit(t *testing.T) {
	c := qt.New(t)

	obj := resource.Unstructured(map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]any{
			"name": "example",
			"labels": map[string]interface{}{
				"app":     "old-app",
				"version": "old-version",
			},
		},
	})

	patch := StrategicMergePatch{
		patch: map[string]any{
			"metadata": map[string]any{
				"labels": map[string]interface{}{
					"app":     "new-app",
					"version": "new-version",
				},
			},
		},
	}

	err := patch.Visit(obj)
	c.Assert(err, qt.IsNil)

	meta := obj.Object["metadata"].(map[string]any)

	c.Assert(meta["name"], qt.Equals, "example")
	c.Assert(meta["labels"].(map[string]any)["app"], qt.Equals, "new-app")
	c.Assert(meta["labels"].(map[string]interface{})["version"], qt.Equals, "new-version")
}
