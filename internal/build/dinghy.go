package build

import (
	"bytes"
	"fmt"
	"github.com/johnhoman/dinghy/internal/codec"
	"github.com/johnhoman/dinghy/internal/context"
	"github.com/johnhoman/dinghy/internal/generate"
	"github.com/johnhoman/dinghy/internal/mutate"
	"github.com/johnhoman/dinghy/internal/path"
	"github.com/johnhoman/dinghy/internal/resource"
	"github.com/johnhoman/dinghy/internal/types"
	"github.com/johnhoman/dinghy/internal/visitor"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"
)

const (
	DinghyFile        = "dinghyfile.yaml"
	ErrReadDinghyFile = "failed to read required file " + DinghyFile
)

type ErrMutateFailedApply struct {
	// the name of the mutator
	Name string
	// description is the reason the error occured
	Description string
	// Err is the actual error message
	Err error
}

func ErrMutateFailedYAMLDecode(name string, description string, parent error) error {
	return &ErrMutateFailedApply{
		Name:        name,
		Description: description,
		Err:         parent,
	}
}

func (err *ErrMutateFailedApply) Error() string {
	return fmt.Sprintf(err.Description)
}

type Option func(o *options)

// WithTree injects an existing tree into the build context, so that
// sub packages can mutate/validate the existing resource.
func WithTree(tree resource.Tree) Option {
	return func(o *options) {
		o.tree = tree
	}
}

// WithPath sets the current build path. This is required
// if the Config has references to relative paths. If all paths
// in the build are remote or absolute, it's not required.
func WithPath(path path.Path) Option {
	return func(o *options) {
		o.path = path
	}
}

type options struct {
	// tree is an optional resource Tree to augment. If a tree
	// is provided, mutations and validations will consider existing
	// resources.
	tree resource.Tree
	// path is the current build path, which is required for relative references
	// to files in the build path
	path path.Path
}

type dinghy struct{}

func (d *dinghy) BuildFromConfig(ctx *context.Context, c *types.Config, opts ...Option) (resource.Tree, error) {
	o := newOptions(opts...)

	// build resources
	for _, r := range c.Resources {
		// sub-resources, such as other dinghy packages can contain
		// transformers that should only act on their set of resources,
		// so we need to provide a new tree so that none of the current
		// resources are mutated
		rt := resource.NewTree()
		if err := d.buildResource(ctx, r, o.path, rt); err != nil {
			return nil, err
		}
		if err := resource.CopyTree(o.tree, rt); err != nil {
			return nil, err
		}
	}

	for _, r := range c.Overlays {
		if err := d.buildResource(ctx, r, o.path, o.tree); err != nil {
			return nil, err
		}
	}

	for _, m := range c.Mutations {
		typed, err := mutate.Get(m.Uses)
		if err != nil {
			return nil, err
		}
		data, err := yaml.Marshal(m.With)
		if err != nil {
			return nil, err
		}
		d := yaml.NewDecoder(bytes.NewBuffer(data))
		d.KnownFields(true)
		if err = d.Decode(typed); err != nil {
			return nil, err
		}
		vis := typed.(visitor.Visitor)

		kinds := make([]schema.GroupVersionKind, 0)
		for _, kind := range m.Selector.Kinds {
			kinds = append(kinds, parseKind(kind))
		}
		if se, ok := vis.(mutate.SideEffectVisitor); ok {
			vis = mutate.SideEffect(se, o.tree)
		}

		if err := o.tree.Visit(vis,
			resource.MatchLabels(m.Selector.MatchLabels),
			resource.MatchNames(m.Selector.Names...),
			resource.MatchNamespaces(m.Selector.Namespaces...),
			resource.MatchKinds(kinds...)); err != nil {
			return nil, err
		}
	}
	for _, spec := range c.Generators {
		sub, err := d.doGenerate(ctx, spec)
		if err != nil {
			return nil, err
		}
		if err = resource.CopyTree(o.tree, sub); err != nil {
			return nil, err
		}
	}

	return o.tree, nil
}

func (d *dinghy) buildResource(ctx *context.Context, r string, root path.Path, tree resource.Tree) error {
	target := root.Join(r)
	if !path.IsRelative(r) {
		parsed, err := path.Parse(r)
		if err != nil {
			return err
		}
		target = parsed
	}

	isDir, err := target.IsDir()
	if err != nil {
		return err
	}
	if isDir {
		var sub resource.Tree
		sub, err = d.Build(ctx, target)
		if err != nil {
			return err
		}
		return resource.CopyTree(tree, sub)
	}

	f, err := target.Reader()
	if err != nil {
		return err
	}
	dec := yaml.NewDecoder(f)
	for {
		var m map[string]any
		if err := dec.Decode(&m); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if m == nil {
			continue
		}
		obj := &unstructured.Unstructured{Object: m}
		if err = tree.Insert(obj); err != nil {
			return err
		}
	}
	return nil
}

func (d *dinghy) doGenerate(ctx *context.Context, spec types.GeneratorSpec) (resource.Tree, error) {
	typed, err := generate.Get(spec.Uses)
	if err != nil {
		return nil, err
	}
	data, err := yaml.Marshal(spec.With)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, typed); err != nil {
		return nil, err
	}

	return typed.(generate.Generator).Emit(ctx)
}

func (d *dinghy) Build(ctx *context.Context, path path.Path, opts ...Option) (resource.Tree, error) {
	c, err := ReadDinghyFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, ErrReadDinghyFile)
	}
	return d.BuildFromConfig(ctx, c, append(opts, WithPath(path))...)
}

func newOptions(opts ...Option) *options {
	o := &options{
		tree: resource.NewTree(),
	}
	for _, f := range opts {
		f(o)
	}
	return o
}

// ReadDinghyFile loads the dingy file in the current path if it exists. If,
// it doesn't exist, os.ErrNotExist is returned
func ReadDinghyFile(p path.Path) (*types.Config, error) {
	data, err := p.ReadFile(DinghyFile)
	if err != nil {
		return nil, err
	}
	c := &types.Config{}
	return c, codec.YAMLDecoder(bytes.NewReader(data)).Decode(c)
}

func parseKind(kind string) schema.GroupVersionKind {
	parts := strings.Split(kind, "/")
	switch len(parts) {
	case 0:
		return schema.GroupVersionKind{Group: "*", Version: "*", Kind: parts[0]}
	case 1:
		return schema.GroupVersionKind{Group: parts[0], Version: "*", Kind: parts[1]}
	case 2:
		return schema.GroupVersionKind{Group: parts[0], Version: parts[1], Kind: parts[2]}
	default:
		panic("invalid gvk string")
	}
}
