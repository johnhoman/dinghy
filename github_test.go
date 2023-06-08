package kustomize

import (
	"context"
	"testing"

	qt "github.com/frankban/quicktest"
	yaml "gopkg.in/yaml.v3"
)

func TestGithub_GetDefaultBranch(t *testing.T) {
	gh := NewGitHub("")
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	branch, err := gh.GetDefaultBranch(ctx, "johnhoman", "nop")
	qt.Assert(t, err, qt.IsNil)
	qt.Assert(t, branch, qt.Equals, "main")
	qt.Assert(t, gh.(*github).cache["johnhoman|nop"], qt.Equals, "main")
}

func TestGithub_GetCommitSha(t *testing.T) {
	gh := NewGitHub("")
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	sha, err := gh.GetCommitSha(ctx, "johnhoman", "nop", "main")
	qt.Assert(t, err, qt.IsNil)
	qt.Assert(t, sha, qt.Equals, "76b46f2ecc5c896217f5cb5a0bfdf3346365050b")
	qt.Assert(t, gh.(*github).cache["johnhoman|nop|main"], qt.Equals, "76b46f2ecc5c896217f5cb5a0bfdf3346365050b")
}

func TestGithubRepo_Open(t *testing.T) {
	gh := NewGitHub("")
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	sha, err := gh.GetCommitSha(ctx, "johnhoman", "nop", "main")
	qt.Assert(t, err, qt.IsNil)

	r, err := gh.Repo("johnhoman", "nop").OpenFile(ctx, "1/kustomization.yaml", sha)
	qt.Assert(t, err, qt.IsNil)
	qt.Assert(t, r, qt.IsNotNil)

	var m map[string]any
	qt.Assert(t, yaml.NewDecoder(r).Decode(&m), qt.IsNil)
	qt.Assert(t, m, qt.DeepEquals, map[string]any{
		"apiVersion": "kustomize.config.k8s.io/v1beta1",
		"kind":       "Kustomization",
		"resources":  []any{"configmap.yaml"},
	})
}

func TestGithubRepo_ListDir(t *testing.T) {
	gh := NewGitHub("")
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	sha, err := gh.GetCommitSha(ctx, "johnhoman", "nop", "main")
	qt.Assert(t, err, qt.IsNil)

	contents, err := gh.Repo("johnhoman", "nop").ListDir(ctx, "1", sha)
	qt.Assert(t, err, qt.IsNil)

	qt.Assert(t, contents, qt.HasLen, 2)
}
