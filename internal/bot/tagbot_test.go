package bot

import (
	"context"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/nicjohnson145/hlp"
	"github.com/nicjohnson145/tagbot/internal/config"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

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
				Tags:    []string{"v0.1.0"},
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
				Tags:    []string{"bar/v0.1.0"},
				Files: []string{
					"bar/qux.txt",
				},
			},
			testCommit{
				Message: "feat: do a thing",
				Tags:    []string{"foo/v0.1.0"},
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
				Tags:    []string{"bar/v0.1.0"},
				Files: []string{
					"bar/qux.txt",
				},
			},
			testCommit{
				Message: "feat: do a thing",
				Tags:    []string{"foo/v0.1.0"},
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
				Tags:    []string{"v0.1.0"},
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

	t.Run("monorepo iteration stop", func(t *testing.T) {
		repo := newMemoryRepo(
			t,
			testCommit{
				Message: "feat!: big breaking change",
				Files: []string{
					"common-file",
				},
			},
			testCommit{
				Message: "feat: foo change",
				Tags:    []string{"foo/v0.1.0"},
				Files: []string{
					"foo",
				},
			},
			testCommit{
				Message: "feat: bar change",
				Tags:    []string{"bar/v0.2.0"},
				Files: []string{
					"bar",
				},
			},
			testCommit{
				Message: "fix: double change",
				Files: []string{
					"bar",
					"foo",
				},
			},
		)

		bot := NewTagbot(TagbotConfig{
			MonorepoConfig: &config.MonoRepoConfig{
				Components: map[string]config.MonoRepoComponent{
					"foo": {
						Name:           "foo",
						ChangeSetGlobs: []string{"foo", "common-file"},
						MaintainLatest: hlp.Ptr(false),
						LatestName:     hlp.Ptr("latest"),
						NoV:            hlp.Ptr(false),
						AlwaysPatch:    hlp.Ptr(false),
					},
					"bar": {
						Name:           "bar",
						ChangeSetGlobs: []string{"foo", "common-file"},
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
		mustHaveTags(t, repo, []string{"foo/v0.1.0", "foo/v0.1.1", "bar/v0.2.0", "bar/v0.2.1"})
	})
}

func TestCommitRelevantToComponent(t *testing.T) {
	t.Parallel()

	testData := []struct {
		name     string
		pattern  string
		file     string
		expected bool
	}{
		{
			name:     "anything",
			pattern:  "**/*",
			file:     "internal/bot/version.go",
			expected: true,
		},
		{
			name:     "top level file",
			pattern:  "**/*",
			file:     "version.go",
			expected: true,
		},
		{
			name:     "subdir - no match",
			pattern:  "foo/*",
			file:     "internal/bot/version.go",
			expected: false,
		},
		{
			name:     "subdir multiglob - no match",
			pattern:  "foo/**/*",
			file:     "internal/foo/version.go",
			expected: false,
		},
		{
			name:     "subdir - yes match",
			pattern:  "internal/**/*",
			file:     "internal/bot/subdir/version.go",
			expected: true,
		},
		{
			name:     "multiglob top level",
			pattern:  "internal/**/*",
			file:     "internal/version.go",
			expected: true,
		},
	}
	for _, tc := range testData {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			bot := &Tagbot{}

			got, err := bot.commitRelevantToComponent(
				&config.MonoRepoComponent{
					ChangeSetGlobs: []string{
						tc.pattern,
					},
				},
				&Commit{
					Files: []string{
						tc.file,
					},
				},
			)
			require.NoError(t, err)
			require.Equal(t, tc.expected, got)
		})
	}
}
