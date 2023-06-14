package resource

import (
	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp"
	"github.com/johnhoman/dinghy/internal/visitor"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/sets"
	"testing"
)

func TestTreeNode_Insert(t *testing.T) {
	tests := map[string]struct {
		obj *unstructured.Unstructured
		key Key
	}{
		"RoundTripInsertPop": {
			obj: &unstructured.Unstructured{Object: map[string]any{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]any{
					"name":      "p1",
					"namespace": "n1",
				},
			}},
			key: Key{
				GroupVersion: "v1",
				Kind:         "Pod",
				Name:         "p1",
				Namespace:    "n1",
			},
		},
	}
	for name, tt := range tests {
		tree := NewTree()
		t.Run(name, func(t *testing.T) {
			qt.Assert(t, tree.Insert(tt.obj), qt.IsNil)
			obj, err := tree.Pop(tt.key)
			qt.Assert(t, err, qt.IsNil)
			qt.Assert(t, obj, qt.DeepEquals, obj)
		})
	}
}

func TestTreeNode_Visit(t *testing.T) {
	tests := map[string]struct {
		initObjs []*unstructured.Unstructured
		options  []MatchOption
		// want is a list of names that were matched on
		want sets.Set[string]
	}{
		"MatchNames": {
			initObjs: []*unstructured.Unstructured{
				{
					Object: map[string]any{
						"apiVersion": "v1",
						"kind":       "Pod",
						"metadata": map[string]any{
							"name":      "p1",
							"namespace": "n1",
						},
					},
				},
			},
			options: []MatchOption{
				MatchNames("p1"),
			},
			want: sets.New[string]("v1.Pod/n1/p1"),
		},
	}

	v := func(set sets.Set[string]) visitor.Visitor {
		return visitor.Func(func(obj *unstructured.Unstructured) error {
			set.Insert(newResourceKey(obj).String())
			return nil
		})
	}
	for name, tt := range tests {
		tree := NewTree()
		t.Run(name, func(t *testing.T) {
			for _, obj := range tt.initObjs {
				qt.Assert(t, tree.Insert(obj), qt.IsNil)
			}
			set := sets.New[string]()
			err := tree.Visit(v(set), tt.options...)
			qt.Assert(t, err, qt.IsNil)
			qt.Assert(t, set, qt.Not(qt.HasLen), 0)
			qt.Assert(t, set.Difference(tt.want), qt.HasLen, 0)
		})
	}
}

func TestTreeNode_Pop(t *testing.T) {
	tests := map[string]struct {
		initObjs []*unstructured.Unstructured
		// want is a list of names that were matched on
		want Key
	}{
		"PopItem": {
			initObjs: []*unstructured.Unstructured{
				{
					Object: map[string]any{
						"apiVersion": "v1",
						"kind":       "Pod",
						"metadata": map[string]any{
							"name":      "p1",
							"namespace": "n1",
						},
					},
				},
			},
			want: Key{
				Name:         "p1",
				Namespace:    "n1",
				Kind:         "Pod",
				GroupVersion: "v1",
			},
		},
	}

	for name, tt := range tests {
		tree := NewTree()
		t.Run(name, func(t *testing.T) {
			for _, obj := range tt.initObjs {
				qt.Assert(t, tree.Insert(obj), qt.IsNil)
			}
			obj, err := tree.Pop(tt.want)
			qt.Assert(t, err, qt.IsNil)
			qt.Assert(t,
				newResourceKey(obj),
				qt.CmpEquals(cmp.AllowUnexported(Key{})),
				tt.want,
			)
		})
	}
}
