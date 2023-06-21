package mutate

import (
	qt "github.com/frankban/quicktest"
	"github.com/johnhoman/dinghy/internal/resource"
	"testing"
)

func TestCommonLabels_Visit(t *testing.T) {
	tests := map[string]struct {
		l       MatchLabels
		obj     *resource.Object
		want    map[string]any
		wantErr error
	}{
		"Deployment": {
			l: MatchLabels{
				m: map[string]string{
					"app.kubernetes.io/part-of": "feast",
				},
			},
			obj: resource.Unstructured(map[string]any{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]any{
					"name":      "api-server",
					"namespace": "feast-system",
				},
			}),
			want: map[string]any{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]any{
					"name":      "api-server",
					"namespace": "feast-system",
					"labels": map[string]any{
						"app.kubernetes.io/part-of": "feast",
					},
				},
				"spec": map[string]any{
					"selector": map[string]any{
						"matchLabels": map[string]any{
							"app.kubernetes.io/part-of": "feast",
						},
					},
					"template": map[string]any{
						"metadata": map[string]any{
							"labels": map[string]any{
								"app.kubernetes.io/part-of": "feast",
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
			qt.Assert(t, tt.l.Visit(tt.obj), qt.IsNil)
			qt.Assert(t, tt.obj.UnstructuredContent(), qt.DeepEquals, tt.want)
		})
	}
}
