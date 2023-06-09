package kustomize

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"reflect"
	"sort"
)

const (
	GroupName  = "kustomize.config.k8s.io"
	Version    = "v1beta1"
	ConfigKing = "Kustomization"
)

type FieldPath string

func (fp FieldPath) SetValue(obj *unstructured.Unstructured, value any) error {
	l := newFieldPathLexer(string(fp))
	tokens, err := l.parseTokens()
	if err != nil {
		return err
	}
	_ = tokens
	m := obj.UnstructuredContent()
	obj.SetUnstructuredContent(m)
	return nil
}

func (fp FieldPath) GetValue(obj *unstructured.Unstructured) (any, error) {
	return nil, nil
}

type MatchExpression struct{}
type FieldRef struct {
	FieldPath FieldPath `yaml:"fieldPath"`
}

// ResourceSelector selects resources based on attributes of the resource,
// such as labels, annotations, or fields
type ResourceSelector struct {
	MatchLabels           map[string]string `yaml:"matchLabels"`
	MatchLabelExpressions []MatchExpression `yaml:"matchExpressions"`
	MatchAnnotations      map[string]string `yaml:"matchAnnotations"`
	MatchFields           []FieldRef        `yaml:"matchFields"`
	Kinds                 []string          `yaml:"kinds"`
}

func (s ResourceSelector) Matches(obj *unstructured.Unstructured) bool {
	if reflect.DeepEqual(s, ResourceSelector{}) {
		return true
	}
	return false
}

// MutationSpec is a spec for resource mutation rules.
type MutationSpec struct {
	// Name is a unique name for the mutation
	Name string `yaml:"name"`
	// Selector selects resources to apply the mutation to
	Selector ResourceSelector `yaml:"selector"`
	// Uses is the name or path to the plugin
	Uses string         `yaml:"uses"`
	With map[string]any `yaml:"with"`
}

func (m MutationSpec) Matches(obj *unstructured.Unstructured) bool {
	return m.Selector.Matches(obj)
}

type Validation struct {
	Selector ResourceSelector `yaml:"selector"`
	// Uses is the name of the plugin uses for the mutation.
	Uses string            `yaml:"uses"`
	With map[string]string `yaml:"with"`
}

type Config struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`

	Resources   []string       `yaml:"resources"`
	Mutations   []MutationSpec `yaml:"mutate"`
	Validations []Validation   `yaml:"validate"`
}

// AddResource to the config file. If the resource already exists
// in the config, it won't be added, otherwise, it will be appended
// to the end of the resource list
func (c *Config) AddResource(resource string) {
	if c.Resources == nil {
		c.Resources = []string{resource}
		return
	}
	for _, item := range c.Resources {
		if item == resource {
			return
		}
	}
	c.Resources = append(c.Resources, resource)
}

// GetResources returns a list of resources included in the config
// in sorted order.
func (c *Config) GetResources() []string {
	resources := make([]string, 0, len(c.Resources))
	for _, item := range c.Resources {
		resources = append(resources, item)
	}
	sort.Strings(resources)
	return resources
}

// SetResources resets the resources included in the config to the
// provided list
func (c *Config) SetResources(r []string) { c.Resources = r }

// NewConfig creates and returns a new config file with the provided
// options applied.
func NewConfig(opts ...ConfigOption) *Config {
	c := &Config{APIVersion: GroupName + "/" + Version, Kind: ConfigKing}
	for _, f := range opts {
		f(c)
	}
	return c
}

type ConfigOption func(o *Config)

// WithResource adds a resource to the Configuration
func WithResource(path string) ConfigOption {
	return func(o *Config) {
		o.AddResource(path)
	}
}
