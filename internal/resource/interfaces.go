package resource

var (
	_ Tree = &treeNode{}
)

type Tree interface {
	Visit(visitor Visitor, opts ...MatchOption) error
	// Insert an Object into the tree
	Insert(obj *Object) error
	// Pop returns the resource associated with the resource key and removes.
	// the resource from the tree. ErrNotFound will be returned if the key
	// isn't found
	Pop(key Key) (*Object, error)
}

// NewTree returns a new tree for managing Kubernetes resource
// manifests.
func NewTree() Tree {
	return &treeNode{nodes: make(map[string]*treeNode)}
}

// ParseKey creates a resource key from the provided resource. The
// resource key uniquely identifies a resource in a resource Tree. Keys
// can be used to query toe Tree using Tree.Pop()
func ParseKey(obj *Object) Key {
	return Key{
		GroupVersion: obj.GetAPIVersion(),
		Kind:         obj.GetKind(),
		Name:         obj.GetName(),
		Namespace:    obj.GetNamespace(),
	}
}

// GetResource returns a resource from the tree without changing the state of
// the Tree.
func GetResource(tree Tree, key Key) (*Object, error) {
	obj, err := tree.Pop(key)
	if err != nil {
		return nil, err
	}
	return obj, tree.Insert(obj)
}
