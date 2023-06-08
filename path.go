package kustomize

import (
	"context"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"time"

	"github.com/spf13/afero"
)

var (
	_ Path = &fsPath{}
	_ Path = &githubPath{}
)

type Path interface {
	fmt.Stringer
	Join(path ...string) Path
	Open() (io.Reader, error)
	Exists() (bool, error)
}

func NewPath(fs afero.Fs, path string) Path {
	return &fsPath{fs: fs, cur: path}
}

func NewPathGitHub(gh GitHubRepo, path, sha string) Path {
	if path == "/" {
		path = ""
	}
	return &githubPath{
		repo: gh,
		path: path,
		sha:  sha,
	}
}

type fsPath struct {
	fs  afero.Fs
	cur string
}

func (f *fsPath) Exists() (bool, error) {
	fs := afero.Afero{Fs: f.fs}
	if fs.Fs == nil {
		fs.Fs = afero.NewOsFs()
	}
	return fs.Exists(f.cur)
}

func (f *fsPath) String() string { return f.cur }
func (f *fsPath) Join(path ...string) Path {
	fs := f.fs
	if fs == nil {
		fs = afero.NewOsFs()
	}
	return NewPath(f.fs, filepath.Join(f.cur, filepath.Join(path...)))
}

func (f *fsPath) Open() (io.Reader, error) {
	fs := f.fs
	if fs == nil {
		fs = afero.NewOsFs()
	}
	return fs.Open(f.cur)
}

type githubPath struct {
	repo GitHubRepo
	path string
	sha  string
}

func (g *githubPath) Exists() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return g.repo.FileExists(ctx, g.path, g.sha)
}

func (g *githubPath) String() string {
	s := fmt.Sprintf("https://github.com/%s/%s.git", g.repo.Owner(), g.repo.Name())
	if g.path != "" {
		s = fmt.Sprintf("%s/%s", s, g.path)
	}
	if g.sha != "" {
		s = fmt.Sprintf("%s?ref=%s", s, g.sha)
	}
	return s
}

func (g *githubPath) Join(parts ...string) Path {
	rp := path.Join(g.path, path.Join(parts...))
	if rp == "/" {
		rp = ""
	}
	rv := &githubPath{repo: g.repo, path: rp, sha: g.sha}
	return rv
}

func (g *githubPath) Open() (io.Reader, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	sha := g.sha
	if sha == "" {
		branch, err := g.repo.GetDefaultBranch(ctx)
		if err != nil {
			return nil, err
		}
		sha, err = g.repo.GetCommitSha(ctx, branch)
	}

	return g.repo.OpenFile(ctx, g.path, sha)
}
