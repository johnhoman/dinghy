package mutate

import "github.com/johnhoman/dinghy/internal/visitor"

type NewConfigFunc func() any

func newAnyMap() any {
	return make(map[string]any)
}

func newStringMap() any {
	return make(map[string]string)
}

func newConfig[T any]() NewConfigFunc {
	return func() any {
		t := new(T)
		return t
	}
}

func newNamespaceConfig() any {
	return &visitor.NamespaceConfig{}
}

func newNameConfig() any {
	return &visitor.NameConfig{}
}
