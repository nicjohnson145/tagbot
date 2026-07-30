package bot

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

type testCommit struct {
	Message string
	Files   []string
	Tags    []string
}

type unitTestRepo struct {
	IRepo
	repo *gogit.Repository
	fs   billy.Filesystem
}

func (u *unitTestRepo) MakeCommits(t *testing.T, commits ...testCommit) []plumbing.Hash {
	t.Helper()

	return createCommits(t, u.repo, u.fs, commits...)
}

func (u *unitTestRepo) PushTags(ctx context.Context) error {
	return nil
}

func newMemoryRepo(t *testing.T, commits ...testCommit) *unitTestRepo {
	t.Helper()

	fs := memfs.New()

	repo, err := gogit.Init(memory.NewStorage(), fs)
	require.NoError(t, err)

	testRepo := &unitTestRepo{
		IRepo: &GitRepo{
			repo: repo,
		},
		repo: repo,
		fs:   fs,
	}
	testRepo.MakeCommits(t, commits...)

	return testRepo
}

func createCommits(t *testing.T, repo *gogit.Repository, fs billy.Filesystem, commits ...testCommit) []plumbing.Hash {
	t.Helper()

	w, err := repo.Worktree()
	require.NoError(t, err, "getting work tree")

	hashes := []plumbing.Hash{}

	for _, commit := range commits {
		// generate some random content, ensuring we get a change set even if we touch the same file
		content := ulid.Make().String()

		for _, file := range commit.Files {
			// ensure any containing directories are created
			dir := filepath.Dir(file)
			require.NoError(t, fs.MkdirAll(dir, 0755))

			// write out the updated contents to the file
			f, err := fs.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
			require.NoError(t, err)
			defer func() {
				_ = f.Close()
			}()
			_, err = io.Copy(f, strings.NewReader(content))
			require.NoError(t, err)

			// add the changes to the worktree
			_, err = w.Add(file)
			require.NoError(t, err)
		}

		// create a commit with the change set
		hash, err := w.Commit(commit.Message, &gogit.CommitOptions{
			Author: &object.Signature{
				Name:  "tagbot",
				Email: "tagbot@example.com",
				When:  time.Now(),
			},
		})
		require.NoError(t, err)
		hashes = append(hashes, hash)

		// if this commit should be associated with a tag, tag it as well
		if len(commit.Tags) > 0 {
			head, err := repo.Head()
			require.NoError(t, err)

			for _, tag := range commit.Tags {
				_, err = repo.CreateTag(tag, head.Hash(), nil)
				require.NoError(t, err)
			}
		}
	}

	return hashes
}
