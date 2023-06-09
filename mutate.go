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

type mutatePrependNamePrefixConfig struct {
	NamePrefix *string `yaml:"namePrefix"`
}

func (c *mutatePrependNamePrefixConfig) validate() error {
	if c.NamePrefix == nil {
		return errors.Errorf("missing required argument %q", "namePrefix")
	}
	return nil
}

type mutateAppendNameSuffixConfig struct {
	NameSuffix *string `yaml:"nameSuffix"`
}

func (c *mutateAppendNameSuffixConfig) validate() error {
	if c.NameSuffix == nil {
		return errors.Errorf("missing required argument %q", "nameSuffix")
	}
	return nil
}

type mutateReplaceNamespaceConfig struct {
	Namespace *string `yaml:"namespace"`
}

func (c *mutateReplaceNamespaceConfig) validate() error {
	if c.Namespace == nil {
		return errors.Errorf("missing required argument %q", "namespace")
	}
	return nil
}

type mutatePatchConfig struct {
	FieldPath *string `yaml:"fieldPath"`
	Value     *string `yaml:"value"`
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
	mutatePrependNamePrefix = MutatorFunc(func(obj *unstructured.Unstructured, _ ResourceMap, config any) error {
		c, ok := config.(*mutatePrependNamePrefixConfig)
		if !ok {
			return ErrMutatorConfig
		}
		obj.SetName(*c.NamePrefix + obj.GetName())
		return nil
	})

	mutateAppendNameSuffix = MutatorFunc(func(obj *unstructured.Unstructured, _ ResourceMap, config any) error {
		c, ok := config.(*mutateAppendNameSuffixConfig)
		if !ok {
			return ErrMutatorConfig
		}
		obj.SetName(obj.GetName() + *c.NameSuffix)
		return nil
	})

	mutateReplaceNamespace = MutatorFunc(func(obj *unstructured.Unstructured, _ ResourceMap, config any) error {
		c, ok := config.(*mutateReplaceNamespaceConfig)
		if !ok {
			return ErrMutatorConfig
		}
		obj.SetNamespace(*c.Namespace)
		return nil
	})

	mutatePatch = MutatorFunc(func(obj *unstructured.Unstructured, _ ResourceMap, config any) error {
		_, ok := config.(*mutatePatchConfig)
		if !ok {
			return ErrMutatorConfig
		}

		// You can't set a value with jmes path
		m := obj.UnstructuredContent()
		obj.SetUnstructuredContent(m)
		return nil
	})
)

func init() {
	RegisterMutate("builtin.kustomize.k8s.io/prependNamePrefix", mutatePrependNamePrefix, func() any {
		return &mutatePrependNamePrefixConfig{}
	})

	RegisterMutate("builtin.kustomize.k8s.io/appendNameSuffix", mutateAppendNameSuffix, func() any {
		return &mutateAppendNameSuffixConfig{}
	})

	RegisterMutate("builtin.kustomize.k8s.io/replaceNamespace", mutateReplaceNamespace, func() any {
		return &mutateReplaceNamespaceConfig{}
	})

	RegisterMutate("builtin.kustomize.k8s.io/replaceNamespace", mutateReplaceNamespace, func() any {
		return &mutateReplaceNamespaceConfig{}
	})

	RegisterMutate("builtin.kustomize.k8s.io/patch", mutatePatch, func() any {
		return &mutatePatchConfig{}
	})
}
