package path

import (
	"bytes"
	"context"
	"encoding/json"
	goerr "errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

type GitHubOption func(gh *github)

func WithGitHubToken(token string) GitHubOption {
	return func(gh *github) {
		gh.token = token
	}
}

func WithGitHubRepoRef(ref string) GitHubOption {
	return func(gh *github) {
		gh.sha = ref
	}
}

func WithGitHubHTTPClient(client HTTPClient) GitHubOption {
	return func(gh *github) {
		gh.client = client
	}
}

func WithGitHubCache(path Path) GitHubOption {
	return func(gh *github) {
		gh.cache = path
	}
}

func NewGitHub(owner, repo string, opts ...GitHubOption) Path {
	if owner == "" || repo == "" {
		panic("BUG: must provide values for owner and repo")
	}

	g := &github{
		client:  http.DefaultClient,
		token:   os.Getenv("GITHUB_TOKEN"),
		owner:   owner,
		repo:    repo,
		sha:     "",
		path:    make([]string, 0),
		history: make([]string, 0),
		tree:    make(map[string]githubTreeListItem),
		cache:   NewFSPath(afero.NewMemMapFs(), "/tmp"),
	}
	for _, f := range opts {
		f(g)
	}
	return g
}

type githubTreeListItem struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
	Type string `json:"type"`
	SHA  string `json:"sha"`
	Size int    `json:"size"`
	URL  string `json:"url"`
}

type github struct {
	// repo  GitHubRepo
	client httpClient
	token  string
	owner  string
	repo   string
	sha    string

	path    []string
	history []string

	tree map[string]githubTreeListItem

	cache Path
}

func (g *github) init() error {
	return g.doInit()
}

func (g *github) doInit() error {
	if len(g.tree) > 0 {
		return nil
	}
	return initGithub(g)
}

func (g *github) IsDir() (bool, error) {
	if err := g.doInit(); err != nil {
		return false, err
	}
	if len(g.path) == 0 {
		return true, nil
	}
	ok, err := g.Exists()
	if err != nil {
		return false, err
	}
	if !ok {
		return false, os.ErrNotExist
	}

	key := strings.Join(g.path, "/")
	info, ok := g.tree[key]
	if !ok {
		panic("BUG: Exists didn't return false but the tree is missing the path key")
	}
	// so far, I've only see types tree and blob,
	return info.Type == "tree", nil
}

func (g *github) Exists() (bool, error) {
	if err := g.doInit(); err != nil {
		return false, err
	}

	key := strings.Join(g.path, "/")
	if key == "" {
		// the root of the repo always exists
		return true, nil
	}
	_, ok := g.tree[key]
	return ok, nil
}

func (g *github) String() string {
	u := newGithubURL()
	u.Path = path.Join(g.owner, g.repo+".git")
	if len(g.path) > 0 {
		u.Path = u.Path + "/" + strings.Join(g.path, "/")
	}
	u.RawQuery = makeURLQuery(map[string]string{"ref": g.sha})
	return u.String()
}

func (g *github) Join(parts ...string) Path {
	cur := g.path
	for _, part := range parts {
		items := strings.Split(strings.Trim(part, "/"), "/")
		for _, item := range items {
			if len(item) > 0 {

				switch item {
				case "..":
					if len(cur) > 0 {
						// remove the last element
						cur = cur[:len(cur)-1]
					}
				case ".":
					// do nothing
				default:
					cur = append(cur, item)
				}
			}
		}
	}

	// Copy the history and the tree over to the new path object.
	// Copying once should allow the child to not require requesting
	// the tree a second time, which won't change because we're passing
	// around the SHA
	return &github{
		client:  g.client,
		token:   g.token,
		owner:   g.owner,
		repo:    g.repo,
		sha:     g.sha,
		history: append([]string{}, g.history...),
		tree:    g.tree,
		cache:   g.cache,
		path:    cur,
	}
}

func (g *github) Open() (io.Reader, error) {
	if err := g.doInit(); err != nil {
		return nil, err
	}
	// check the cache before trying to reread from the remote
	target := g.cache.Join(g.owner, g.repo, g.sha, filepath.Join(g.path...))
	ok, err := target.Exists()
	if err == nil && ok {
		return target.Open()
	}

	if ok, err := g.Exists(); err != nil || !ok {
		return nil, err
	}

	key := strings.Join(g.path, "/")
	info := g.tree[key]
	if info.Type != "blob" {
		// blob is a file, tree is a directory
		return nil, errors.Wrap(os.ErrInvalid, "path is not a file")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	u := newGithubRawURL()
	u.Path = path.Join(g.owner, g.repo, g.sha, strings.Join(g.path, "/"))

	req, err := newGithubRequest(ctx, u.String(), g.token)
	if err != nil {
		return nil, err
	}

	res, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}

	switch res.StatusCode {
	case http.StatusNotFound:
		panic("BUG: path.Exists returned found, but the response is saying the resource is not found")
	default:
		defer func() {
			_ = res.Body.Close()
		}()
		data, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		r := bytes.NewReader(data)
		if w, ok := target.(Writer); ok {
			f, err := w.Create()
			if err != nil {
				return nil, err
			}
			if _, err := io.Copy(f, bytes.NewReader(data)); err != nil {
				return nil, err
			}
		}
		return r, nil
	}
}

// newGithubRawURL returns the parsed url for GitHub's api endpoint
// or panics
func newGithubAPIURL() *url.URL {
	u, err := url.Parse("https://api.github.com/")
	if err != nil {
		panic(err)
	}
	return u
}

// newGithubRawURL returns the parsed url for GitHub's raw user content
// or panics
func newGithubRawURL() *url.URL {
	u, err := url.Parse("https://raw.githubusercontent.com")
	if err != nil {
		panic(err)
	}
	return u
}

// newGithubURL returns the parsed url for GitHub or panics
func newGithubURL() *url.URL {
	u, err := url.Parse("https://github.com")
	if err != nil {
		panic(err)
	}
	return u
}

// newGithubRequest creates a GET request with the required GitHub api headers, including
// the Authorization header if a token is provided. If a nil context is provided, a context
// with a timeout of 1 second will be used
func newGithubRequest(ctx context.Context, url, token string) (*http.Request, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
		defer cancel()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return req, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req, nil
}

// makeURLQuery turns the provided map into an encoded url query. Make URL
// query doesn't support adding multiple values for the same query key.
func makeURLQuery(in map[string]string) string {
	m := make(url.Values)
	for key, value := range in {
		m.Add(key, value)
	}
	return m.Encode()
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func initGithub(g *github) error {

	ref := g.sha
	if g.sha == "" {
		// if the SHA is empty, get the sha for the default branch
		u := newGithubAPIURL()
		u.Path = fmt.Sprintf("repos/%v/%v", g.owner, g.repo)

		var body struct {
			Branch string `json:"default_branch"`
		}

		err := func() error {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()

			req, err := newGithubRequest(ctx, u.String(), g.token)
			if err != nil {
				return err
			}

			res, err := g.client.Do(req)
			if err != nil {
				return err
			}
			defer func() {
				_ = res.Body.Close()
			}()

			if res.StatusCode != http.StatusOK {
				err := goerr.New(http.StatusText(res.StatusCode))
				if res.StatusCode == http.StatusNotFound {
					err = goerr.Join(os.ErrNotExist, err)
				}
				return err
			}

			if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
				return err
			}
			ref = body.Branch
			return nil
		}()
		if err != nil {
			return errors.Wrapf(err, ErrGitHubGetDefaultBranch)
		}
	}
	if ref == "" {
		panic("BUG: must be ignoring an error somewhere")
	}

	u := newGithubAPIURL()
	u.Path = fmt.Sprintf("repos/%s/%s/commits/%s", g.owner, g.repo, ref)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	req, err := newGithubRequest(ctx, u.String(), g.token)
	if err != nil {
		return errors.Wrapf(err, ErrGitHubGetCommit)
	}
	res, err := g.client.Do(req)
	if err != nil {
		return errors.Wrapf(err, ErrGitHubGetCommit)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	var body struct {
		SHA string `json:"sha"`
	}

	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return errors.Wrapf(err, ErrGitHubGetCommit)
	}

	g.sha = body.SHA
	sha := g.sha

	u = newGithubAPIURL()
	u.Path = fmt.Sprintf("/repos/%s/%s/git/trees/%s", g.owner, g.repo, sha)
	u.RawQuery = makeURLQuery(map[string]string{"recursive": "true"})

	req, err = newGithubRequest(ctx, u.String(), g.token)
	if err != nil {
		return errors.Wrapf(err, ErrGitHubGetWorktree)
	}

	res, err = g.client.Do(req)
	if err != nil {
		return errors.Wrapf(err, ErrGitHubGetWorktree)
	}
	defer func() {
		if err = res.Body.Close(); err != nil {
			panic(errors.Wrapf(err, "failed to close response body"))
		}
	}()

	switch res.StatusCode {
	case http.StatusUnprocessableEntity:
		return os.ErrInvalid
	case http.StatusNotFound:
		return os.ErrNotExist
	default:
		var body struct {
			SHA       string               `json:"sha"`
			URL       string               `json:"url"`
			Tree      []githubTreeListItem `json:"tree"`
			Truncated bool                 `json:"truncated"`
		}
		if err = json.NewDecoder(res.Body).Decode(&body); err != nil {
			return err
		}
		if g.tree == nil {
			g.tree = make(map[string]githubTreeListItem)
		}
		for _, item := range body.Tree {
			if item.Type != "tree" && item.Type != "blob" {
				panic("BUG: unrecognized tree item type: " + item.Type)
			}
			g.tree[item.Path] = item
		}
	}
	return nil
}
