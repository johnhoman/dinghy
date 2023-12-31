package main

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestCmdBuild_Run(t *testing.T) {
	afs := afero.NewOsFs()
	examples := "../../examples/"

	matches := make([]string, 0)
	err := afero.Walk(afs, examples, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}
		pkg := filepath.Join(path, "dinghyfile.yaml")
		if _, err := afs.Stat(pkg); err == nil {
			matches = append(matches, path)
		}
		return nil
	})
	qt.Assert(t, err, qt.IsNil)

	for _, dir := range matches {
		t.Run(strings.TrimPrefix(dir, examples), func(t *testing.T) {
			cmd := &cmdBuild{Dir: dir}
			buf := new(bytes.Buffer)
			qt.Assert(t, cmd.Run(buf), qt.IsNil)
			got := decodeStream(t, buf)
			want := decodeExpected(t, filepath.Join(dir))
			qt.Assert(t, got, qt.DeepEquals, want)
		})
	}
}

func decodeStream(t *testing.T, r io.Reader) []any {
	rv := make([]any, 0)
	d := yaml.NewDecoder(r)
	for {
		var obj map[string]any
		err := d.Decode(&obj)
		if errors.Is(err, io.EOF) {
			break
		}
		qt.Assert(t, err, qt.IsNil)
		if obj == nil {
			continue
		}
		rv = append(rv, obj)
	}
	sort.Slice(rv, func(i, j int) bool {
		u1 := &unstructured.Unstructured{Object: rv[i].(map[string]any)}
		u2 := &unstructured.Unstructured{Object: rv[j].(map[string]any)}
		return u1.GetNamespace() < u2.GetNamespace()
	})
	sort.Slice(rv, func(i, j int) bool {
		u1 := &unstructured.Unstructured{Object: rv[i].(map[string]any)}
		u2 := &unstructured.Unstructured{Object: rv[j].(map[string]any)}
		return u1.GetName() < u2.GetName()
	})
	sort.Slice(rv, func(i, j int) bool {
		u1 := &unstructured.Unstructured{Object: rv[i].(map[string]any)}
		u2 := &unstructured.Unstructured{Object: rv[j].(map[string]any)}
		return u1.GroupVersionKind().String() < u2.GroupVersionKind().String()
	})
	return rv
}

func decodeExpected(t *testing.T, dir string) []any {
	f, err := os.Open(filepath.Join(dir, "expected.yaml"))
	qt.Assert(t, err, qt.IsNil)
	t.Cleanup(func() {
		qt.Assert(t, f.Close(), qt.IsNil)
	})
	return decodeStream(t, f)
}
