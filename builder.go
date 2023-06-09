package kustomize

import (
	"gopkg.in/yaml.v3"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var (
	ErrResourceAlreadyExists = errors.New("duplicate resource name")
)

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

type RenderOption func(r *Renderer)

func NewRenderer(config *Config, current Path, opts ...RenderOption) *Renderer {
	r := &Renderer{config: config, current: current}
	for _, f := range opts {
		f(r)
	}
	return r
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

// Build the provided Config into Kubernetes resource.
func (r *Renderer) Build(rm ResourceMap) (ResourceMap, error) {

	// TODO: maybe run this in a go routine using a channel so that
	//   I can process/transform the results as they're created

	// for now, just focus on local resources
	if rm == nil {
		rm = ResourceMap{}
	}
	for _, resource := range r.config.Resources {
		// render the resources
		// - resource rendering can happen recursively-ish

		var path Path
		switch ParseReferenceType(resource) {
		case ReferenceTypeRemoteGitHub:

			// Parse the resource path into a URL, then create
			// the GitHub path from the resource
			if !strings.HasPrefix(resource, "https://") && strings.HasPrefix(resource, "github.com") {
				// url.Parse does weird things when a URL doesn't have a scheme.
				resource = "https://" + resource
			}
			u, err := url.Parse(resource)
			if err != nil {
				return nil, err
			}

			if u.Path[0] == '/' {
				u.Path = u.Path[1:]
			}

			parts := strings.Split(u.Path, "/")
			owner := parts[0]
			repo := strings.TrimSuffix(parts[1], ".git")

			path, err = NewGitHubPath(owner, repo, u.Query().Get("ref"), os.Getenv("GITHUB_TOKEN"))
			if err != nil {
				return nil, err
			}
			path = path.Join(parts[2:]...)

		case ReferenceTypeLocal:
			path = r.current.Join(resource)
		}

		ok, err := path.IsDir()
		if err != nil {
			return nil, err
		}
		if ok {
			// read in config
			c := &Config{}
			err := func() error {
				f, err := path.Join("kustomization.yaml").Open()
				if err != nil {
					return err
				}
				if err := yaml.NewDecoder(f).Decode(c); err != nil {
					return err
				}
				return nil
			}()
			if err != nil {
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
			err := func() error {
				f, err := path.Open()
				if err != nil {
					return err
				}
				// this is messing up the ordering
				var m map[string]any
				if err := yaml.NewDecoder(f).Decode(&m); err != nil {
					return err
				}
				obj := &unstructured.Unstructured{Object: m}
				return rm.Append(obj)
			}()
			if err != nil {
				return nil, err
			}
		}
	}
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
