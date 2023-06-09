package kustomize

import (
	"github.com/spf13/afero"
	"os"
	"path/filepath"
	"strings"
)

// CacheDir resolves the global tree directory, but it doesn't clean
// up the directory
func CacheDir() Path { return cacheDir(afero.NewOsFs(), os.Environ()) }

func cacheDir(fs afero.Fs, environ []string) Path {
	if environ == nil {
		environ = os.Environ()
	}
	for _, env := range environ {
		if strings.HasPrefix(env, "KUSTOMIZE_CACHE_DIR=") {
			cd := strings.TrimPrefix(env, "KUSTOMIZE_CACHE_DIR=")
			return NewPath(fs, cd)
		}
	}
	return NewPath(fs, filepath.Join(os.TempDir(), "kustomize"))
}
