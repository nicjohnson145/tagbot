package git

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/require"
)

type testcommit struct {
	Msg string
	Tag string
}

func newMemoryTestRepo(t *testing.T, commits ...testcommit) testRepo {
	t.Helper()

	fs := memfs.New()

	repo, err := git.Init(memory.NewStorage(), fs)
	require.NoError(t, err, "creating repo")

	testRepo := testRepo{
		Repo: repo,
		fs:   fs,
		filename: "commit-message",
	}

	for _, commit := range commits {
		testRepo.makeCommit(t, commit)
	}

	return testRepo
}

type testRepo struct {
	Repo     *git.Repository
	fs       billy.Filesystem
	filename string
	Dir      string
}

func (r *testRepo) makeCommit(t *testing.T, commit testcommit) plumbing.Hash {
	w, err := r.Repo.Worktree()
	require.NoError(t, err, "getting work tree")

	f, err := r.fs.OpenFile(r.filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	require.NoError(t, err, "opening file in memory")

	_, err = io.Copy(f, strings.NewReader(commit.Msg))
	require.NoError(t, err, "copying file contents")
	err = f.Close()
	require.NoError(t, err, "closing file")

	_, err = w.Add(r.filename)
	require.NoError(t, err, "adding file to work tree")

	hash, err := w.Commit(commit.Msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "tagbot",
			Email: "tagbot@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err, "commiting file")

	if commit.Tag != "" {
		head, err := r.Repo.Head()
		require.NoError(t, err, "getting repo head")

		_, err = r.Repo.CreateTag(commit.Tag, head.Hash(), nil)
		require.NoError(t, err, "creating tag")
	}

	return hash
}

func TestLatestTag(t *testing.T) {
	repo := newMemoryTestRepo(
		t,
		testcommit{
			Msg: "foobar1",
			Tag: "v0.1.0",
		},
		testcommit{
			Msg: "foobar3",
			Tag: "v0.1.1",
		},
		testcommit{
			Msg: "foobar3",
			Tag: "some-non-semver-tag",
		},
	)

	g := GitRepo{
		repo: repo.Repo,
	}

	tag, err := g.LatestTag()
	require.NoError(t, err, "getting latest tag")
	require.NotNil(t, tag)
	require.Equal(t, "v0.1.1", tag.Tag.Original())
}

func TestCommitsSinceTag(t *testing.T) {
	repo := newMemoryTestRepo(t)
	tagHash := repo.makeCommit(t, testcommit{
		Msg: "foobar1",
		Tag: "v0.1.0",
	})
	repo.makeCommit(t, testcommit{
		Msg: "some commit 1",
	})
	repo.makeCommit(t, testcommit{
		Msg: "some commit 2",
	})

	g := GitRepo{
		repo: repo.Repo,
	}

	commitsSince, err := g.CommitsSinceHash(tagHash.String())
	require.NoError(t, err)
	require.Equal(t, []string{"some commit 1", "some commit 2"}, commitsSince)
}

func TestReverseArray(t *testing.T) {
	testData := []struct {
		name   string
		input  []int
		output []int
	}{
		{
			name: "multiple_even",
			input: []int{1, 2, 3, 4},
			output: []int{4, 3, 2, 1},
		},
		{
			name: "multiple_odd",
			input: []int{1, 2, 3},
			output: []int{3, 2, 1},
		},
		{
			name: "one",
			input: []int{1},
			output: []int{1},
		},
		{
			name: "zero",
			input: []int{},
			output: []int{},
		},
	}
	for _, tc := range testData {
		t.Run(tc.name, func(t *testing.T) {
			out := reverseArray(tc.input)
			require.Equal(t, tc.output, out)
		})
	}
}
