package visitor

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Visitor interface {
	Visit(obj *unstructured.Unstructured) error
}

type Func func(obj *unstructured.Unstructured) error

func (f Func) Visit(obj *unstructured.Unstructured) error {
	return f(obj)
}

func NewVisitorChain(visitors ...Visitor) Visitor {
	return Chain(visitors)
}

type Chain []Visitor

func (chain Chain) Visit(obj *unstructured.Unstructured) error {
	for _, visitor := range chain {
		if err := visitor.Visit(obj); err != nil {
			return err
		}
	}
	return nil
}

func Visit(visitor Visitor, m map[string]any) error {
	return visitor.Visit(&unstructured.Unstructured{Object: m})
}

func Nop() Visitor {
	return Func(func(_ *unstructured.Unstructured) error {
		return nil
	})
}
