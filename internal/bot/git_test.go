package bot

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/require"
)

func TestGetFilesForCommit(t *testing.T) {

	t.Run("first commit", func(t *testing.T) {
		fs := memfs.New()
		repo, err := gogit.Init(memory.NewStorage(), fs)
		require.NoError(t, err)

		hashes := createCommits(
			t,
			repo,
			fs,
			testCommit{
				Message: "some message",
				Files:   []string{"dir/foo", "dir/bar"},
			},
		)

		c, err := repo.CommitObject(hashes[0])
		require.NoError(t, err)

		r := GitRepo{}
		got, err := r.getFilesForCommit(c)
		require.NoError(t, err)
		require.ElementsMatch(t, got, []string{"dir/foo", "dir/bar"})
	})

	t.Run("subsequent commit", func(t *testing.T) {
		fs := memfs.New()
		repo, err := gogit.Init(memory.NewStorage(), fs)
		require.NoError(t, err)

		hashes := createCommits(
			t,
			repo,
			fs,
			testCommit{
				Message: "some message",
				Files:   []string{"dir/foo", "dir/bar"},
			},
			testCommit{
				Message: "some other message",
				Files:   []string{"dir/bar", "dir/baz"},
			},
		)

		c, err := repo.CommitObject(hashes[1])
		require.NoError(t, err)

		r := GitRepo{}
		got, err := r.getFilesForCommit(c)
		require.NoError(t, err)
		require.ElementsMatch(t, got, []string{"dir/bar", "dir/baz"})
	})

	t.Run("add edit rename delete", func(t *testing.T) {
		fs := memfs.New()
		repo, err := gogit.Init(memory.NewStorage(), fs)
		require.NoError(t, err)

		worktree, err := repo.Worktree()
		require.NoError(t, err)

		// Add a file
		f, err := fs.OpenFile("foo", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		require.NoError(t, err)
		defer func() {
			_ = f.Close()
		}()
		_, err = io.Copy(f, strings.NewReader("content-one"))
		require.NoError(t, err)
		_, err = worktree.Add("foo")
		require.NoError(t, err)
		addCommit, err := worktree.Commit("some message", &gogit.CommitOptions{
			Author: &object.Signature{
				Name:  "tagbot",
				Email: "tagbot@example.com",
				When:  time.Now(),
			},
		})
		require.NoError(t, err)

		// edit the content of that file
		f, err = fs.OpenFile("foo", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		require.NoError(t, err)
		defer func() {
			_ = f.Close()
		}()
		_, err = io.Copy(f, strings.NewReader("content-two"))
		require.NoError(t, err)
		_, err = worktree.Add("foo")
		require.NoError(t, err)
		editCommit, err := worktree.Commit("some message", &gogit.CommitOptions{
			Author: &object.Signature{
				Name:  "tagbot",
				Email: "tagbot@example.com",
				When:  time.Now(),
			},
		})
		require.NoError(t, err)

		// rename the file, aka delete the old path, write the same content to the new path
		f, err = fs.OpenFile("bar", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		require.NoError(t, err)
		defer func() {
			_ = f.Close()
		}()
		require.NoError(t, fs.Remove("foo"))
		_, err = io.Copy(f, strings.NewReader("content-two"))
		require.NoError(t, err)
		_, err = worktree.Add("bar")
		require.NoError(t, err)
		_, err = worktree.Remove("foo")
		require.NoError(t, err)
		renameCommit, err := worktree.Commit("some message", &gogit.CommitOptions{
			Author: &object.Signature{
				Name:  "tagbot",
				Email: "tagbot@example.com",
				When:  time.Now(),
			},
		})
		require.NoError(t, err)

		// delete the file
		require.NoError(t, fs.Remove("bar"))
		_, err = worktree.Remove("bar")
		require.NoError(t, err)
		deleteCommit, err := worktree.Commit("some message", &gogit.CommitOptions{
			Author: &object.Signature{
				Name:  "tagbot",
				Email: "tagbot@example.com",
				When:  time.Now(),
			},
		})
		require.NoError(t, err)

		g := GitRepo{}

		t.Run("add", func(t *testing.T) {
			c, err := repo.CommitObject(addCommit)
			require.NoError(t, err)

			got, err := g.getFilesForCommit(c)
			require.NoError(t, err)
			require.ElementsMatch(t, got, []string{"foo"})
		})

		t.Run("edit", func(t *testing.T) {
			c, err := repo.CommitObject(editCommit)
			require.NoError(t, err)

			got, err := g.getFilesForCommit(c)
			require.NoError(t, err)
			require.ElementsMatch(t, got, []string{"foo"})
		})

		t.Run("rename", func(t *testing.T) {
			c, err := repo.CommitObject(renameCommit)
			require.NoError(t, err)

			got, err := g.getFilesForCommit(c)
			require.NoError(t, err)
			require.ElementsMatch(t, got, []string{"foo", "bar"})
		})

		t.Run("delete", func(t *testing.T) {
			c, err := repo.CommitObject(deleteCommit)
			require.NoError(t, err)

			got, err := g.getFilesForCommit(c)
			require.NoError(t, err)
			require.ElementsMatch(t, got, []string{"bar"})
		})
	})
}
