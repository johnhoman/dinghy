package generate

import (
	"bytes"
	_ "embed"
	"io"
	"text/template"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/johnhoman/dinghy/internal/resource"
)

//go:embed service.tmpl
var svcTemplate string

type ServiceConfig struct {
	Name  string `yaml:"name"  dinghy:"required"`
	Image string `yaml:"image" dinghy:"required"`
}

// Service generates the minimum necessary resources to
// deploy a web app on a cluster.
func Service() Generator {
	return Func(func(config any, _ ...Option) (resource.Tree, error) {
		c, ok := config.(*ServiceConfig)
		if !ok {
			return nil, ErrTypedConfig
		}

		tree := resource.NewTree()
		tmpl, err := template.New("").Parse(svcTemplate)
		if err != nil {
			panic("BUG: internal serviceTemplates failed to parse")
		}
		buf := new(bytes.Buffer)
		if err := tmpl.Execute(buf, c); err != nil {
			return nil, err
		}
		d := yaml.NewDecoder(buf)
		for {
			var m map[string]any
			if err := d.Decode(&m); err != nil {
				if err == io.EOF {
					break
				}
				return nil, err
			}
			if m == nil {
				continue
			}
			obj := &unstructured.Unstructured{Object: m}
			if err := tree.Insert(obj); err != nil {
				return nil, err
			}
		}
		return tree, nil
	})
}
