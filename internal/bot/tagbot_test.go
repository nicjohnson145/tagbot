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
	"github.com/nicjohnson145/hlp"
	"github.com/nicjohnson145/tagbot/internal/config"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

type testCommit struct {
	Message string
	Tag     string
	Files   []string
}

type unitTestRepo struct {
	IRepo
	repo *gogit.Repository
	fs   billy.Filesystem
}

func (u *unitTestRepo) MakeCommits(t *testing.T, commits ...testCommit) {
	w, err := u.repo.Worktree()
	require.NoError(t, err, "getting work tree")

	for _, commit := range commits {
		// generate some random content, ensuring we get a change set even if we touch the same file
		content := ulid.Make().String()

		for _, file := range commit.Files {
			// ensure any containing directories are created
			dir := filepath.Dir(file)
			require.NoError(t, u.fs.MkdirAll(dir, 0755))

			// write out the updated contents to the file
			f, err := u.fs.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
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
		_, err := w.Commit(commit.Message, &gogit.CommitOptions{
			Author: &object.Signature{
				Name:  "tagbot",
				Email: "tagbot@example.com",
				When:  time.Now(),
			},
		})
		require.NoError(t, err)

		// if this commit should be associated with a tag, tag it as well
		if commit.Tag != "" {
			head, err := u.repo.Head()
			require.NoError(t, err)

			_, err = u.repo.CreateTag(commit.Tag, head.Hash(), nil)
			require.NoError(t, err)
		}
	}

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

func mustHaveTags(t *testing.T, repo *unitTestRepo, expected []string) {
	t.Helper()

	gotIter, err := repo.repo.Tags()
	require.NoError(t, err)
	gotTags := []string{}
	require.NoError(t, gotIter.ForEach(func(tag *plumbing.Reference) error {
		gotTags = append(gotTags, tag.Name().Short())
		return nil
	}))

	require.ElementsMatch(t, expected, gotTags)
}

func newCtxWithLog(t *testing.T) context.Context {
	logger := zerolog.New(zerolog.NewTestWriter(t))
	logger.Level(zerolog.TraceLevel)
	return logger.WithContext(context.Background())
}

func TestRun(t *testing.T) {
	t.Run("non monorepo, existing commits and tags", func(t *testing.T) {
		repo := newMemoryRepo(
			t,
			testCommit{
				Message: "fix: fix a thing",
				Tag:     "v0.1.0",
				Files: []string{
					"bar/qux.txt",
				},
			},
			testCommit{
				Message: "feat: do a thing",
				Files: []string{
					"foo/bar.txt",
				},
			},
		)
		bot := NewTagbot(TagbotConfig{
			MonorepoConfig: &config.MonoRepoConfig{
				Components: map[string]config.MonoRepoComponent{
					"core": {
						Name:           "core",
						ChangeSetGlobs: []string{"**/*"},
						Prefix:         hlp.Ptr(""),
						MaintainLatest: hlp.Ptr(false),
						LatestName:     hlp.Ptr("latest"),
						NoV:            hlp.Ptr(false),
						AlwaysPatch:    hlp.Ptr(false),
					},
				},
			},
			Repo: repo,
		})

		require.NoError(t, bot.Run(newCtxWithLog(t)))
		mustHaveTags(t, repo, []string{"v0.1.0", "v0.2.0"})
	})

	t.Run("monorepo, existing commits and tags", func(t *testing.T) {
		repo := newMemoryRepo(
			t,
			testCommit{
				Message: "fix: fix a thing",
				Tag:     "bar/v0.1.0",
				Files: []string{
					"bar/qux.txt",
				},
			},
			testCommit{
				Message: "feat: do a thing",
				Tag:     "foo/v0.1.0",
				Files: []string{
					"foo/bar.txt",
				},
			},
			testCommit{
				Message: "feat: do a thing",
				Files: []string{
					"foo/other.txt",
					"bar/other.txt",
				},
			},
		)
		bot := NewTagbot(TagbotConfig{
			MonorepoConfig: &config.MonoRepoConfig{
				Components: map[string]config.MonoRepoComponent{
					"foo": {
						Name:           "foo",
						ChangeSetGlobs: []string{"foo/*"},
						MaintainLatest: hlp.Ptr(false),
						LatestName:     hlp.Ptr("latest"),
						NoV:            hlp.Ptr(false),
						AlwaysPatch:    hlp.Ptr(false),
					},
					"bar": {
						Name:           "bar",
						ChangeSetGlobs: []string{"bar/*"},
						MaintainLatest: hlp.Ptr(false),
						LatestName:     hlp.Ptr("latest"),
						NoV:            hlp.Ptr(false),
						AlwaysPatch:    hlp.Ptr(false),
					},
				},
			},
			Repo: repo,
		})

		require.NoError(t, bot.Run(newCtxWithLog(t)))
		mustHaveTags(t, repo, []string{
			"foo/v0.1.0",
			"foo/v0.2.0",
			"bar/v0.1.0",
			"bar/v0.2.0",
		})
	})

	t.Run("monorepo always patch", func(t *testing.T) {
		repo := newMemoryRepo(
			t,
			testCommit{
				Message: "fix: fix a thing",
				Tag:     "bar/v0.1.0",
				Files: []string{
					"bar/qux.txt",
				},
			},
			testCommit{
				Message: "feat: do a thing",
				Tag:     "foo/v0.1.0",
				Files: []string{
					"foo/bar.txt",
				},
			},
			testCommit{
				Message: "do a thing",
				Files: []string{
					"foo/other.txt",
				},
			},
		)
		bot := NewTagbot(TagbotConfig{
			MonorepoConfig: &config.MonoRepoConfig{
				Components: map[string]config.MonoRepoComponent{
					"foo": {
						Name:           "foo",
						ChangeSetGlobs: []string{"foo/*"},
						MaintainLatest: hlp.Ptr(false),
						LatestName:     hlp.Ptr("latest"),
						NoV:            hlp.Ptr(false),
						AlwaysPatch:    hlp.Ptr(true),
					},
					"bar": {
						Name:           "bar",
						ChangeSetGlobs: []string{"bar/*"},
						MaintainLatest: hlp.Ptr(false),
						LatestName:     hlp.Ptr("latest"),
						NoV:            hlp.Ptr(false),
						AlwaysPatch:    hlp.Ptr(true),
					},
				},
			},
			Repo: repo,
		})
		require.NoError(t, bot.Run(newCtxWithLog(t)))
		mustHaveTags(t, repo, []string{
			"foo/v0.1.0",
			"foo/v0.1.1",
			"bar/v0.1.0",
		})
	})

	t.Run("non monorepo, empty", func(t *testing.T) {
		repo := newMemoryRepo(
			t,
			testCommit{
				Message: "feat: do a thing",
				Files: []string{
					"foo/bar.txt",
				},
			},
		)
		bot := NewTagbot(TagbotConfig{
			MonorepoConfig: &config.MonoRepoConfig{
				Components: map[string]config.MonoRepoComponent{
					"core": {
						Name:           "core",
						ChangeSetGlobs: []string{"**/*"},
						Prefix:         hlp.Ptr(""),
						MaintainLatest: hlp.Ptr(false),
						LatestName:     hlp.Ptr("latest"),
						NoV:            hlp.Ptr(false),
						AlwaysPatch:    hlp.Ptr(false),
					},
				},
			},
			Repo: repo,
		})

		require.NoError(t, bot.Run(newCtxWithLog(t)))
		mustHaveTags(t, repo, []string{"v0.0.1"})
	})

	t.Run("convert from single tag to monorepo", func(t *testing.T) {
		repo := newMemoryRepo(
			t,
			testCommit{
				Message: "im a message",
				Tag:     "v0.1.0",
				Files: []string{
					"bar/qux.txt",
				},
			},
			testCommit{
				Message: "do a thing",
				Files: []string{
					"foo/qux.txt",
				},
			},
		)
		bot := NewTagbot(TagbotConfig{
			MonorepoConfig: &config.MonoRepoConfig{
				Components: map[string]config.MonoRepoComponent{
					"foo": {
						Name:           "foo",
						ChangeSetGlobs: []string{"foo/*"},
						MaintainLatest: hlp.Ptr(false),
						LatestName:     hlp.Ptr("latest"),
						NoV:            hlp.Ptr(false),
						AlwaysPatch:    hlp.Ptr(false),
					},
					"bar": {
						Name:           "bar",
						ChangeSetGlobs: []string{"bar/*"},
						MaintainLatest: hlp.Ptr(false),
						LatestName:     hlp.Ptr("latest"),
						NoV:            hlp.Ptr(false),
						AlwaysPatch:    hlp.Ptr(false),
					},
				},
			},
			Repo: repo,
		})

		require.NoError(t, bot.Run(newCtxWithLog(t)))
		mustHaveTags(t, repo, []string{"v0.1.0", "foo/v0.0.1"})
	})
}

func TestCommitRelevantToComponent(t *testing.T) {
	t.Parallel()

	testData := []struct {
		name       string
		changeSets []string
		files      []string
		expected   bool
	}{
		{
			name: "anything",
			changeSets: []string{
				"*",
			},
			files: []string{
				"internal/bot/version.go",
			},
			expected: true,
		},
		{
			name: "subdir - no match",
			changeSets: []string{
				"foo/*",
			},
			files: []string{
				"internal/bot/version.go",
			},
			expected: false,
		},
		{
			name: "subdir - yes match",
			changeSets: []string{
				"internal/*",
			},
			files: []string{
				"internal/bot/version.go",
			},
			expected: false,
		},
	}
	for _, tc := range testData {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			bot := &Tagbot{}

			got, err := bot.commitRelevantToComponent(
				&config.MonoRepoComponent{
					ChangeSetGlobs: tc.changeSets,
				},
				&Commit{
					Files: tc.files,
				},
			)
			require.NoError(t, err)
			require.Equal(t, tc.expected, got)
		})
	}
}
