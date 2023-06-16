package mutate

import (
	"github.com/johnhoman/dinghy/internal/codec"
	"github.com/johnhoman/dinghy/internal/resource"
	"github.com/johnhoman/dinghy/internal/visitor"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func SideEffect(visitor SideEffectVisitor, tree resource.Tree) visitor.Visitor {
	return &sideEffectVisitor{tree: tree, visitor: visitor}
}

type SideEffectVisitor interface {
	SideEffect(oldObj *unstructured.Unstructured, tree resource.Tree) error
	visitor.Visitor
}

type sideEffectVisitor struct {
	tree    resource.Tree
	visitor SideEffectVisitor
}

// Visit copies the object before running the visitor, then gives the original
// and the resource tree to the SideEffect visitor
func (se *sideEffectVisitor) Visit(obj *unstructured.Unstructured) error {
	var m map[string]any
	if err := codec.YAMLCopyTo(&m, obj.Object); err != nil {
		return err
	}
	if err := se.visitor.Visit(obj); err != nil {
		return nil
	}
	objBefore := &unstructured.Unstructured{Object: m}
	return se.visitor.SideEffect(objBefore, se.tree)
}
