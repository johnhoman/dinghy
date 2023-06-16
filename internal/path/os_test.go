package path

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/frankban/quicktest"
)

func TestLocalReadFile(t *testing.T) {
	c := quicktest.New(t)

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		c.Assert(os.RemoveAll(tempDir), quicktest.IsNil)
	}()

	basePath := tempDir
	local := NewLocal()

	// Write a test file
	testData := []byte("Hello, World!")
	testFilePath := "testfile.txt"
	c.Assert(os.WriteFile(filepath.Join(basePath, testFilePath), testData, 0644), quicktest.IsNil)

	// Read file
	data, err := local.ReadFile(filepath.Join(basePath, testFilePath))
	c.Assert(err, quicktest.IsNil)
	c.Assert(data, quicktest.DeepEquals, testData)
}

func TestLocalIsDir(t *testing.T) {
	c := quicktest.New(t)

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "")
	c.Assert(err, quicktest.IsNil)
	defer func() {
		c.Assert(os.RemoveAll(tempDir), quicktest.IsNil)
	}()

	isDir, err := NewLocal().IsDir(tempDir)
	c.Assert(err, quicktest.IsNil)
	c.Assert(isDir, quicktest.IsTrue)

	// Check if file is a directory
	isDir, err = NewLocal().IsDir("nonexistant")
	c.Assert(err, quicktest.ErrorIs, os.ErrNotExist)
	c.Assert(isDir, quicktest.IsFalse)
}
