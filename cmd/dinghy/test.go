package main

import (
	"bytes"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/johnhoman/dinghy/internal/build"
	"github.com/johnhoman/dinghy/internal/logging"
	"github.com/johnhoman/dinghy/internal/path"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
	"io"
	"io/fs"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"os"
	"path/filepath"
	"sort"
)

var (
	logger = logging.New()
)

type cmdTest struct {
	Root     string `kong:"name=dir,arg"`
	ShowDiff bool
	V        int `kong:"name=verbose,short=v,type=counter"`
}

func (cmd *cmdTest) Run() error {
	wd, err := os.Getwd()
	if err != nil {
		logger.Fatal("%s", err)
	}

	logger.SetPrefix("[TEST]")
	matches := make([]string, 0)
	err = filepath.Walk(cmd.Root, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() && isDinghyDir(path) {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return err
	}
	for _, match := range matches {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error(logging.UseGray("%s")+" %s", match, r)
					if cmd.V < 2 {
						logger.Debug("%s - use --v=2 to show trace", match)
					}
					if err, ok := r.(error); ok && cmd.V == 2 {
						logger.Error(fmt.Sprintf(logging.UseGray("%+v"), errors.WithStack(err)))
					}
				}
			}()

			logger.Debug(match)
			f, err := os.Open(filepath.Join(match, "expected.yaml"))
			if err != nil {
				logger.Fatal("%s", err)
			}
			expected, err := decodeReader(f)
			if err != nil {
				logger.Fatal("%s", err)
			}
			b := &cmdBuild{Dir: match}

			buf := new(bytes.Buffer)
			err = b.Run(afero.NewOsFs(), path.NewFSPath(afero.NewOsFs(), wd), buf)
			if err != nil {
				logger.Fatal("%s", err)
			}
			got, err := decodeReader(buf)
			if err != nil {
				logger.Fatal("%s", err)
			}
			if diff := cmp.Diff(got, expected); diff != "" {
				if cmd.ShowDiff {
					logger.Debug("Mismatch (-got,+want):\n", diff)
				}
			}
			logger.Ok(logging.UseGray(match))
		}()
	}
	return nil
}

var isDinghyDir = func(path string) bool {
	_, err := os.Stat(filepath.Join(path, build.DinghyFile))
	return err == nil
}

func decodeReader(r io.Reader) ([]any, error) {
	rv := make([]any, 0)
	d := yaml.NewDecoder(r)
	for {
		var obj map[string]any
		err := d.Decode(&obj)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
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
	return rv, nil
}

func unmarshalExpected(dir string) ([]any, error) {
	f, err := os.Open(filepath.Join(dir, "expected.yaml"))
	if err != nil {
		return nil, err
	}
	s, err := decodeReader(f)
	if err != nil {
		logger.Fatal("%s", err)
	}
	return s, nil
}
