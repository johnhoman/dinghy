package resource

import (
	goerr "errors"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	ErrResourceConflict = errors.New("cannot overwrite existing resource resource")
	ErrNotFound         = errors.New("the requested resource was not found")
	ErrInsertResource   = errors.New("failed to insert resource into tree")
	ErrPopResource      = errors.New("failed to pop resource from tree")
)

// treeNode stores resources in a hierarchy based on the resource
// group, version, Kind, Namespace, and Name to more easily search
// for resources on selection.
type treeNode struct {
	nodes map[string]*treeNode
	obj   *Object
}

func (tree *treeNode) Visit(visitor Visitor, opts ...MatchOption) error {
	o := newOptions(opts...)
	// the replacer visitor resets a node in the tree if
	// any of the identifying information changes
	visitor = replaceVisitor(tree, visitor)
	if len(o.matchLabels) > 0 {
		visitor = matchLabels(o.matchLabels, visitor)
	}
	if len(o.matchAnnotations) > 0 {
		visitor = matchAnnotations(o.matchAnnotations, visitor)
	}

	if o.kinds.Len() == 0 {
		o.kinds.Insert(schema.GroupVersionKind{Group: "*", Version: "*", Kind: "*"})
	}

	if o.namespaces.Len() == 0 {
		o.namespaces.Insert("*")
	}

	if o.names.Len() == 0 {
		o.names.Insert("*")
	}

	errs := make([]error, 0)
	for gvk := range o.kinds {
		for namespace := range o.namespaces {
			for name := range o.names {
				path := []string{gvk.Group, gvk.Version, gvk.Kind, namespace, name}
				if err := tree.visit(visitor, path...); err != nil {
					errs = append(errs, err)
				}
			}
		}
	}
	if len(errs) > 0 {
		return goerr.Join(errs...)
	}
	return nil
}

func (tree *treeNode) path(key Key) []string {
	gv, err := schema.ParseGroupVersion(key.GroupVersion)
	if err != nil {
		panic(err)
	}
	return []string{gv.Group, gv.Version, key.Kind, key.Namespace, key.Name}
}

// Insert a resource into the treeNode. If a resource already exists in the tree
// and the resource content is different, and error will be returned.
func (tree *treeNode) Insert(obj *Object) error {
	err := tree.insert(obj, tree.path(newResourceKey(obj))...)
	return errors.Wrapf(err, "%s: %s", ErrInsertResource, treeError(obj))
}

// Pop finds, removes, and returns a resource from the treeNode. If the resource
// doesn't exist, ErrNotFound will be returned
func (tree *treeNode) Pop(key Key) (*Object, error) {
	obj, err := tree.pop(tree.path(key)...)
	return obj, errors.Wrapf(err, "%s: %s", ErrPopResource, key.String())
}

func (tree *treeNode) visitNodes(visitor Visitor, path ...string) error {
	errs := make([]error, 0)
	for _, node := range tree.nodes {
		if err := node.visit(visitor, path...); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return goerr.Join(errs...)
	}
	return nil
}

func (tree *treeNode) insert(obj *Object, path ...string) error {
	if len(path) == 0 {
		if tree.obj != nil && !reflect.DeepEqual(obj, tree.obj) {
			return ErrResourceConflict
		}
		tree.obj = obj
		return nil
	}
	zero, path := strings.ToLower(path[0]), path[1:]
	if _, ok := tree.nodes[zero]; !ok {
		tree.nodes[zero] = &treeNode{nodes: make(map[string]*treeNode)}
	}
	return tree.nodes[zero].insert(obj, path...)
}

func (tree *treeNode) empty() bool {
	return len(tree.nodes) == 0
}

// pop removes and returns the child element at path if it exists. If
// it doesn't exit ErrNotFound is returned
func (tree *treeNode) pop(path ...string) (*Object, error) {
	if len(path) == 0 {
		if tree.obj == nil {
			return nil, ErrNotFound
		}
		obj := tree.obj
		tree.obj = nil
		return obj, nil
	}
	// normalize the path (strings.ToLower) during insertion to make
	// queries case-insensitive
	zero, path := strings.ToLower(path[0]), path[1:]
	t, ok := tree.nodes[zero]
	if !ok {
		return nil, ErrNotFound
	}
	defer func() {
		if t.empty() {
			// cleanup the child nodes if they're empty
			delete(tree.nodes, zero)
		}
	}()
	obj, err := t.pop(path...)
	if err != nil {
		// don't wrap this error, it'll get wrapped at the top of
		// the tree
		return nil, err
	}
	return obj, nil
}

func (tree *treeNode) visit(visitor Visitor, path ...string) error {
	if len(path) == 0 {
		if tree.obj == nil {
			return nil
		}
		return visitor.Visit(tree.obj)
	}
	zero, path := strings.ToLower(path[0]), path[1:]
	if zero == "*" {
		// matches on all nodes in the tree
		return tree.visitNodes(visitor, path...)
	}
	t, ok := tree.nodes[zero]
	if !ok {
		return nil
	}
	return t.visit(visitor, path...)
}

func newOptions(opts ...MatchOption) *matchOptions {
	o := &matchOptions{
		kinds:            sets.New[schema.GroupVersionKind](),
		names:            sets.New[string](),
		namespaces:       sets.New[string](),
		matchLabels:      make(map[string]string),
		matchAnnotations: make(map[string]string),
	}
	for _, f := range opts {
		f(o)
	}
	return o
}

func treeError(obj *Object) string {
	return newResourceKey(obj).String()
}
