package build

import (
	"bytes"
	"github.com/johnhoman/dinghy/internal/generate"
	"io"
	"os"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/types"

	"github.com/johnhoman/dinghy/internal/codec"
	"github.com/johnhoman/dinghy/internal/mutate"
	"github.com/johnhoman/dinghy/internal/path"
	"github.com/johnhoman/dinghy/internal/resource"
)

type KustomizeGenerator struct {
	Source string `yaml:"source"`
}

func (c *KustomizeGenerator) Emit() (resource.Tree, error) {
	b := NewKustomize()
	source, err := path.Parse(c.Source)
	if err != nil {
		return nil, err
	}

	return b.Build(source)
}

type kustomize struct{}

func (k *kustomize) Build(path path.Path, opts ...Option) (resource.Tree, error) {
	c, err := ReadKustomizationFile(path)
	if err != nil {
		return nil, err
	}
	return k.buildFromConfig(c, append(opts, WithPath(path))...)
}

func (k *kustomize) buildFromConfig(c *types.Kustomization, opts ...Option) (resource.Tree, error) {

	o := newOptions(opts...)
	errs := &errList{}
	for _, r := range c.Resources {
		tree := resource.NewTree()
		if err := buildResource(r, o.path, tree, NewKustomize); err != nil {
			errs.Append(err)
			continue
		}
		errs.Append(resource.CopyTree(o.tree, tree))
	}
	for _, r := range c.Components {
		errs.Append(buildResource(r, o.path, o.tree, NewKustomize))
	}
	// namePrefix
	// nameSuffix
	if len(c.NamePrefix) > 0 || len(c.NameSuffix) > 0 {
		errs.Append(o.tree.Visit(&mutate.Name{Prefix: c.NamePrefix, Suffix: c.NameSuffix}))
	}
	// namespace
	if len(c.Namespace) > 0 {
		errs.Append(o.tree.Visit(&mutate.Namespace{Name: c.Namespace}))
	}
	// commonLabels
	// commonAnnotations
	if len(c.CommonAnnotations) > 0 {
		mu := mutate.Annotations(c.CommonAnnotations)
		errs.Append(o.tree.Visit(&mu))
	}
	// patches is either a strategicMergePatch or a json patch. I guess it's
	// up to me to guess which one, since they deprecated the method of
	// explicitly choosing one
	for _, patch := range c.Patches {
		errs.Append(kustomizePatch(o.path, patch, o.tree))
	}
	if !errs.Empty() {
		return nil, errs
	}
	return o.tree, nil
}

func ReadKustomizationFile(path path.Path) (*types.Kustomization, error) {
	for _, name := range konfig.RecognizedKustomizationFileNames() {
		ok, err := path.Exists(name)
		if err != nil {
			return nil, err
		}
		if ok {
			f, err := path.Reader(name)
			if err != nil {
				return nil, err
			}
			c := &types.Kustomization{}
			return c, yaml.NewDecoder(f).Decode(c)
		}
	}
	return nil, errors.Wrapf(os.ErrNotExist, path.String(konfig.DefaultKustomizationFileName()))
}

func kustomizePatch(path path.Path, patch types.Patch, tree resource.Tree) error {
	var o []resource.MatchOption
	if target := patch.Target; target != nil {
		o = append(o,
			resource.MatchNames(target.Name),
			resource.MatchNamespaces(target.Namespace),
			resource.MatchKinds(schema.GroupVersionKind{
				Group:   target.Group,
				Version: target.Version,
				Kind:    target.Kind,
			}),
		)
	}
	var content io.Reader = strings.NewReader(patch.Patch)
	if len(patch.Path) > 0 {
		pp := path.Join(patch.Path)
		ok, err := pp.Exists()
		if err != nil {
			return err
		}
		if !ok {
			return errors.Wrapf(os.ErrNotExist, pp.String())
		}
		content, err = pp.Reader()
		if err != nil {
			return err
		}
	}
	raw, err := io.ReadAll(content)
	if err != nil {
		return err
	}

	// TODO: read the json patch spec
	var jp jsonpatch.Patch
	// try jsonpatch first, if it doesn't decode, assume it's a strategicMergePatch
	if err = codec.YAMLDecoder(bytes.NewReader(raw)).Decode(&jp); err == nil {
		patch := mutate.JSONPatch(jp)
		return tree.Visit(&patch)
	}
	var m map[string]any
	if err := yaml.Unmarshal(raw, &m); err != nil {
		return err
	}
	p := mutate.StrategicMergePatch(m)
	return tree.Visit(&p, o...)
}

func init() {
	// This package imports generate
	generate.Register("builtin.dinghy.dev/kustomize", &KustomizeGenerator{})
}
