package mutate

import (
	"github.com/johnhoman/dinghy/internal/visitor"
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
