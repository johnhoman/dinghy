package visitor

import (
	qt "github.com/frankban/quicktest"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"testing"
)

func TestScript(t *testing.T) {
	tests := map[string]struct {
		script string
		config map[string]any
		in     *unstructured.Unstructured
		want   *unstructured.Unstructured
	}{
		"SetsUpAMutator": {
			script: `
function mutate(o, c) {
  if (o.metadata === undefined) {
    o.metadata = {}
  }
  o.metadata.name = "foo"
  o.metadata.namespace = "bar"
}
`,
			config: make(map[string]any),
			in:     U(map[string]any{}),
			want: U(map[string]any{
				"metadata": map[string]any{
					"name":      "foo",
					"namespace": "bar",
				},
			}),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			vis, err := Script(tt.script, tt.config)
			qt.Assert(t, err, qt.IsNil)
			err = vis.Visit(tt.in)
			qt.Assert(t, err, qt.IsNil)
			qt.Assert(t, tt.in, qt.DeepEquals, tt.want)
		})
	}
}

func U(m map[string]any) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: m}
}
