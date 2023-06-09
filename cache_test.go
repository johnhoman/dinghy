package kustomize

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/spf13/afero"
)

func TestCacheDir(t *testing.T) {
	environ := []string{"KUSTOMIZE_CACHE_DIR=/home/jhoman/.tree/kustomize"}
	path := cacheDir(afero.NewMemMapFs(), environ)
	qt.Assert(t, path.String(), qt.Equals, "/home/jhoman/.tree/kustomize")
}
