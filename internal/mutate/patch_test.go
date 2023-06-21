package mutate

import (
	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp"
	"github.com/johnhoman/dinghy/internal/fieldpath"
	"github.com/johnhoman/dinghy/internal/resource"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"testing"
)

func TestPatch_UnmarshalYAML_DecodesFieldsAndParsesFieldPaths(t *testing.T) {
	c := qt.New(t)

	yamlData := `
fieldPaths:
  - spec.replicas
  - metadata.labels.app
value: 10
`
	patch := &Patch{}
	c.Assert(yaml.Unmarshal([]byte(yamlData), patch), qt.IsNil)

	expectedFieldPaths := []*fieldpath.FieldPath{
		fieldpath.MustParse("spec.replicas"),
		fieldpath.MustParse("metadata.labels.app"),
	}
	c.Assert(patch.FieldPaths, qt.CmpEquals(
		cmp.AllowUnexported(fieldpath.FieldPath{}),
		cmp.AllowUnexported(fieldpath.Index{}),
		cmp.AllowUnexported(fieldpath.Query{}),
	), expectedFieldPaths)
	c.Assert(patch.Value, qt.Equals, 10)
}

func TestPatch_UnmarshalYAML_ReturnsErrorIfFieldPathParsingFails(t *testing.T) {
	c := qt.New(t)

	yamlData := `
fieldPaths:
  - spec.replicas
  - invalid.[app.kubernetes.io/name
value: 10
`
	patch := &Patch{}
	err := yaml.Unmarshal([]byte(yamlData), patch)
	c.Assert(err, qt.Not(qt.IsNil))
}

func TestPatch_Visit_AppliesFieldPatchesToResourceObject(t *testing.T) {
	c := qt.New(t)

	patch := &Patch{
		FieldPaths: []*fieldpath.FieldPath{
			fieldpath.MustParse("metadata.labels.app"),
			fieldpath.MustParse("spec.selector.matchLabels['app.kubernetes.io/name']"),
			fieldpath.MustParse("spec.template.metadata.labels['app.kubernetes.io/name']"),
		},
		Value: "example",
	}

	obj := &resource.Object{
		Unstructured: &unstructured.Unstructured{},
	}
	obj.SetAPIVersion("apps/v1")
	obj.SetKind("Deployment")
	obj.SetName("example")
	obj.SetLabels(map[string]string{
		"app": "my-app",
	})

	err := patch.Visit(obj)
	c.Assert(err, qt.IsNil)

	// Validate the field patches are applied correctly.
	expected := resource.Unstructured(map[string]any{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]any{
			"name": "example",
			"labels": map[string]any{
				"app": "example",
			},
		},
		"spec": map[string]any{
			"selector": map[string]any{
				"matchLabels": map[string]any{
					"app.kubernetes.io/name": "example",
				},
			},
			"template": map[string]any{
				"metadata": map[string]any{
					"labels": map[string]any{
						"app.kubernetes.io/name": "example",
					},
				},
			},
		},
	})
	c.Assert(obj.UnstructuredContent(), qt.DeepEquals, expected.UnstructuredContent())
}
