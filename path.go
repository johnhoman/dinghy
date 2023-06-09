package kustomize

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
	"gopkg.in/yaml.v3"
)

var (
	_ Path   = &fsPath{}
	_ Path   = &githubPath{}
	_ Reader = &reader{}
	_ Writer = &fsPath{}

	ErrGitHubGetDefaultBranch = errors.New("failed to get default branch")
	ErrGitHubGetCommit        = errors.New("failed to get commit for branch")
	ErrGitHubNotAFile         = errors.New("the requested content is not a file")
	ErrGitHubGetWorktree      = errors.New("the requested content is not a file")
)

type Reader interface {
	UnmarshalYAML(obj any) error
}

type Writer interface {
	WriteJSON(obj any) error
	WriteYAML(obj any) error
	Create() (io.Writer, error)
	WriteString(content string) error
	Copy(r io.Reader) error
	WriteBytes(b []byte) error
}

type Path interface {
	fmt.Stringer
	Join(path ...string) Path
	Open() (io.Reader, error)
	IsDir() (bool, error)
	Exists() (bool, error)
}

func NewPath(fs afero.Fs, path string) Path {
	return &fsPath{fs: fs, cur: path}
}

// NewMemoryPath returns a Path implementation backed by an
// in memory filesystem
func NewMemoryPath(path string) Path {
	return &fsPath{fs: afero.NewMemMapFs(), cur: path}
}

// NewOsPath returns a new Path implementation backed by the
// local filesystem fs
func NewOsPath(path string) Path {
	return &fsPath{fs: afero.NewOsFs(), cur: path}
}

// NewGitHubPathWithClient returns a GitHub path implementation using the provided
// client
func NewGitHubPathWithClient(owner, repo, ref, token string, client httpClient) (Path, error) {
	if owner == "" || repo == "" {
		return nil, errors.New("cannot create GitHub path without repo and owner")
	}
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}

	g := &githubPath{
		client:  client,
		token:   token,
		owner:   owner,
		repo:    repo,
		sha:     ref,
		path:    make([]string, 0),
		history: make([]string, 0),
		tree:    make(map[string]githubTreeListItem),
		cache:   CacheDir(),
	}

	if ref == "" {
		// if the SHA is empty, get the sha for the default branch
		u := newGithubAPIURL()
		u.Path = fmt.Sprintf("repos/%v/%v", owner, repo)

		var body struct {
			Branch string `json:"default_branch"`
		}

		err := func() error {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()

			req, err := newGithubRequest(ctx, u.String(), token)
			if err != nil {
				return err
			}

			res, err := client.Do(req)
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
			return nil, goerr.Join(ErrGitHubGetDefaultBranch, err)
		}
	}
	if ref == "" {
		panic("BUG: must be ignoring an error somewhere")
	}

	u := newGithubAPIURL()
	u.Path = fmt.Sprintf("repos/%s/%s/commits/%s", owner, repo, ref)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	req, err := newGithubRequest(ctx, u.String(), g.token)
	if err != nil {
		return nil, goerr.Join(ErrGitHubGetCommit, err)
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, goerr.Join(ErrGitHubGetCommit, err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	var body struct {
		SHA string `json:"sha"`
	}

	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return nil, goerr.Join(ErrGitHubGetCommit, err)
	}

	g.sha = body.SHA
	sha := g.sha

	u = newGithubAPIURL()
	u.Path = fmt.Sprintf("/repos/%s/%s/git/trees/%s", owner, repo, sha)
	u.RawQuery = makeURLQuery(map[string]string{"recursive": "true"})

	req, err = newGithubRequest(ctx, u.String(), token)
	if err != nil {
		return nil, err
	}

	res, err = client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err = res.Body.Close(); err != nil {
			return
		}
	}()

	switch res.StatusCode {
	case http.StatusUnprocessableEntity:
		err = os.ErrInvalid
	case http.StatusNotFound:
		err = os.ErrNotExist
	default:
		var body struct {
			SHA       string               `json:"sha"`
			URL       string               `json:"url"`
			Tree      []githubTreeListItem `json:"tree"`
			Truncated bool                 `json:"truncated"`
		}
		if err = json.NewDecoder(res.Body).Decode(&body); err != nil {
			return nil, err
		}
		if g.tree == nil {
			g.tree = make(map[string]githubTreeListItem)
		}
		// TODO: index tree?
		for _, item := range body.Tree {
			if item.Type != "tree" && item.Type != "blob" {
				panic("BUG: unrecognized tree item type: " + item.Type)
			}
			g.tree[item.Path] = item
		}
	}
	return g, nil
}

// NewGitHubPath returns a Path implementation for a commit in
// a single GitHub repo.
func NewGitHubPath(owner, repo, sha, token string) (Path, error) {
	return NewGitHubPathWithClient(owner, repo, sha, token, http.DefaultClient)
}

func NewReader(path Path) Reader {
	return &reader{path: path}
}

type fsPath struct {
	fs  afero.Fs
	cur string
}

func (f *fsPath) Copy(r io.Reader) error {
	fp, err := f.Create()
	if err != nil {
		return err
	}
	_, err = io.Copy(fp, r)
	return err
}

func (f *fsPath) WriteBytes(b []byte) error {
	fp, err := f.Create()
	if err != nil {
		return err
	}
	_, err = fp.Write(b)
	return err
}

func (f *fsPath) WriteJSON(obj any) error {
	fp, err := f.Create()
	if err != nil {
		return err
	}
	return json.NewEncoder(fp).Encode(obj)
}

func (f *fsPath) WriteYAML(obj any) error {
	fp, err := f.Create()
	if err != nil {
		return err
	}
	return yaml.NewEncoder(fp).Encode(obj)
}

func (f *fsPath) WriteString(content string) error {
	w, err := f.Create()
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, content)
	return err
}

// Create opens a new file (or truncates an existing one) for writing and
// creates an intermediate directories in the path.
func (f *fsPath) Create() (io.Writer, error) {

	parent := f.Join("..").(*fsPath)
	if err := f.fs.MkdirAll(parent.cur, 0700); err != nil {
		return nil, err
	}
	fp, err := f.fs.Create(f.cur)
	if err != nil {
		return nil, err
	}
	return fp, nil
}

func (f *fsPath) IsDir() (bool, error) {
	ok, err := f.Exists()
	if err != nil {
		return false, err
	}
	if !ok {
		return false, os.ErrNotExist
	}

	info, err := f.fs.Stat(f.cur)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
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
	fp, err := fs.Open(f.cur)
	if err != nil {
		return nil, err
	}
	return fp, nil
}

type githubTreeListItem struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
	Type string `json:"type"`
	SHA  string `json:"sha"`
	Size int    `json:"size"`
	URL  string `json:"url"`
}

type githubPath struct {
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

func (g *githubPath) IsDir() (bool, error) {
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
		panic("BUG: g.Exists didn't return false but the tree is missing the path key")
	}
	// so far, I've only see types tree and blob,
	return info.Type == "tree", nil
}

func (g *githubPath) Exists() (bool, error) {
	key := strings.Join(g.path, "/")
	if key == "" {
		// the root of the repo always exists
		return true, nil
	}
	_, ok := g.tree[key]
	return ok, nil
}

func (g *githubPath) String() string {
	u := newGithubURL()
	u.Path = path.Join(g.owner, g.repo+".git")
	if len(g.path) > 0 {
		u.Path = u.Path + "/" + strings.Join(g.path, "/")
	}
	u.RawQuery = makeURLQuery(map[string]string{"ref": g.sha})
	return u.String()
}

func (g *githubPath) Join(parts ...string) Path {
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

	history := make([]string, 0, len(g.history))
	for _, h := range g.history {
		history = append(history, h)
	}
	// Copy the history and the tree over to the new path object.
	// Copying once should allow the child to not require requesting
	// the tree a second time, which won't change because we're passing
	// around the SHA
	return &githubPath{
		client:  g.client,
		token:   g.token,
		owner:   g.owner,
		repo:    g.repo,
		sha:     g.sha,
		history: g.history,
		tree:    g.tree,
		cache:   g.cache,
		path:    cur,
	}
}

func (g *githubPath) Open() (io.Reader, error) {
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

type reader struct {
	path Path
}

func (r *reader) UnmarshalYAML(obj any) error {
	f, err := r.path.Open()
	if err != nil {
		return err
	}
	return yaml.NewDecoder(f).Decode(obj)
}
