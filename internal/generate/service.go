package generate

import (
	"bytes"
	_ "embed"
	"github.com/johnhoman/dinghy/internal/context"
	"text/template"

	"github.com/johnhoman/dinghy/internal/resource"
)

//go:embed service.tmpl
var serviceContent string

type Service struct {
	Name  string `yaml:"name"  dinghy:"required"`
	Image string `yaml:"image" dinghy:"required"`
}

// Emit generates the minimum necessary resources to deploy a
// web app on a cluster. Combine this generator with mutators
// to customize the web app
func (s *Service) Emit(ctx *context.Context) (resource.Tree, error) {
	tree := resource.NewTree()
	tmpl, err := template.New("").Parse(serviceContent)
	if err != nil {
		panic("BUG: internal serviceTemplates failed to parse")
	}
	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, s); err != nil {
		return nil, err
	}
	return tree, resource.InsertFromReader(tree, buf)
}
