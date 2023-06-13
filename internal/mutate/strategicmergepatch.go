package mutate

import "github.com/johnhoman/dinghy/internal/visitor"

func StrategicMergePatch(config any) (visitor.Visitor, error) {
	patch, ok := config.(map[string]any)
	if !ok {
		return nil, ErrTypedConfig
	}
	return visitor.StrategicMergePatch(patch), nil
}
