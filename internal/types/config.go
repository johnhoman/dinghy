package types

import (
	"fmt"
	"github.com/johnhoman/dinghy/internal/codec"
	"github.com/johnhoman/dinghy/internal/generate"
	"github.com/johnhoman/dinghy/internal/mutate"
	"gopkg.in/yaml.v3"
	"sort"
)

const (
	GroupName  = "dinghy.dev"
	Version    = "v1alpha1"
	ConfigKing = "Config"
)

type Path string

type FieldRef struct {
	FieldPath string `yaml:"fieldPath"`
}

// A Module is an external javascript file that can be loaded
// as either a generator, validator, or mutator. Modules cannot
// have any external dependencies. If Checksum is provided, the sha256 sum
// of the file will be computed on download and will be compared with
// the provided sum.
type Module struct {
	// Checksum is the sha256 sum of the referenced content. Checksum
	// will be used for comparison after download.
	Checksum string `yaml:"checksum"`
	// The Name attribute will be used for component/plugin registration. If
	// the provided name is `foo-service` for example, then a generator can
	// reference it by that name in the generate section of the dinghyfile.
	Name string `yaml:"name"`
	// The Source is the path to the source. Most likely this will be
	// a GitHub URL, but it can also be a local path. Source is the only
	// required attribute, but it is strongly encouraged to also provide
	// a checksum.
	Source Path `yaml:"source"`
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
	Uses string `yaml:"uses"`
	With any    `yaml:"with"`
}

type (
	MutationSpec   = PluginSpec
	ValidationSpec = PluginSpec
)

var (
	_ codec.Validator = &Config{}
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

func (c *Config) Validate() []string {

	errs := make([]string, 0)
	for k, gen := range c.Generators {
		if gen.Uses == "" {
			errs = append(errs, fmt.Sprintf("generate[%d]: %q is a required field", k, "uses"))
			continue
		}
		if !generate.Has(gen.Uses) {
			errs = append(errs, fmt.Sprintf("generate[%d]: %q does not exist", k, gen.Uses))
			continue
		}
		vis, err := generate.Get(gen.Uses)
		if err != nil {
			panic("BUG: why is this happening. I just made sure it existed")
		}
		in, err := yaml.Marshal(gen.With)
		if err != nil {
			panic("BUG: this was just decoded")
		}
		if err := yaml.Unmarshal(in, vis); err != nil {
			errs = append(errs, fmt.Sprintf("mutate[%d].with: could not convert to typed config %T", k, vis))
		}
		c.Generators[k].With = vis
	}

	for k, mu := range c.Mutations {
		if mu.Uses == "" {
			errs = append(errs, fmt.Sprintf("mutate[%d]: %q is a required field", k, "uses"))
			continue
		}
		if !mutate.Has(mu.Uses) {
			errs = append(errs, fmt.Sprintf("mutate[%d]: %q does not exist", k, mu.Uses))
			continue
		}
		vis, err := mutate.Get(mu.Uses)
		if err != nil {
			panic("BUG: why is this happening. I just made sure it existed")
		}
		in, err := yaml.Marshal(mu.With)
		if err != nil {
			panic("BUG: this was just decoded")
		}
		if err := yaml.Unmarshal(in, vis); err != nil {
			errs = append(errs, fmt.Sprintf("mutate[%d].with: could not convert to typed config %T", k, vis))
		}
		c.Mutations[k].With = vis
	}
	return nil
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
