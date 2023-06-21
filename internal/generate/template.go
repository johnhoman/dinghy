package generate

import (
	"bytes"
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
	// the root will likely be a directory and not a single file, but
	// we should be able to handle either
	rv := resource.NewList()

	isDir, err := source.IsDir()
	if err != nil {
		return nil, err
	}

	if isDir {
		// Now I have to list everything under this directory -- need to example
		// path to include list dir
		panic("not implemented: directories are current not supported")
	}

	content, err := source.ReadText()
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New(source.String()).Option(templateOptionErrOnMissingKey).Parse(content)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if err = tmpl.Execute(buf, t.values); err != nil {
		return nil, err
	}
	return rv, resource.InsertFromReader(rv, buf)
}
