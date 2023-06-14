package mutate

import (
	"github.com/johnhoman/dinghy/internal/fieldpath"
	"github.com/johnhoman/dinghy/internal/visitor"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type patchConfig struct {
	FieldPath  string   `yaml:"fieldPath"`
	FieldPaths []string `yaml:"fieldPaths"`
	Value      any      `yaml:"value" dinghy:"required"`
}

func Patch(config any) (visitor.Visitor, error) {
	c, ok := config.(*patchConfig)
	if !ok {
		return nil, ErrTypedConfig
	}
	if c.FieldPath != "" {
		c.FieldPaths = append(c.FieldPaths, c.FieldPath)
	}
	if len(c.FieldPaths) == 0 {
		return nil, errors.Wrap(ErrTypedConfig, "missing required field fieldPath")
	}
	fpList := make([]*fieldpath.FieldPath, 0)
	for _, fieldPath := range c.FieldPaths {
		fp, err := fieldpath.Parse(fieldPath)
		if err != nil {
			return nil, err
		}
		fpList = append(fpList, fp)
	}
	return visitor.Func(func(obj *unstructured.Unstructured) error {
		for _, fp := range fpList {
			if err := fp.SetValue(obj.Object, c.Value); err != nil {
				return err
			}
		}
		return nil
	}), nil
}
