package path

import (
	"bytes"
	"github.com/frankban/quicktest"
	"os"
	"strings"
	"testing"
)

func TestMemory_ReadFile(t *testing.T) {
	m := Memory{
		"path": Memory{
			"to": Memory{
				"file.txt": []byte("This is the content of the file"),
			},
		},
	}

	// Test reading an existing file
	content, err := m.ReadFile("path/to/file.txt")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	expectedContent := []byte("This is the content of the file")
	if !bytes.Equal(content, expectedContent) {
		t.Errorf("Expected content: %v, got: %v", expectedContent, content)
	}

	// Test reading a non-existing file
	_, err = m.ReadFile("non/existing/file.txt")
	if err == nil {
		t.Error("Expected error, got nil")
	} else if !os.IsNotExist(err) {
		t.Errorf("Expected os.ErrNotExist, got: %v", err)
	}

	// Test reading a directory as a file
	_, err = m.ReadFile("path/to")
	if err == nil {
		t.Error("Expected error, got nil")
	} else if !strings.Contains(err.Error(), "is a directory") {
		t.Errorf("Expected directory error, got: %v", err)
	}
}

func TestMemory_IsDir(t *testing.T) {
	c := quicktest.New(t)
	m := Memory{
		"path": Memory{
			"to": Memory{
				"file.txt": []byte("This is the content of the file"),
			},
		},
	}

	// Test existing directory
	existingDir, err := m.IsDir("path/to")
	c.Assert(err, quicktest.IsNil)
	c.Assert(existingDir, quicktest.IsTrue)

	// Test non-existing directory
	nonExistingDir, err := m.IsDir("non/existing/directory")
	c.Assert(err, quicktest.ErrorIs, os.ErrNotExist)
	c.Assert(nonExistingDir, quicktest.IsFalse)
}

func TestMemory_Join(t *testing.T) {
	m := Memory{}
	root := "path"
	segments := []string{"to", "file.txt"}

	expectedPath := "path/to/file.txt"
	joinedPath := m.join(root, segments...)
	if joinedPath != expectedPath {
		t.Errorf("Expected path: %s, got: %s", expectedPath, joinedPath)
	}
}

func TestMemory_WriteFile(t *testing.T) {
	m := Memory{}

	// Test writing a file
	err := m.WriteFile("path/to/file.txt", []byte("This is the content"))
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Test reading the written file
	content, err := m.ReadFile("path/to/file.txt")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	expectedContent := []byte("This is the content")
	if !bytes.Equal(content, expectedContent) {
		t.Errorf("Expected content: %v, got: %v", expectedContent, content)
	}

	// Test writing a file to a nested path
	err = m.WriteFile("path/to/nested/file.txt", []byte("Nested file content"))
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Test reading the nested file
	nestedContent, err := m.ReadFile("path/to/nested/file.txt")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	expectedNestedContent := []byte("Nested file content")
	if !bytes.Equal(nestedContent, expectedNestedContent) {
		t.Errorf("Expected nested content: %v, got: %v", expectedNestedContent, nestedContent)
	}
}
