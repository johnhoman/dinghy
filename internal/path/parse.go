package path

import (
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"net/url"
	"regexp"
	"strings"
)

type Option func(o *options)

// WithRelativeRoot gives the parser a root as a base for relative paths,
// so that they can be resolved to absolute paths
func WithRelativeRoot(root Path) Option {
	return func(o *options) {
		o.root = root
	}
}

// PreparePath does any path initialization required, such as syncing
// the GitHub path worktree cache, or pulling s3 credentials from the
// host
func PreparePath(path Path) error {
	return path.init()
}

// Parse parses the provided path and returns the path object
// most closely aligned with the path.
func Parse(in string, opts ...Option) (Path, error) {
	o := newOptions(opts...)

	switch {
	case git.MatchString(in):
		u, err := url.Parse(requireSecure(in))
		if err != nil {
			return nil, errors.Wrapf(err, ErrParseGitHubURL)
		}
		owner, repo, ok := cutGitHubPath(u.Path)
		if !ok {
			return nil, errors.Wrapf(err, "%s: failed to parse owner/repo from URL path", ErrParseGitHubURL)
		}
		ref := u.Query().Get("ref")
		return NewGitHub(owner, repo, WithGitHubRepoRef(ref)), nil
	case s3.MatchString(in):
		panic("s3-path: not implemented")
	case remote.MatchString(in):
		panic("remote-path: not implemented")
	}
	if err := PreparePath(o.root); err != nil {
		return nil, err
	}
	p := o.root.Join(in)
	return p, PreparePath(p)
}

func requireSecure(in string) string {
	in = strings.TrimPrefix(in, "http://")
	if !strings.HasPrefix(in, "https://") {
		in = "https://" + in
	}
	return in
}

func cutGitHubPath(path string) (string, string, bool) {
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) >= 2 {
		owner := parts[0]
		repo := strings.TrimPrefix(parts[1], ".git")
		return owner, repo, true
	}
	return "", "", false
}

var (
	git    = regexp.MustCompile(`^(?:https:\/\/)?(?:www\.)?github\.com\b(?:[-a-zA-Z0-9()@:%_\+.~#?&\/=]*)$`)
	s3     = regexp.MustCompile(`^s3:\/\/[-a-z0-9@:%._\+~#=]{1,256}\b(?:[-a-zA-Z0-9()@:%_\+.~#?&\/=]*)$`)
	remote = regexp.MustCompile(`^https?:\/\/(?:www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b(?:[-a-zA-Z0-9()@:%_\+.~#?&\/=]*)$`)
)

func newOptions(opts ...Option) *options {
	o := &options{
		root: NewFSPath(afero.NewOsFs(), ""),
	}
	for _, f := range opts {
		f(o)
	}
	return o
}

type options struct {
	// root is the root of an existing file system to consider
	// when parsing a path. Any relative paths should be resolved
	// to their full path, so when parsing relative paths that don't
	// match on any full paths, it should be considered relative to
	// this path. If the path is not provided
	root Path
}
