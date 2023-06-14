package mutate

import (
	"github.com/johnhoman/dinghy/internal/visitor"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func AddAnnotations(config any) (visitor.Visitor, error) {
	m, ok := config.(*map[string]string)
	if !ok {
		return nil, ErrTypedConfig
	}
	return visitor.AddAnnotations(*m), nil
}

func SetAnnotations(config any) (visitor.Visitor, error) {
	m, ok := config.(*map[string]string)
	if !ok {
		return nil, ErrTypedConfig
	}
	return visitor.SetAnnotations(*m), nil
}

func Name(config any) (visitor.Visitor, error) {
	c, ok := config.(*visitor.NameConfig)
	if !ok {
		return nil, ErrTypedConfig
	}
	return visitor.Name(*c), nil
}

func Namespace(config any) (visitor.Visitor, error) {
	c, ok := config.(*visitor.NamespaceConfig)
	if !ok {
		return nil, ErrTypedConfig
	}
	return visitor.Namespace(*c), nil
}

func Metadata(patch any) (visitor.Visitor, error) {
	return StrategicMergePatch(map[string]any{"metadata": patch})
}

func Labels(config any) (visitor.Visitor, error) {
	m, ok := config.(*map[string]string)
	if !ok {
		return nil, ErrTypedConfig
	}
	if len(*m) == 0 {
		return visitor.Nop(), nil
	}
	return visitor.Func(func(obj *unstructured.Unstructured) error {
		l := obj.GetLabels()
		if l == nil {
			l = make(map[string]string)
		}
		for key, value := range *m {
			l[key] = value
		}
		obj.SetLabels(l)
		return nil
	}), nil
}
