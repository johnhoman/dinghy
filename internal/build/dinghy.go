package build

import (
	goerr "errors"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/johnhoman/dinghy/internal/codec"
	"github.com/johnhoman/dinghy/internal/generate"
	"github.com/johnhoman/dinghy/internal/mutate"
	"github.com/johnhoman/dinghy/internal/path"
	"github.com/johnhoman/dinghy/internal/resource"
	"github.com/johnhoman/dinghy/internal/types"
)

const (
	DinghyFile        = "dinghyfile.yaml"
	ErrReadDinghyFile = "failed to read required file " + DinghyFile
)

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

func (d *dinghy) BuildFromConfig(c *types.Config, opts ...Option) (resource.Tree, error) {

	// validate config
	errs := &errList{}
	for _, m := range c.Mutations {
		use := m.Uses
		if !mutate.Has(use) {
			errs.Append(fmt.Errorf("mutator %q does not exist", use))
			continue
		}
		_, newConfig, _ := mutate.Get(use)
		// also check typed configs
		with := m.With
		typed := newConfig()
		if err := codec.YAMLCopyTo(typed, with); err != nil {
			errs.Append(err)
			continue
		}

		for _, kind := range m.Selector.Kinds {
			if strings.Count(kind, "/") > 2 {
				errs.Append(fmt.Errorf("kind string must be in the form group/version/kind"))
			}
		}
	}
	for _, m := range c.Generators {
		use := m.Uses
		if !generate.Has(use) {
			errs.Append(fmt.Errorf("generator %q does not exist", use))
			continue
		}
		_, newConfig, _ := generate.Get(use)
		// also check typed configs
		with := m.With
		typed := newConfig()
		if err := codec.YAMLCopyTo(typed, with); err != nil {
			errs.Append(err)
			continue
		}
	}
	if !errs.Empty() {
		return nil, errs
	}

	o := newOptions(opts...)
	// build resources
	errs = &errList{}
	for _, r := range c.Resources {
		rt := resource.NewTree()
		if err := buildResource(r, o.path, rt, New); err != nil {
			errs.Append(err)
		}
		if err := resource.CopyTree(o.tree, rt); err != nil {
			errs.Append(err)
		}
	}
	for _, r := range c.Overlays {
		if err := buildResource(r, o.path, o.tree, New); err != nil {
			errs.Append(err)
		}
	}

	errs = &errList{}
	for _, m := range c.Mutations {
		f, newConfig, err := mutate.Get(m.Uses)
		if err != nil {
			errs.Append(err)
			continue
		}
		typed := newConfig()
		if err := codec.YAMLCopyTo(typed, m.With); err != nil {
			errs.Append(err)
			continue
		}
		kinds := make([]schema.GroupVersionKind, 0)
		for _, kind := range m.Selector.Kinds {
			kinds = append(kinds, parseKind(kind))
		}
		vis, err := f(typed)
		if err != nil {
			errs.Append(err)
			continue
		}
		err = o.tree.Visit(vis,
			resource.MatchLabels(m.Selector.MatchLabels),
			resource.MatchNames(m.Selector.Names...),
			resource.MatchNamespaces(m.Selector.Namespaces...),
			resource.MatchKinds(kinds...),
		)
		if err != nil {
			errs.Append(err)
		}
	}
	for _, spec := range c.Generators {
		gen, newConfig, err := generate.Get(spec.Uses)
		if err != nil {
			errs.Append(err)
		}
		typed := newConfig()
		if err := codec.YAMLCopyTo(typed, spec.With); err != nil {
			errs.Append(err)
		}
		tr, err := gen.Emit(typed, generate.WithDirectoryRoot(o.path))
		if err != nil {
			errs.Append(err)
		} else {
			errs.Append(resource.CopyTree(o.tree, tr))
		}
	}
	if !errs.Empty() {
		return nil, errs
	}

	return o.tree, nil
}

func (d *dinghy) Build(path path.Path, opts ...Option) (resource.Tree, error) {
	c, err := ReadDinghyFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, ErrReadDinghyFile)
	}
	return d.BuildFromConfig(c, append(opts, WithPath(path))...)
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
func ReadDinghyFile(path path.Path) (*types.Config, error) {
	r, err := path.Join(DinghyFile).Open()
	if err != nil {
		return nil, err
	}
	c := &types.Config{}
	return c, codec.YAMLDecoder(r).Decode(c)
}

type errList []error

func (err *errList) Append(e error) {
	if e == nil {
		return
	}
	*err = append(*err, e)
}
func (err *errList) Error() string { return goerr.Join(*err...).Error() }

func (err *errList) Len() int {
	if err == nil {
		return 0
	}
	return len(*err)
}

func (err *errList) Empty() bool {
	return err.Len() == 0
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
