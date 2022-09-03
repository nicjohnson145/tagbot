package main

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

type testcommit struct {
	Msg string
	Tag string
}

type testRepo struct {
	Repo     *git.Repository
	fs       billy.Filesystem
	filename string
	Dir      string
}

func newMemoryTestRepo(t *testing.T, commits ...testcommit) testRepo {
	t.Helper()
	return newTestRepo(t, memory.NewStorage(), memfs.New(), commits...)
}

func newDiskTestRepo(t *testing.T, commits ...testcommit) testRepo {
	t.Helper()
	dir := t.TempDir()
	fs := osfs.New(dir)
	repo := newTestRepo(t, filesystem.NewStorage(fs, cache.NewObjectLRUDefault()), fs, commits...)
	repo.Dir = dir
	return repo
}

func newTestRepo(t *testing.T, store storage.Storer, fs billy.Filesystem, commits ...testcommit) testRepo {
	t.Helper()

	repo, err := git.Init(store, fs)
	require.NoError(t, err, "creating repo")

	testRepo := testRepo{
		Repo: repo,
		fs:   fs,
		filename: "commit-message",
	}

	for _, commit := range commits {
		testRepo.MakeCommit(t, commit)
	}

	return testRepo
}

func (r *testRepo) MakeCommit(t *testing.T, commit testcommit) plumbing.Hash {
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
