package path

import (
	"bytes"
	"github.com/pkg/errors"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var (
	_ impl = Local{}
	_ impl = &GitHub{}
	_ impl = Memory{}
)

const (
	ErrGithubURL = "invalid GitHub URL format: expected 'https://github.com/owner/repo/path/to'"
)

type impl interface {
	ReadFile(path string) ([]byte, error)
	IsDir(path string) (bool, error)
	join(root string, segments ...string) string
	toString(root string, segments ...string) string
}

func Exists(in impl, path string) (bool, error) {
	if _, err := in.ReadFile(path); err == nil {
		return true, nil
	}
	return in.IsDir(path)
}

func ReadText(in impl, path string) (string, error) {
	data, err := in.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func ReadBytes(in impl, path string) ([]byte, error) {
	return in.ReadFile(path)
}

func Reader(in impl, path string) (io.Reader, error) {
	b, err := in.ReadFile(path)
	return bytes.NewReader(b), err
}

func IsRelative(path string) bool {
	p, err := Parse(path)
	if err == nil {
		// a local path didn't match on anything like GitHub or s3 during parsing,
		// so assume it's a relative path
		_, ok := p.path.(Local)
		return ok && !filepath.IsAbs(path)
	}
	// if there was an error, it's likely a parse error, so it wouldn't be relative
	// because relative paths don't match on URLs
	return false
}

func MustParse(in string) Path {
	parsed, err := Parse(in)
	if err != nil {
		panic(err)
	}
	return parsed
}

// Parse an input path. Paths can be URLs, s3 paths, or local paths
func Parse(in string) (Path, error) {
	switch {
	case strings.HasPrefix(in, "https://github.com"):
		u, err := url.Parse(in)
		if err != nil {
			return Path{}, errors.Wrapf(err, "failed to parse GitHub URL: %q", in)
		}

		owner, path, ok := strings.Cut(strings.TrimPrefix(u.Path, "/"), "/")
		if !ok {
			return Path{}, errors.Errorf("%s: %q", ErrGithubURL, in)
		}

		repo, path, _ := strings.Cut(path, "/")
		if repo == "" {
			return Path{}, errors.Errorf("%s: %q", ErrGithubURL, in)
		}
		ref := u.Query().Get("ref")
		return Path{
			path: NewGitHub(owner, repo, ref, os.Getenv("GITHUB_TOKEN")),
			root: path,
		}, nil
	case strings.HasPrefix(in, "https://"):
		return Path{}, errors.Errorf("URLs sources are unsupported at this time, but will be supported in future releases: %q", in)
	case strings.HasPrefix(in, "s3://"):
		return Path{}, errors.Errorf("s3 is unsupported at this time, but will be supported in future releases: %q", in)
	case strings.HasPrefix(in, "github.com"):
		return Parse("https://" + in)
	case strings.HasPrefix(in, "memory://"):
		return Path{path: NewMemory(), root: strings.TrimPrefix(in, "memory://")}, nil
	default:
		return Path{path: NewLocal(), root: in}, nil
	}
}
