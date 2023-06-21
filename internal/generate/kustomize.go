package generate

import (
	"bytes"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/types"

	"github.com/johnhoman/dinghy/internal/codec"
	"github.com/johnhoman/dinghy/internal/context"
	"github.com/johnhoman/dinghy/internal/mutate"
	"github.com/johnhoman/dinghy/internal/path"
	"github.com/johnhoman/dinghy/internal/resource"
)

type Kustomize struct {
	Source string `yaml:"source"`
}

func (c *Kustomize) Name() string {
	return "builtin.dinghy.dev/kustomize"
}

// Emit a kustomization package tree
func (c *Kustomize) Emit(ctx *context.Context) (resource.Tree, error) {

	// if the source is a relative path, then we need to get the root
	// of the build from the context, e.g. where the dinghy file is relative
	// to the current working directory, because the generator referencing the
	// kustomization package will be relative to that.
	src := c.Source
	if path.IsRelative(c.Source) {
		// I need to join it to the relative root, which could potentially
		// be an external path
		pth, err := path.Parse(ctx.Root())
		if err != nil {
			return nil, err
		}
		src = pth.String(c.Source)
	}

	b := &kustomize{}
	source, err := path.Parse(src)
	if err != nil {
		return nil, err
	}

	return b.Build(source)
}

type kustomize struct{}

func (k *kustomize) Build(path path.Path) (resource.Tree, error) {
	c, err := ReadKustomizationFile(path)
	if err != nil {
		return nil, err
	}
	return k.buildFromConfig(c, path)
}

func (k *kustomize) buildResource(r string, dir path.Path, tree resource.Tree) error {
	target := dir.Join(r)
	if !path.IsRelative(r) {
		var err error
		target, err = path.Parse(r)
		if err != nil {
			return err
		}
	}
	ok, err := target.IsDir()
	if err != nil {
		return err
	}
	if ok {
		sub, err := k.Build(target)
		if err != nil {
			return err
		}
		if err := resource.CopyTree(tree, sub); err != nil {
			return err
		}
		return nil
	}
	// read the file
	f, err := target.Reader()
	if err != nil {
		return err
	}
	return resource.InsertFromReader(tree, f)
}

func (k *kustomize) buildFromConfig(c *types.Kustomization, dir path.Path) (resource.Tree, error) {

	tree := resource.NewTree()
	for _, r := range c.Resources {
		// This resource could be a relative local path, which means it needs to get
		// joined from the provided dir, otherwise it's an absolute path and should be parsed
		t := resource.NewTree()
		if err := k.buildResource(r, dir, t); err != nil {
			return nil, err
		}
		if err := resource.CopyTree(tree, t); err != nil {
			return nil, err
		}
	}
	for _, r := range c.Components {
		if err := k.buildResource(r, dir, tree); err != nil {
			return nil, err
		}
	}
	// namePrefix
	// nameSuffix
	if len(c.NamePrefix) > 0 || len(c.NameSuffix) > 0 {
		if err := tree.Visit(&mutate.Name{Prefix: c.NamePrefix, Suffix: c.NameSuffix}); err != nil {
			return nil, err
		}
	}
	// namespace
	if len(c.Namespace) > 0 {
		if err := tree.Visit(&mutate.Namespace{Namespace: c.Namespace}); err != nil {
			return nil, err
		}
	}
	// commonLabels
	// commonAnnotations
	if len(c.CommonAnnotations) > 0 {
		mu := mutate.Annotations(c.CommonAnnotations)
		if err := tree.Visit(&mu); err != nil {
			return nil, err
		}
	}
	// patches is either a strategicMergePatch or a json patch. I guess it's
	// up to me to guess which one, since they deprecated the method of
	// explicitly choosing one
	for _, patch := range c.Patches {
		if err := kustomizePatch(dir, patch, tree); err != nil {
			return nil, err
		}
	}
	return tree, nil
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
	var jp mutate.JSONPatch
	// try jsonpatch first, if it doesn't decode, assume it's a strategicMergePatch
	if err = codec.YAMLDecoder(bytes.NewReader(raw)).Decode(&jp); err == nil {
		return tree.Visit(&jp)
	}
	var mergePatch mutate.StrategicMergePatch
	if err := yaml.Unmarshal(raw, &mergePatch); err != nil {
		return err
	}
	return tree.Visit(&mergePatch, o...)
}
