package path

import (
	qt "github.com/frankban/quicktest"
	"testing"
)

type mockImpl struct {
	joinFunc     func(string, ...string) string
	readFileFunc func(string) ([]byte, error)
	isDirFunc    func(string) (bool, error)
}

func (m *mockImpl) ReadFile(path string) ([]byte, error) {
	return m.readFileFunc(path)
}

func (m *mockImpl) IsDir(path string) (bool, error) {
	return m.isDirFunc(path)
}

func (m *mockImpl) join(root string, segments ...string) string {
	return m.joinFunc(root, segments...)
}

func (m *mockImpl) toString(root string, segments ...string) string {
	return m.joinFunc(root, segments...)
}

func TestPath_ReadFile(t *testing.T) {
	c := qt.New(t)

	mockPath := &mockImpl{}
	root := "path/to"

	path := Path{
		path: mockPath,
		root: root,
	}

	// Test calling ReadFile with joined path
	expectedJoinedPath := "path/to/file.txt"
	mockPath.joinFunc = func(root string, segments ...string) string {
		return expectedJoinedPath
	}

	mockPath.readFileFunc = func(path string) ([]byte, error) {
		return nil, nil
	}

	_, err := path.ReadFile("file.txt")
	c.Assert(err, qt.IsNil)

	// Test calling ReadFile with root path
	expectedRootPath := "path/to"
	mockPath.joinFunc = func(root string, segments ...string) string {
		return expectedRootPath
	}

	_, err = path.ReadFile("")
	c.Assert(err, qt.IsNil)
}

func TestPath_IsDir(t *testing.T) {
	c := qt.New(t)

	mockPath := &mockImpl{}
	root := "path/to"

	path := Path{
		path: mockPath,
		root: root,
	}

	// Test calling IsDir with joined path
	expectedJoinedPath := "path/to/dir"
	mockPath.joinFunc = func(root string, segments ...string) string {
		return expectedJoinedPath
	}

	mockPath.isDirFunc = func(path string) (bool, error) {
		return false, nil
	}

	_, err := path.IsDir("dir")
	c.Assert(err, qt.IsNil)

	// Test calling IsDir with root path
	expectedRootPath := "path/to"
	mockPath.joinFunc = func(root string, segments ...string) string {
		return expectedRootPath
	}

	_, err = path.IsDir("")
	c.Assert(err, qt.IsNil)
}

func TestPath_Join(t *testing.T) {
	c := qt.New(t)

	mockPath := &mockImpl{}
	root := "path/to"

	path := Path{
		path: mockPath,
		root: root,
	}

	segments := []string{"dir", "file.txt"}
	expectedJoinedPath := "path/to/dir/file.txt"
	mockPath.joinFunc = func(root string, segments ...string) string {
		return expectedJoinedPath
	}

	joinedPath := path.Join(segments...)
	c.Assert(joinedPath.root, qt.Equals, expectedJoinedPath)
}
