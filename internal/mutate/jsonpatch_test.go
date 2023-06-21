package mutate

import (
	qt "github.com/frankban/quicktest"
	"github.com/johnhoman/dinghy/internal/resource"
	"gopkg.in/yaml.v3"
	"testing"
)

func TestJSONPatch_UnmarshalYAML(t *testing.T) {
	jp := &JSONPatch{}
	patch := []byte(`
- op: replace
  path: /spec/template/spec/containers/0/name
  value: main
`)
	qt.Assert(t, yaml.Unmarshal(patch, jp), qt.IsNil)
	qt.Assert(t, len(jp.patch), qt.Equals, 1)
}

func TestJSONPatch_Visit(t *testing.T) {

	tests := map[string]struct {
		jp      *JSONPatch
		obj     *resource.Object
		want    map[string]any
		wantErr error
	}{
		"PatchesAResource": {
			jp: jsonpatchDecode(t, `
- op: replace
  path: /spec/template/spec/serviceAccountName
  value: editor
- op: replace
  path: /spec/template/spec/nodeSelector
  value: {kind: spot}
`),
			obj: resource.Unstructured(map[string]any{
				"spec": map[string]any{
					"template": map[string]any{
						"spec": map[string]any{
							"serviceAccountName": "default",
						},
					},
				},
			}),
			want: map[string]any{
				"spec": map[string]any{
					"template": map[string]any{
						"spec": map[string]any{
							"serviceAccountName": "editor",
							"nodeSelector": map[string]any{
								"kind": "spot",
							},
						},
					},
				},
			},
			wantErr: nil,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			qt.Assert(t, tt.jp.Visit(tt.obj), qt.Equals, tt.wantErr)
			qt.Assert(t, tt.obj.UnstructuredContent(), qt.DeepEquals, tt.want)
		})
	}
}

func jsonpatchDecode(t *testing.T, data string) *JSONPatch {
	patch := JSONPatch{}
	qt.Assert(t, yaml.Unmarshal([]byte(data), &patch), qt.IsNil)
	return &patch
}
