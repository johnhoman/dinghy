package generate

import (
	"bytes"
	"github.com/imdario/mergo"
	"github.com/invopop/jsonschema"
	"text/template"

	"gopkg.in/yaml.v3"

	"github.com/johnhoman/dinghy/internal/context"
	"github.com/johnhoman/dinghy/internal/errors"
	"github.com/johnhoman/dinghy/internal/path"
	"github.com/johnhoman/dinghy/internal/resource"
)

const (
	templateOptionErrOnMissingKey = "missingkey=error"
)

var (
	_ yaml.Unmarshaler = &Template{}
	_ Generator        = &Template{}
)

type Template struct {
	source path.Path
	values map[string]any
}

func (t *Template) Name() string {
	return "builtin.dinghy.dev/template"
}

func (t *Template) UnmarshalYAML(value *yaml.Node) error {
	var in struct {
		Source string         `yaml:"source" json:"source"`
		Values map[string]any `yaml:"values" json:"values"`
	}

	var m map[string]any
	if err := value.Decode(&m); err != nil {
		return &errors.ErrDecodeGenerator{
			Name:   t.Name(),
			Schema: jsonschema.Reflect(in),
			Err:    err,
		}
	}

	// the value.Decode method doesn't provide a way to only
	// parse known fields, so I have to decode it with the KnownFields flag
	// set

	data, _ := yaml.Marshal(m)
	d := yaml.NewDecoder(bytes.NewReader(data))
	d.KnownFields(true)

	if err := d.Decode(&in); err != nil {
		return &errors.ErrDecodeGenerator{
			Name:   t.Name(),
			Schema: jsonschema.Reflect(in),
			Err:    err,
		}
	}

	source, err := path.Parse(in.Source)
	if err != nil {
		return &errors.ErrParseSourcePath{
			Path: in.Source,
			Err:  err,
		}
	}

	t.source = source
	t.values = in.Values
	return nil
}

func (t *Template) Emit(ctx *context.Context) (resource.Tree, error) {
	source := t.source
	if source.Relative() {
		// source is a relative path, to join the build prefix
		root, err := path.Parse(ctx.Root())
		if err != nil {
			return nil, err
		}
		source = root.Join(source.String())
	}

	templates, vars, err := templateBuildDir(source)
	if err != nil {
		return nil, err
	}

	values := t.values
	if values == nil {
		values = make(map[string]any)
	}
	if err = mergo.Merge(&values, vars); err != nil {
		return nil, err
	}

	// the root will likely be a directory and not a single file, but
	// we should be able to handle either
	rv := resource.NewList()
	for _, content := range templates {
		tmpl, err := template.New(source.String()).Option(templateOptionErrOnMissingKey).Parse(content)
		if err != nil {
			return nil, err
		}

		buf := new(bytes.Buffer)
		if err = tmpl.Execute(buf, values); err != nil {
			return nil, err
		}
		if err := resource.InsertFromReader(rv, buf); err != nil {
			return nil, err
		}
	}
	return rv, nil
}

func templateReadConfig(source path.Path) (TemplateConfig, error) {
	c := TemplateConfig{}

	f, err := source.Reader("template.dinghyfile.yaml")
	if err != nil {
		return c, err
	}
	return c, yaml.NewDecoder(f).Decode(&c)
}

func templateBuildDir(source path.Path) (templates []string, vars map[string]any, err error) {
	var c TemplateConfig
	c, err = templateReadConfig(source)
	if err != nil {
		return
	}

	vars = c.Values

	var src path.Path
	for _, res := range c.Templates {
		src, err = path.Parse(res)
		if err != nil {
			return
		}
		if src.Relative() {
			src = source.Join(res)
		}

		var ok bool
		ok, err = src.IsDir()
		if err != nil {
			return
		}
		if ok {
			var (
				items []string
				defs  map[string]any
			)

			items, defs, err = templateBuildDir(src)
			if err != nil {
				return
			}
			templates = append(templates, items...)
			if err = mergo.Merge(&vars, defs); err != nil {
				return
			}
			continue
		}

		var s string
		s, err = src.ReadText()
		if err != nil {
			return
		}
		templates = append(templates, s)
	}
	return
}

type TemplateConfig struct {
	APIVersion string         `json:"apiVersion" yaml:"apiVersion"`
	Kind       string         `json:"kind" yaml:"kind"`
	Templates  []string       `json:"templates" yaml:"templates"`
	Values     map[string]any `json:"values" yaml:"values"`
}
