package mutate

import (
	"gopkg.in/yaml.v3"

	"github.com/johnhoman/dinghy/internal/resource"
)

func SideEffect(visitor SideEffectVisitor, tree resource.Tree) resource.Visitor {
	return &sideEffectVisitor{tree: tree, visitor: visitor}
}

type SideEffectVisitor interface {
	SideEffect(oldObj *resource.Object, tree resource.Tree) error
	resource.Visitor
}

type sideEffectVisitor struct {
	tree    resource.Tree
	visitor SideEffectVisitor
}

// Visit copies the object before running the visitor, then gives the original
// and the resource tree to the SideEffect visitor
func (se *sideEffectVisitor) Visit(obj *resource.Object) error {
	var m map[string]any
	if err := copyResource(&m, obj.Object); err != nil {
		return err
	}
	if err := se.visitor.Visit(obj); err != nil {
		return nil
	}
	objBefore := resource.Unstructured(m)
	return se.visitor.SideEffect(objBefore, se.tree)
}

func copyResource(to any, from any) error {
	data, err := yaml.Marshal(from)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, to)
}
