package kustomize

import (
	qt "github.com/frankban/quicktest"
	"testing"
)

func TestRenderer_Render(t *testing.T) {
	c := NewConfig()
	c.AddResource("./examples")
	r := &Renderer{config: c, current: NewOsPath(".")}
	rm, err := r.Build(nil)
	qt.Assert(t, err, qt.IsNil)
	qt.Assert(t, rm, qt.HasLen, 3)
}
