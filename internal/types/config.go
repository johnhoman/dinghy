package types

import (
	"sort"
)

const (
	GroupName  = "dinghy.dev"
	Version    = "v1alpha1"
	ConfigKing = "Config"
)

type FieldRef struct {
	FieldPath string `yaml:"fieldPath"`
}

// ResourceSelector selects resources based on attributes of the resource,
// such as labels, annotations.
type ResourceSelector struct {
	MatchLabels map[string]string `yaml:"matchLabels"`
	Kinds       []string          `yaml:"kinds"`
	Names       []string          `yaml:"names"`
	Namespaces  []string          `yaml:"namespaces"`
}

// GeneratorSpec is a spec for resource generation rules.
type GeneratorSpec struct {
	// Name is a unique name for the mutation
	Name string `yaml:"name"`
	// Uses is the name or path to the plugin
	Uses string `yaml:"uses" dinghy:"required"`
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
	Uses string `yaml:"uses" dinghy:"required"`
	With any    `yaml:"with"`
}

type (
	MutationSpec   = PluginSpec
	ValidationSpec = PluginSpec
)

type Config struct {
	APIVersion string `yaml:"apiVersion" dinghy:"required"`
	Kind       string `yaml:"kind" dinghy:"required"`

	Resources   []string         `yaml:"resources"`
	Overlays    []string         `yaml:"overlays"`
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
	resources = append(resources, c.Resources...)
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
