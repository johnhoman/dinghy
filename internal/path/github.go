package path

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sync"
)

type GitHub struct {
	Owner string
	Repo  string
	Ref   string
	Token string

	mu        sync.Mutex
	githubGet func(url, pat string) ([]byte, error)
}

func (g *GitHub) toString(root string, segments ...string) string {
	u := &url.URL{}
	u.Scheme = "https"
	u.Path = path.Join(g.Owner, g.Repo, root, filepath.Join(segments...))
	u.Host = "github.com"
	if g.Ref != "" {
		q := u.Query()
		q.Set("ref", g.Ref)
		u.RawPath = q.Encode()
	}
	return u.String()
}

func NewGitHub(owner, repo, ref, token string) *GitHub {
	return &GitHub{
		Owner:     owner,
		Repo:      repo,
		Token:     token,
		Ref:       ref,
		mu:        sync.Mutex{},
		githubGet: githubGet,
	}
}

func (g *GitHub) join(root string, segments ...string) string {
	segments = append([]string{root}, segments...)
	return path.Join(root, path.Join(segments...))
}

// ReadFile reads the contents of the file at the specified file path from the GitHub repository.
// It retrieves the file content using the GitHub Raw API and caches it for subsequent calls.
// If the file is not found or any error occurs during the retrieval, it returns an error.
func (g *GitHub) ReadFile(filePath string) ([]byte, error) {
	// I need to resolve an empty sha to a commit.
	// Get the default branch, store the default branch name in the cache
	// Get the commit SHA using the ref and cache the commit sha
	if err := g.checkRef(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", g.Owner, g.Repo, g.Ref, filePath)
	return githubGet(url, g.Token)
}

func (g *GitHub) IsDir(path string) (bool, error) {
	tree, err := g.getGitHubTree()
	if err != nil {
		return false, err
	}

	for _, node := range tree.Tree {
		if node.Path == path {
			return node.Type == "tree", nil
		}
	}

	return false, os.ErrNotExist
}

func (g *GitHub) getGitHubTree() (githubTreeResponse, error) {

	var tree githubTreeResponse
	if err := g.checkRef(); err != nil {
		return tree, err
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees/%s", g.Owner, g.Repo, g.Ref)

	data, err := githubGet(url, g.Token)
	if err != nil {
		return tree, err
	}

	if err := json.Unmarshal(data, &tree); err != nil {
		return tree, err
	}

	return tree, nil
}

func (g *GitHub) checkRef() error {
	if g.Ref != "" {
		return nil
	}
	ref, err := getGitHubSHA(g.Owner, g.Repo, g.Ref, g.Token)
	if err != nil {
		return err
	}
	g.mu.Lock()
	g.Ref = ref
	g.mu.Unlock()
	return nil
}

var github struct {
	sync.RWMutex
	cache map[string][]byte
}

func init() {
	github.RWMutex = sync.RWMutex{}
	github.cache = make(map[string][]byte)
}

func githubGetDefaultBranch(owner, repo, pat string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	data, err := githubGet(url, pat)

	var body struct {
		DefaultBranch string `json:"default_branch"`
	}

	if err = json.Unmarshal(data, &body); err != nil {
		return "", err

	}
	return body.DefaultBranch, nil
}

// getGitHubSHA retrieves the SHA for a given ref of a GitHub repository
func getGitHubSHA(owner, repo, ref, pat string) (string, error) {
	if ref == "" {
		key := owner + "|" + repo
		github.RLock()
		branch, ok := github.cache[key]
		github.RUnlock()
		if !ok {
			branchName, err := githubGetDefaultBranch(owner, repo, pat)
			if err != nil {
				return "", errors.Wrapf(err, "failed to get default branch from github repo \"%s/%s\"", owner, repo)
			}
			github.Lock()
			branch = []byte(branchName)
			github.cache[key] = branch
			github.Unlock()
		}
		ref = string(branch)
	}

	key := owner + "|" + repo + "|" + ref
	github.RLock()
	sha, ok := github.cache[key]
	github.RUnlock()
	if ok {
		return string(sha), nil
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/refs/%s", owner, repo, ref)
	data, err := githubGet(url, pat)
	if err != nil {
		return "", err
	}

	var body struct {
		SHA string `json:"sha"`
	}

	if err := json.Unmarshal(data, &body); err != nil {
		return "", err
	}

	github.Lock()
	github.cache[key] = []byte(body.SHA)
	github.Unlock()
	return body.SHA, nil
}

func githubGet(url, pat string) ([]byte, error) {
	github.RLock()
	data, ok := github.cache[url]
	github.RUnlock()
	if ok {
		// convert the data into the worktree
		return data, nil
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if pat != "" {
		req.Header.Set("Authorization", "Bearer "+pat)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub request failed: %q", url)
	}

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	github.Lock()
	github.cache[url] = data
	github.Unlock()
	return data, nil
}

type githubTreeResponse struct {
	SHA  string `json:"sha"`
	URL  string `json:"url"`
	Tree []struct {
		Path string `json:"path"`
		Type string `json:"type"`
	} `json:"tree"`
	Truncated bool `json:"truncated"`
}
