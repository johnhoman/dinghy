package resource

import (
	"github.com/johnhoman/dinghy/internal/codec"
	"github.com/johnhoman/dinghy/internal/visitor"
	"io"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func CopyTree(to, from Tree) error {
	return from.Visit(copyTree(to))
}

func copyTree(to Tree) visitor.Visitor {
	return visitor.Func(func(obj *unstructured.Unstructured) error {
		return to.Insert(obj)
	})
}

func printTree(e codec.Encoder) visitor.Visitor {
	return visitor.Func(func(obj *unstructured.Unstructured) error {
		return e.Encode(obj.Object)
	})
}

func PrintTree(tree Tree, w io.Writer) error {
	return tree.Visit(printTree(codec.YAMLEncoder(w)))
}
