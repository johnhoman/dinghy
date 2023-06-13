package visitor

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

type NameConfig struct {
	Prefix string `yaml:"prefix"`
	Suffix string `yaml:"suffix"`
}

func Name(c NameConfig) Visitor {
	return Func(func(obj *unstructured.Unstructured) error {
		obj.SetName(c.Prefix + obj.GetName() + c.Suffix)
		return nil
	})
}

type NamespaceConfig struct {
	Namespace string `yaml:"namespace" dinghy:"required"`
}

func Namespace(c NamespaceConfig) Visitor {
	return Func(func(obj *unstructured.Unstructured) error {
		obj.SetNamespace(c.Namespace)
		return nil
	})
}

// AddAnnotation is a visitor that adds a single key value pair
// to a resource's annotations.
func AddAnnotation(key, value string) Visitor {
	return AddAnnotations(map[string]string{key: value})
}

// RemoveAnnotation is a visitor that adds a single key value pair
// to a resource's annotations.
func RemoveAnnotation(key string) Visitor {
	return Func(func(obj *unstructured.Unstructured) error {
		annotations := obj.GetAnnotations()
		delete(annotations, key)
		obj.SetAnnotations(annotations)
		return nil
	})
}

// AddAnnotations is a visitor that adds a map of annotations to a resource's
// annotations
func AddAnnotations(in map[string]string) Visitor {
	return Func(func(obj *unstructured.Unstructured) error {
		annotations := obj.GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}
		for key, value := range in {
			annotations[key] = value
		}
		obj.SetAnnotations(annotations)
		return nil

	})
}

// SetAnnotations is a visitor that replaces the incoming resource's
// annotations with the provided annotations
func SetAnnotations(in map[string]string) Visitor {
	return Func(func(obj *unstructured.Unstructured) error {
		obj.SetAnnotations(in)
		return nil
	})
}
