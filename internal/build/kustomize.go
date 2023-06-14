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
	"github.com/johnhoman/dinghy/internal/visitor"
)

type KustomizeGeneratorConfig struct {
	Source string `yaml:"source"`
}

func KustomizeGenerator() generate.Generator {
	return generate.Func(func(config any, opts ...generate.Option) (resource.Tree, error) {
		c, ok := config.(*KustomizeGeneratorConfig)
		if !ok {
			return nil, generate.ErrTypedConfig
		}
		o := generate.Options{}
		for _, f := range opts {
			f(&o)
		}
		b := NewKustomize()
		// I need the build directory here
		dir, err := path.Parse(c.Source, path.WithRelativeRoot(o.Root))
		if err != nil {
			return nil, err
		}
		return b.Build(dir, WithPath(o.Root))
	})
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
	// namespace
	if len(c.Namespace) > 0 {
		vis, err := mutate.Namespace(&visitor.NamespaceConfig{
			Namespace: c.Namespace,
		})
		if err != nil {
			return nil, err
		}
		errs.Append(o.tree.Visit(vis))
	}
	// commonLabels
	// commonAnnotations
	if len(c.CommonAnnotations) > 0 {
		vis, err := mutate.AddAnnotations(&c.CommonAnnotations)
		if err != nil {
			errs.Append(err)
		} else {
			errs.Append(o.tree.Visit(vis))
		}
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
		ok, err := path.Join(name).Exists()
		if err != nil {
			return nil, err
		}
		if ok {
			f, err := path.Join(name).Open()
			if err != nil {
				return nil, err
			}
			c := &types.Kustomization{}
			return c, yaml.NewDecoder(f).Decode(c)
		}
	}
	return nil, errors.Wrapf(os.ErrNotExist, path.Join(konfig.DefaultKustomizationFileName()).String())
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
		content, err = pp.Open()
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
		vis, err := mutate.JSONPatch(jp)
		if err != nil {
			return err
		}
		return tree.Visit(vis)
	}
	var m map[string]any
	if err := yaml.Unmarshal(raw, &m); err != nil {
		return err
	}
	vis, err := mutate.StrategicMergePatch(&m)
	if err != nil {
		return err
	}
	return tree.Visit(vis, o...)
}

func init() {
	// This package imports generate
	generate.MustRegister("builtin.dinghy.dev/kustomize", KustomizeGenerator(), func() any {
		return &KustomizeGeneratorConfig{}
	})
}
