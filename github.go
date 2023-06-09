package kustomize

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	gh "github.com/google/go-github/v53/github"
	"github.com/pkg/errors"
)

var (
	_ GitHub         = &github{}
	_ GitHubRepo     = &githubRepo{}
	_ GithubFileInfo = &githubFileInfo{}
)

// GitHub resolves references to upstream GitHub repos, downloads
// kustomize files from the repo, and clones repos to a local tree
type GitHub interface {
	GetDefaultBranch(ctx context.Context, owner, repo string) (string, error)
	GetCommitSha(ctx context.Context, owner, repo, ref string) (string, error)

	Repo(owner, repo string) GitHubRepo
}

type GitHubRepo interface {
	GetDefaultBranch(ctx context.Context) (string, error)
	GetCommitSha(ctx context.Context, ref string) (string, error)
	Owner() string
	Name() string

	ListDir(ctx context.Context, path string, ref string) ([]GithubFileInfo, error)
	OpenFile(ctx context.Context, path string, ref string) (io.Reader, error)
	FileExists(ctx context.Context, path string, ref string) (bool, error)
}

type GithubFileInfo interface {
	Name() string
	Size() int
	IsDir() bool
	Path() string
	Owner() string
	Repo() string
}

// NewGitHub returns a new GitHub client
func NewGitHub(token string) GitHub {
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}

	transport := http.DefaultTransport
	if token != "" {
		transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
			req.Header.Set("Authorization", "Basic "+token)
			return http.DefaultTransport.RoundTrip(req)
		})
	}
	return &github{
		mu:    &sync.Mutex{},
		cache: make(map[string]string),
		gh: gh.NewClient(&http.Client{
			Transport: transport,
		}),
	}
}

type github struct {
	gh *gh.Client

	mu    *sync.Mutex
	cache map[string]string
}

func (g *github) Repo(owner, repo string) GitHubRepo {
	return newGithubRepo(g.gh, g, owner, repo)
}

func (g *github) GetDefaultBranch(ctx context.Context, owner, repo string) (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.cache == nil {
		g.cache = make(map[string]string)
	}
	key := owner + "|" + repo
	branch, ok := g.cache[key]
	if !ok {
		r, resp, err := g.gh.Repositories.Get(ctx, owner, repo)
		if err != nil {
			return "", err
		}
		if resp.StatusCode != http.StatusOK {
			return "", errors.Wrap(ErrGitHubGetDefaultBranch, http.StatusText(resp.StatusCode))
		}
		g.cache[key] = r.GetDefaultBranch()
		branch = g.cache[key]
	}
	return branch, nil
}

func (g *github) GetCommitSha(ctx context.Context, owner, repo, ref string) (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.cache == nil {
		g.cache = make(map[string]string)
	}
	key := owner + "|" + repo + "|" + ref
	if sha, ok := g.cache[key]; ok {
		return sha, nil
	}

	commit, resp, err := g.gh.Repositories.GetCommit(ctx, owner, repo, ref, &gh.ListOptions{})
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.Wrap(ErrGitHubGetCommit, http.StatusText(resp.StatusCode))
	}
	sha := commit.GetSHA()
	g.cache[key] = sha
	return sha, nil
}

func newGithubRepo(gh *gh.Client, g GitHub, owner, repo string) *githubRepo {
	return &githubRepo{
		gh:     gh,
		github: g,
		owner:  owner,
		repo:   repo,
	}
}

type githubRepo struct {
	github GitHub
	gh     *gh.Client
	owner  string
	repo   string
}

func (g *githubRepo) Owner() string {
	return g.owner
}

func (g *githubRepo) Name() string {
	return g.repo
}

func (g *githubRepo) GetDefaultBranch(ctx context.Context) (string, error) {
	return g.github.GetDefaultBranch(ctx, g.owner, g.repo)
}

func (g *githubRepo) GetCommitSha(ctx context.Context, ref string) (string, error) {
	return g.github.GetCommitSha(ctx, g.owner, g.repo, ref)
}

func (g *githubRepo) ListDir(ctx context.Context, path, ref string) ([]GithubFileInfo, error) {
	_, dc, resp, err := g.gh.Repositories.GetContents(ctx, g.owner, g.repo, path, &gh.RepositoryContentGetOptions{
		Ref: ref,
	})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("Failed to get directory content: %s", http.StatusText(resp.StatusCode))
	}
	if dc == nil {
		return nil, errors.Errorf("directory response empty: %s", http.StatusText(resp.StatusCode))
	}
	rv := make([]GithubFileInfo, 0, len(dc))
	for _, item := range dc {
		rv = append(rv, &githubFileInfo{
			name:  item.GetName(),
			size:  item.GetSize(),
			typ:   item.GetType(),
			path:  item.GetPath(),
			owner: g.owner,
			repo:  g.repo,
		})
	}
	return rv, nil
}

func (g *githubRepo) OpenFile(ctx context.Context, path, ref string) (io.Reader, error) {

	// find the part directory and list the contents, then get the download link from there?

	fc, dc, resp, err := g.gh.Repositories.GetContents(ctx, g.owner, g.repo, path, &gh.RepositoryContentGetOptions{
		Ref: ref,
	})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, os.ErrNotExist
		}
		return nil, errors.Errorf("Failed to get file content: %s", http.StatusText(resp.StatusCode))
	}
	if dc != nil || fc == nil {
		return nil, ErrGitHubNotAFile
	}
	content, err := fc.GetContent()
	if err != nil {
		return nil, err
	}
	return strings.NewReader(content), nil
}

func (g *githubRepo) FileExists(ctx context.Context, path string, ref string) (bool, error) {
	ctx = context.Background()
	_, _, resp, err := g.gh.Repositories.GetContents(ctx, g.owner, g.repo, path, &gh.RepositoryContentGetOptions{
		Ref: ref,
	})
	if err != nil {
		// The error can be as a result of an error response
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, err
	}
	// file exists
	if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	return false, errors.New(http.StatusText(resp.StatusCode))
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

var _ GithubFileInfo = githubFileInfo{}

type githubFileInfo struct {
	name  string
	size  int
	typ   string
	path  string
	owner string
	repo  string
}

func (g githubFileInfo) Name() string {
	return g.name
}

func (g githubFileInfo) Size() int {
	return g.size
}

func (g githubFileInfo) IsDir() bool {
	return g.typ == "dir"
}

func (g githubFileInfo) Path() string {
	return g.path
}

func (g githubFileInfo) Owner() string {
	return g.owner
}

func (g githubFileInfo) Repo() string {
	return g.repo
}
