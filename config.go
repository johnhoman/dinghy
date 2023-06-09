package kustomize

import "sort"

const (
	GroupName  = "kustomize.config.k8s.io"
	Version    = "v1beta1"
	ConfigKing = "Kustomization"
)

type Config struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`

	Resources []string `yaml:"resources"`
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
