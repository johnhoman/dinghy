package generate

import (
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/johnhoman/dinghy/internal/context"
	"github.com/johnhoman/dinghy/internal/path"
	"github.com/johnhoman/dinghy/internal/resource"
)

func TestTemplate_UnmarshalYAML(t *testing.T) {
	c := qt.New(t)
	data := []byte(`
source: "github.com/johnhoman/nop"
values:
  appName: feast
  appNamespace: feature-system
`)

	tmp := &Template{}
	c.Assert(yaml.Unmarshal(data, tmp), qt.IsNil)
	c.Assert(tmp.source.String(), qt.Equals, "https://github.com/johnhoman/nop")
	c.Assert(tmp.values, qt.DeepEquals, map[string]any{
		"appName":      "feast",
		"appNamespace": "feature-system",
	})
}

func TestTemplate_UnmarshalYAML_InvalidConfig(t *testing.T) {
	c := qt.New(t)
	data := []byte(`
source: "github.com/johnhoman/nop"
v:
  appName: feast
  appNamespace: feature-system
`)

	err := yaml.Unmarshal(data, &Template{})
	c.Assert(err, qt.IsNotNil)
}

func TestTemplate_Emit(t *testing.T) {
	c := qt.New(t)

	data := []byte(`
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ .appName }}
  namespace: {{ .appNamespace }}
`)
	tmpdir, err := os.MkdirTemp("", "dinghy.testing.*")
	c.Assert(err, qt.IsNil)
	defer func() {
		c.Assert(os.RemoveAll(tmpdir), qt.IsNil)
	}()
	c.Assert(os.WriteFile(filepath.Join(tmpdir, "template.tmpl"), data, 0755), qt.IsNil)

	tmp := &Template{
		source: path.MustParse("template.tmpl"),
		values: map[string]any{
			"appName":      "feast",
			"appNamespace": "feature-system",
		},
	}
	ctx := context.NewContext(true)
	ctx.SetRoot(tmpdir)
	tree, err := tmp.Emit(ctx)
	c.Assert(err, qt.IsNil)

	expected := resource.Unstructured(map[string]any{
		"apiVersion": "v1",
		"kind":       "ServiceAccount",
		"metadata": map[string]any{
			"name":      "feast",
			"namespace": "feature-system",
		},
	})

	obj, err := tree.Pop(resource.ParseKey(expected))
	c.Assert(err, qt.IsNil)
	c.Assert(obj.UnstructuredContent(), qt.DeepEquals, expected.UnstructuredContent())
}

func TestTemplate_Emit_MissingKeys(t *testing.T) {
	c := qt.New(t)

	data := []byte(`
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ .appName }}
  namespace: {{ .appNamespace }}
`)
	tmpdir, err := os.MkdirTemp("", "dinghy.testing.*")
	c.Assert(err, qt.IsNil)
	defer func() {
		c.Assert(os.RemoveAll(tmpdir), qt.IsNil)
	}()
	c.Assert(os.WriteFile(filepath.Join(tmpdir, "template.tmpl"), data, 0755), qt.IsNil)

	tmp := &Template{
		source: path.MustParse("template.tmpl"),
		values: map[string]any{
			"appName": "feast",
		},
	}
	ctx := context.NewContext(true)
	ctx.SetRoot(tmpdir)
	_, err = tmp.Emit(ctx)
	c.Assert(err, qt.IsNotNil)
}
