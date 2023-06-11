package kustomize

import (
	goerr "errors"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var ErrMutatorConfig = errors.New("the provided mutator config is the wrong kind")

// Mutator applies a mutation to a resource.
type Mutator interface {
	Apply(obj *unstructured.Unstructured, rm ResourceMap, config any) error
}

type validator interface {
	validate() error
}

type MutatorFunc func(obj *unstructured.Unstructured, rm ResourceMap, config any) error

func (fn MutatorFunc) Apply(obj *unstructured.Unstructured, rm ResourceMap, config any) error {
	v, ok := config.(validator)
	if ok {
		if err := v.validate(); err != nil {
			return goerr.Join(ErrMutatorConfig, err)
		}
	}
	return fn(obj, rm, config)
}

type mutateNameConfig struct {
	Prefix string `yaml:"Prefix"`
	Suffix string `yaml:"Suffix"`
}

func (c *mutateNameConfig) validate() error {
	return nil
}

type mutateNamespaceConfig struct {
	Prefix  string  `yaml:"prefix"`
	Suffix  string  `yaml:"suffix"`
	Replace *string `yaml:"replace"`
}

func (c *mutateNamespaceConfig) validate() error {
	return nil
}

type mutatePatchConfig struct {
	FieldPath *string `yaml:"fieldPath"`
	Value     any     `yaml:"value"`
}

func (c *mutatePatchConfig) validate() error {
	if c.FieldPath == nil {
		return errors.Errorf("missing required argument %q", "fieldPath")
	}
	if c.Value == nil {
		return errors.Errorf("missing required argument %q", "value")
	}
	return nil
}

var (
	mutateName = MutatorFunc(func(obj *unstructured.Unstructured, _ ResourceMap, config any) error {
		c, ok := config.(*mutateNameConfig)
		if !ok {
			return ErrMutatorConfig
		}
		obj.SetName(c.Prefix + obj.GetName() + c.Suffix)
		return nil
	})

	mutateNamespace = MutatorFunc(func(obj *unstructured.Unstructured, _ ResourceMap, config any) error {
		c, ok := config.(*mutateNamespaceConfig)
		if !ok {
			return ErrMutatorConfig
		}
		namespace := obj.GetNamespace()
		if c.Replace != nil {
			namespace = *c.Replace
		}
		obj.SetNamespace(c.Prefix + namespace + c.Suffix)
		return nil
	})

	mutatePatch = MutatorFunc(func(obj *unstructured.Unstructured, _ ResourceMap, config any) error {
		c, ok := config.(*mutatePatchConfig)
		if !ok {
			return ErrMutatorConfig
		}

		// You can't set a value with jmes path
		fp, err := NewFieldPath(*c.FieldPath)
		if err != nil {
			return err
		}
		return fp.SetValue(obj.Object, c.Value)
	})
)

func init() {
	RegisterMutate("builtin.kustomize.k8s.io/name", mutateName, func() any {
		return &mutateNameConfig{}
	})

	RegisterMutate("builtin.kustomize.k8s.io/namespace", mutateNamespace, func() any {
		return &mutateNamespaceConfig{}
	})

	RegisterMutate("builtin.kustomize.k8s.io/patch", mutatePatch, func() any {
		return &mutatePatchConfig{}
	})
}
