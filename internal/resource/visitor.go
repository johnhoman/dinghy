package resource

import (
	"gopkg.in/yaml.v3"
	"io"
	"k8s.io/apimachinery/pkg/labels"
)

func CopyTree(to, from Tree) error {
	return from.Visit(copyTree(to))
}

func PrintTree(tree Tree, w io.Writer) error {
	return tree.Visit(printTree(yaml.NewEncoder(w)))
}

func copyTree(to Tree) Visitor {
	return VisitorFunc(func(obj *Object) error {
		return to.Insert(obj)
	})
}

func printTree(e *yaml.Encoder) Visitor {
	return VisitorFunc(func(obj *Object) error {
		return e.Encode(obj.Object)
	})
}

// matchLabels is a visitor predicate that only runs the next visitor if the resource
// matches the provided labels.
func matchLabels(l map[string]string, next Visitor) Visitor {
	return VisitorFunc(func(obj *Object) error {
		set := labels.Set(l)
		selector := labels.SelectorFromSet(set)
		if selector.Matches(labels.Set(obj.GetLabels())) {
			return next.Visit(obj)
		}
		return nil
	})
}

// matchLabels is a visitor predicate that only runs the next visitor if the resource
// matches the provided labels.
func matchAnnotations(l map[string]string, next Visitor) Visitor {
	return VisitorFunc(func(obj *Object) error {
		set := labels.Set(l)
		selector := labels.SelectorFromSet(set)
		if selector.Matches(labels.Set(obj.GetAnnotations())) {
			return next.Visit(obj)
		}
		return nil
	})
}

// replaceVisitor wraps a visitor and checks to see if the visitor changed the key. If the
// visitor changed the key, it reinserts the resource in the tree
func replaceVisitor(tree *treeNode, next Visitor) Visitor {
	return VisitorFunc(func(obj *Object) error {
		key := newResourceKey(obj)
		// visit the node
		if err := next.Visit(obj); err != nil {
			return err
		}
		if newResourceKey(obj) != key {
			// the resource key changed, so we need to reinsert it into the
			// tree, so that it can be found again
			u, err := tree.Pop(key)
			if err != nil {
				return err
			}
			return tree.Insert(u)
		}
		return nil
	})
}
