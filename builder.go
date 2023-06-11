package kustomize

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var (
	ErrResourceAlreadyExists = errors.New("duplicate resource name")
	ErrConvertTypedConfig    = errors.New("failed to convert config to typed config")
	ErrApplyMutator          = errors.New("an error occurred when applying mutation")
)

type RenderOption func(r *Renderer)

func NewRenderer(config *Config, current Path, opts ...RenderOption) *Renderer {
	r := &Renderer{config: config, current: current}
	for _, f := range opts {
		f(r)
	}
	return r
}

type Renderer struct {
	config *Config

	current Path
	// newGitHubPath is a factory for a new GitHub path. The current
	// renderer might represent a GitHub target or a local target, but
	// when a GitHub target is included, is always starts with a fresh
	// path, unlike the current path, which can just be modified to
	// accommodate relative paths
	newGitHubPath func() Path
}

// Build the provided Config into Kubernetes resource.
func (r *Renderer) Build(rm ResourceMap) (ResourceMap, error) {

	// TODO: maybe run this fieldPath a go routine using a channel so that
	//   I can process/transform the results as they're created

	// for now, just focus on local resources
	if rm == nil {
		rm = ResourceMap{}
	}
	for _, resource := range r.config.Resources {
		// render the resources
		// - resource rendering can happen recursively-ish

		path, err := Factory(r.current, resource)
		if err != nil {
			return nil, err
		}

		ok, err := path.IsDir()
		if err != nil {
			return nil, err
		}
		if ok {
			// read fieldPath config
			c := &Config{}
			f, err := path.Join("kustomization.yaml").Open()
			if err != nil {
				return nil, err
			}
			if err := yaml.NewDecoder(f).Decode(c); err != nil {
				return nil, err
			}
			sub := &Renderer{config: c, current: path}
			m, err := sub.Build(nil)
			if err != nil {
				return nil, err
			}
			if err := rm.Merge(m); err != nil {
				return nil, err
			}
		} else {
			f, err := path.Open()
			if err != nil {
				return nil, err
			}
			// this is messing up the ordering
			var m map[string]any
			if err := yaml.NewDecoder(f).Decode(&m); err != nil {
				return nil, err
			}
			obj := &unstructured.Unstructured{Object: m}
			if err := rm.Append(obj); err != nil {
				return nil, err
			}
		}
	}

	for k, ms := range r.config.Mutations {
		matcher := ms.Selector.Matcher()

		if ms.Name == "" {
			ms.Name = fmt.Sprintf("position %d", k)
		}

		var m *mutatePlugin
		if isRegisteredMutator(ms.Uses) {
			m = getRegisteredMutator(ms.Uses)
		} else {
			panic("Not implemented")
		}

		typed := m.newConfig()
		raw, err := yaml.Marshal(ms.With)
		if err != nil {
			panic("BUG: this was unmarshalled, so it's not clear why it can't be marshalled")
		}

		if err := yaml.Unmarshal(raw, typed); err != nil {
			return nil, errors.Wrapf(
				ErrConvertTypedConfig,
				"an error occurred when attempting to convert the provided typed config for mutation %q: %s",
				ms.Name,
				err,
			)
		}

		for _, obj := range rm {
			if matcher(obj) {
				if err := m.mutator.Apply(obj, rm, typed); err != nil {
					return nil, errors.Wrapf(ErrApplyMutator, "%q: %s", ms.Name, err)
				}
			}
		}
	}

	// vm := goja.New()
	// _, err := vm.RunString("")
	// if err != nil {
	// 	return nil, err
	// }
	// transform, ok := goja.AssertFunction(vm.Get("transform"))
	// if !ok {
	// }
	return rm, nil
}

// Render builds the current Config and writes the encoded YAML to the provided
// reader. If the provided reader is nil, os.Stdout will be used instead.
func (r *Renderer) Render(w io.Writer) error {
	rm, err := r.Build(nil)
	if err != nil {
		return err
	}

	if w == nil {
		w = os.Stdout
	}

	d := yaml.NewEncoder(w)
	d.SetIndent(2)
	for _, item := range rm {
		if err := d.Encode(item.Object); err != nil {
			return err
		}
	}
	return nil
}

// Print builds the current Config and writes the encoded YAML
// to standard out.
func (r *Renderer) Print() error {
	return r.Render(os.Stdout)
}

type ResourceMap map[ID]*unstructured.Unstructured

// Merge adds a resource map to an existing resource map. An
// error will be returned if there is a conflict
func (rm ResourceMap) Merge(in ResourceMap) error {
	for key, value := range in {
		if _, ok := rm[key]; ok {
			return errors.Wrap(ErrResourceAlreadyExists, key.String())
		}
		rm[key] = value
	}
	return nil
}

// Append adds a resource to the resource map. If an object with the same
// ID already exists, and error will be returned
func (rm ResourceMap) Append(obj *unstructured.Unstructured) error {
	key := ID{
		APIGroup:  obj.GroupVersionKind().Group,
		Kind:      obj.GroupVersionKind().Kind,
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}
	if _, ok := rm[key]; ok {
		return errors.Wrap(ErrResourceAlreadyExists, key.String())
	}
	rm[key] = obj
	return nil
}

// ID is a unique identifier for a resource.
type ID struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
	APIGroup  string `yaml:"apiGroup"`
	Kind      string `yaml:"kind"`
}

func (id ID) String() string {
	name := id.Name
	if id.Namespace != "" {
		name = id.Namespace + "/" + name
	}
	parts := []string{id.APIGroup, id.Kind, id.Namespace, id.Name}
	return strings.ToLower(strings.Join(parts, "/"))
}
