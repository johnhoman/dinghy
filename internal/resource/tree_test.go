package resource

import (
	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/util/sets"
	"testing"
)

func TestTreeNode_Insert(t *testing.T) {
	tests := map[string]struct {
		obj *Object
		key Key
	}{
		"RoundTripInsertPop": {
			obj: Unstructured(map[string]any{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]any{
					"name":      "p1",
					"namespace": "n1",
				},
			}),
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
			qt.Assert(t, obj.Object, qt.DeepEquals, tt.obj.Object)
		})
	}
}

func TestTreeNode_Visit(t *testing.T) {
	tests := map[string]struct {
		initObjs []*Object
		options  []MatchOption
		// want is a list of names that were matched on
		want sets.Set[string]
	}{
		"MatchNames": {
			initObjs: []*Object{
				Unstructured(map[string]any{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]any{
						"name":      "p1",
						"namespace": "n1",
					},
				}),
			},
			options: []MatchOption{
				MatchNames("p1"),
			},
			want: sets.New[string]("v1.Pod/n1/p1"),
		},
	}

	v := func(set sets.Set[string]) Visitor {
		return VisitorFunc(func(obj *Object) error {
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
			qt.Assert(t, tree.Visit(v(set), tt.options...), qt.IsNil)
			qt.Assert(t, set, qt.Not(qt.HasLen), 0)
			qt.Assert(t, set.Difference(tt.want), qt.HasLen, 0)
		})
	}
}

func TestTreeNode_Pop(t *testing.T) {
	tests := map[string]struct {
		initObjs []*Object
		// want is a list of names that were matched on
		want Key
	}{
		"PopItem": {
			initObjs: []*Object{
				Unstructured(map[string]any{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]any{
						"name":      "p1",
						"namespace": "n1",
					},
				}),
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
