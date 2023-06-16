package dinghy

import (
	"github.com/johnhoman/dinghy/internal/fieldpath"
	"github.com/johnhoman/dinghy/internal/generate"
	"github.com/johnhoman/dinghy/internal/mutate"
	"github.com/johnhoman/dinghy/internal/path"
	"github.com/johnhoman/dinghy/internal/resource"
	"github.com/johnhoman/dinghy/internal/scheme"
	"github.com/johnhoman/dinghy/internal/visitor"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type (
	Path       = path.Path
	GitHubPath = path.GitHub
	LocalPath  = path.Local
	MemoryPath = path.Memory

	Visitor     = visitor.Visitor
	VisitorFunc = visitor.Func
	Tree        = resource.Tree
	Key         = resource.Key

	FieldPath = fieldpath.FieldPath
)

var (
	ParsePath         = path.Parse
	RegisterMutator   = mutate.MustRegister
	RegisterGenerator = generate.Register
	Scheme            = scheme.Scheme
)

func AddKnownTypes(kind schema.GroupVersion, o runtime.Object) {
	scheme.Scheme.AddKnownTypes(kind, o)
}
