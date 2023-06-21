package fieldpath

import (
	qt "github.com/frankban/quicktest"
	"testing"
)

func TestFieldPath_Parse_ValidFieldPath(t *testing.T) {
	c := qt.New(t)

	validPath := "spec.replicas"
	path, err := Parse(validPath)
	c.Assert(err, qt.IsNil)
	c.Assert(path.indexes, qt.HasLen, 2)
	c.Assert(path.indexes[0].index, qt.Equals, "spec")
	c.Assert(path.indexes[0].it, qt.Equals, IndexTypeMapKey)
	c.Assert(path.indexes[1].index, qt.Equals, "replicas")
	c.Assert(path.indexes[1].it, qt.Equals, IndexTypeMapKey)

	validPath = `metadata.labels."app.kubernetes.io/name"`
	path, err = Parse(validPath)
	c.Assert(err, qt.IsNil)
	c.Assert(path.indexes, qt.HasLen, 3)
	c.Assert(path.indexes[0].index, qt.Equals, "metadata")
	c.Assert(path.indexes[0].it, qt.Equals, IndexTypeMapKey)
	c.Assert(path.indexes[1].index, qt.Equals, "labels")
	c.Assert(path.indexes[1].it, qt.Equals, IndexTypeMapKey)
	c.Assert(path.indexes[2].index, qt.Equals, "app.kubernetes.io/name")
	c.Assert(path.indexes[2].it, qt.Equals, IndexTypeMapKey)
}
