package main

import (
	"testing"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/require"
	"time"
	"io"
	"os"
	"strings"
)

type testcommit struct {
	Msg string
	Tag string
}

func makeTestRepo(t *testing.T, commits ...testcommit) *git.Repository {
	t.Helper()

	storer := memory.NewStorage()
	fs := memfs.New()

	repo, err := git.Init(storer, fs)
	require.NoError(t, err, "creating repo")

	w, err := repo.Worktree()
	require.NoError(t, err, "getting work tree")

	const fileName = "commit-msg"

	for _, commit := range commits {
		f, err := fs.OpenFile(fileName, os.O_RDWR | os.O_CREATE | os.O_TRUNC, 0666)
		require.NoError(t, err, "opening file in memory")

		_, err = io.Copy(f, strings.NewReader(commit.Msg))
		require.NoError(t, err, "copying file contents")
		err = f.Close()
		require.NoError(t, err, "closing file")

		_, err = w.Add(fileName)
		require.NoError(t, err, "adding file to work tree")

		_, err = w.Commit(commit.Msg, &git.CommitOptions{
			Author: &object.Signature{
				Name: "tagbot",
				Email: "tagbot@example.com",
				When: time.Now(),
			},
		})
		require.NoError(t, err, "commiting file")

		if commit.Tag == "" {
			continue
		}

		head, err := repo.Head()
		require.NoError(t, err, "getting repo head")

		_, err = repo.CreateTag(commit.Tag, head.Hash(), nil)
		require.NoError(t, err, "creating tag")
	}

	return repo
}
