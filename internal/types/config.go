package types

import (
	"k8s.io/apimachinery/pkg/labels"
	"sort"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	GroupName  = "dinghy.dev"
	Version    = "v1beta1"
	ConfigKing = "Config"
)

type FieldRef struct {
	FieldPath string `yaml:"fieldPath"`
}

// ResourceSelector selects resources based on attributes of the resource,
// such as labels, annotations.
type ResourceSelector struct {
	MatchLabels      map[string]string `yaml:"matchLabels"`
	MatchAnnotations map[string]string `yaml:"matchAnnotations"`
	MatchFields      []FieldRef        `yaml:"matchFields"`
	Kinds            []string          `yaml:"kinds"`
}

func (s ResourceSelector) Matcher() func(obj *unstructured.Unstructured) bool {
	kinds := sets.New[string](s.Kinds...)
	matchLabels := labels.SelectorFromSet(s.MatchLabels)
	matchAnnotations := labels.SelectorFromSet(s.MatchAnnotations)

	return func(obj *unstructured.Unstructured) bool {
		if kinds.Len() > 0 && !kinds.Has(obj.GetObjectKind().GroupVersionKind().Kind) {
			return false
		}
		if !matchAnnotations.Empty() && !matchAnnotations.Matches(labels.Set(obj.GetAnnotations())) {
			return false
		}
		if !matchLabels.Empty() && !matchLabels.Matches(labels.Set(obj.GetLabels())) {
			return false
		}
		return true
	}
}

func (s ResourceSelector) Matches(obj *unstructured.Unstructured) bool {
	return false
}

// GeneratorSpec is a spec for resource generation rules.
type GeneratorSpec struct {
	// Name is a unique name for the mutation
	Name string `yaml:"name"`
	// Uses is the name or path to the plugin
	Uses string `yaml:"uses"`
	With any    `yaml:"with"`
}

// PluginSpec is a spec for resource mutation rules.
type PluginSpec struct {
	// Name is a unique name for the mutation
	Name string `yaml:"name"`
	// Selector selects resources to apply the mutation to. If omitted,
	// all resources will be selected.
	Selector ResourceSelector `yaml:"selector"`
	// Uses is the name or path to the plugin
	Uses string `yaml:"uses" bespoke:"required"`
	With any    `yaml:"with"`
}

type (
	MutationSpec   = PluginSpec
	ValidationSpec = PluginSpec
)

func (m PluginSpec) Matches(obj *unstructured.Unstructured) bool {
	return m.Selector.Matches(obj)
}

func (m PluginSpec) Matcher() func(obj *unstructured.Unstructured) bool {
	return m.Selector.Matcher()
}

type Config struct {
	APIVersion string `yaml:"apiVersion" bespoke:"required"`
	Kind       string `yaml:"kind" bespoke:"required"`

	Resources   []string         `yaml:"resources"`
	Generators  []GeneratorSpec  `yaml:"generate"`
	Mutations   []MutationSpec   `yaml:"mutate"`
	Validations []ValidationSpec `yaml:"validate"`
}

// AddResource to the config file. If the resource already exists
// fieldPath the config, it won't be added, otherwise, it will be appended
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

// GetResources returns a list of resources included fieldPath the config
// fieldPath sorted order.
func (c *Config) GetResources() []string {
	resources := make([]string, 0, len(c.Resources))
	for _, item := range c.Resources {
		resources = append(resources, item)
	}
	sort.Strings(resources)
	return resources
}

// SetResources resets the resources included fieldPath the config to the
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
