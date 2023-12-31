package generate

import (
	"github.com/johnhoman/dinghy/internal/context"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/johnhoman/dinghy/internal/resource"
)

func TestService(t *testing.T) {
	c := &Service{
		Name:  "webapp",
		Image: "nginx:latest",
	}
	tree, err := c.Emit(&context.Context{})
	qt.Assert(t, err, qt.IsNil)
	keys := []resource.Key{
		{GroupVersion: "apps/v1", Kind: "Deployment", Name: "webapp"},
		{GroupVersion: "v1", Kind: "Service", Name: "webapp"},
		{GroupVersion: "v1", Kind: "ServiceAccount", Name: "webapp"},
	}
	for _, key := range keys {
		obj, err := resource.GetResource(tree, key)
		qt.Assert(t, err, qt.IsNil)
		qt.Assert(t, obj, qt.IsNotNil)
	}
}
