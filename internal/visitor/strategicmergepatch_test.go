package visitor

import (
	qt "github.com/frankban/quicktest"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"testing"
)

type Meta map[string]any

func TestStrategicMergePatch(t *testing.T) {
	tests := map[string]struct {
		got   map[string]any
		patch map[string]any
		want  map[string]any
	}{
		"PatchesAPodSpec": {
			got: map[string]any{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]any{
					"name": "web-service",
				},
				"spec": map[string]any{
					"containers": []any{
						map[string]any{
							"name": "main",
							"env": []any{
								map[string]any{
									"name":  "FOO",
									"value": "BAR",
								},
							},
						},
					},
				},
			},
			patch: map[string]any{
				"spec": map[string]any{
					"containers": []any{
						map[string]any{
							"name": "main",
							"env": []any{
								map[string]any{
									"name":  "PIP_CONFIG",
									"value": "/var/run/secrets/pypi.python.org/pip.conf",
								},
							},
						},
					},
					"serviceAccountName": "web-service",
				},
			},
			want: map[string]any{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]any{
					"name": "web-service",
				},
				"spec": map[string]any{
					"containers": []any{
						map[string]any{
							"name": "main",
							"env": []any{
								map[string]any{
									"name":  "PIP_CONFIG",
									"value": "/var/run/secrets/pypi.python.org/pip.conf",
								},
								map[string]any{
									"name":  "FOO",
									"value": "BAR",
								},
							},
						},
					},
					"serviceAccountName": "web-service",
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			visitor := StrategicMergePatch(tt.patch)
			got := &unstructured.Unstructured{Object: tt.got}
			qt.Assert(t, visitor.Visit(got), qt.IsNil)
			qt.Assert(t, tt.got, qt.DeepEquals, tt.want)
		})
	}
}
