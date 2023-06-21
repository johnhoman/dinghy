package resource

import (
	"encoding/json"
	"github.com/imdario/mergo"
	"github.com/johnhoman/dinghy/internal/errors"
	"reflect"
	"sync"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/strategicpatch"

	"github.com/johnhoman/dinghy/internal/fieldpath"
	"github.com/johnhoman/dinghy/internal/scheme"
)

type Event struct{}

type Trackable interface {
	Record(event Event)
}

type Visitor interface {
	Visit(obj *Object) error
}

type VisitorFunc func(obj *Object) error

func (f VisitorFunc) Visit(obj *Object) error { return f(obj) }

type Option func(o *Object)

func Unstructured(m map[string]any) *Object {
	o := &Object{
		Unstructured: &unstructured.Unstructured{Object: m},
		mu:           sync.RWMutex{},
		events:       make([]Event, 0),
	}
	gvk := o.GroupVersionKind()
	o.matchKeys = []schema.GroupVersionKind{
		{Group: "*", Version: "*", Kind: "*"},
		{Group: "*", Version: "*", Kind: gvk.Kind},
		{Group: "*", Version: gvk.Version, Kind: "*"},
		{Group: "*", Version: gvk.Version, Kind: gvk.Kind},
		{Group: gvk.Group, Version: "*", Kind: "*"},
		{Group: gvk.Group, Version: "*", Kind: gvk.Kind},
		{Group: gvk.Group, Version: gvk.Version, Kind: "*"},
		{Group: gvk.Group, Version: gvk.Version, Kind: gvk.Kind},
	}
	return o
}

type Object struct {
	*unstructured.Unstructured
	matchKeys []schema.GroupVersionKind
	mu        sync.RWMutex

	events []Event
}

func (o *Object) HasLabels(labels map[string]string) bool {
	l := o.GetLabels()
	for key, value := range labels {
		if v, ok := l[key]; !ok || v != value {
			return false
		}
	}
	return true
}

func (o *Object) HasAnnotations(annotations map[string]string) bool {
	ann := o.GetAnnotations()
	for key, value := range annotations {
		if v, ok := ann[key]; !ok || v != value {
			return false
		}
	}
	return true
}

func (o *Object) Matches(opts ...MatchOption) bool {
	if len(opts) == 0 {
		return true
	}
	opt := newOptions(opts...)
	if opt.kinds.HasAny(o.matchKeys...) && opt.names.HasAny("*", o.GetName()) && opt.namespaces.HasAny(o.GetNamespace(), "*") {
		if len(opt.matchLabels) == 0 || o.HasLabels(opt.matchLabels) {
			if len(opt.matchAnnotations) == 0 || o.HasAnnotations(opt.matchAnnotations) {
				return true
			}
		}
	}
	return false
}

func (o *Object) Diff(o2 *Object) string {
	return cmp.Diff(o.Object, o2.Object)
}

func (o *Object) Equals(o2 *Object) bool {
	return reflect.DeepEqual(o.Object, o2.Object)
}

func (o *Object) AddLabels(labels map[string]string) {
	l := o.GetLabels()
	if l == nil {
		l = make(map[string]string)
	}
	for key, value := range labels {
		l[key] = value
	}
	o.SetLabels(l)
}

func (o *Object) AddAnnotations(annotations map[string]string) {
	ann := o.GetAnnotations()
	if ann == nil {
		ann = make(map[string]string)
	}
	for key, value := range annotations {
		ann[key] = value
	}
	o.SetAnnotations(ann)
}

func (o *Object) AddFinalizers(finalizers []string) {
	o.SetFinalizers(append(o.GetFinalizers(), finalizers...))
}

func (o *Object) SetNamePrefix(fix string) {
	o.SetName(fix + o.GetName())
}

func (o *Object) SetNameSuffix(fix string) {
	o.SetName(o.GetName() + fix)
}

func (o *Object) StrategicMergePatch(patch map[string]any) error {
	// structs registered with the scheme will have
	// the strategicMergePatch tags defined on the struct
	d, err := scheme.Scheme.New(o.GroupVersionKind())
	if err != nil {
		return &errors.ErrPatchStrategicMergeUnregisteredSchema{
			GroupVersionKind: o.GroupVersionKind(),
			Name:             o.GetName(),
			Namespace:        o.GetNamespace(),
			Patch:            patch,
		}
	}

	// I'll need to write this logic myself
	obj, err := strategicpatch.StrategicMergeMapPatch(o.UnstructuredContent(), patch, d)
	if err != nil {
		return &errors.ErrPatchStrategicMerge{
			GroupVersionKind: o.GroupVersionKind(),
			Name:             o.GetName(),
			Namespace:        o.GetNamespace(),
			Resource:         o.UnstructuredContent(),
			Err:              err,
			Patch:            patch,
		}
	}
	o.SetUnstructuredContent(obj)
	return nil
}

func (o *Object) MergePatch(patch map[string]any) error {

	uc := o.UnstructuredContent()
	if err := mergo.Merge(uc, patch, mergo.WithOverride); err != nil {
		return &errors.ErrMergePatch{
			GroupVersionKind: o.GroupVersionKind(),
			Name:             o.GetName(),
			Namespace:        o.GetNamespace(),
			Resource:         o.UnstructuredContent(),
			Err:              err,
			Patch:            patch,
		}
	}
	o.SetUnstructuredContent(uc)
	return nil
}

func (o *Object) FieldPatch(fp *fieldpath.FieldPath, value any) error {
	return fp.SetValue(o.Object, value)
}

func (o *Object) JSONPatch(patch jsonpatch.Patch) error {
	doc, err := json.Marshal(o.Object)
	if err != nil {
		return err
	}
	doc, err = patch.Apply(doc)
	if err != nil {
		return err
	}
	return json.Unmarshal(doc, &o.Object)
}

// Record tracks an event that occurred on the resource,
// such as a mutation
func (o *Object) Record(event Event) {
	o.events = append(o.events, event)
}
