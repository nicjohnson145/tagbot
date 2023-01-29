package bot

import (
	"fmt"
	"testing"

	"github.com/Masterminds/semver"
	"github.com/lithammer/dedent"
	"github.com/nicjohnson145/tagbot/git"
	gitMock "github.com/nicjohnson145/tagbot/mocks/git"
	"github.com/stretchr/testify/require"
)

const hashStr = "3f90193ad5ec9cdd6ee2af363d20b4205405d4bf"

func dedentMsg(s string) string {
	return dedent.Dedent(s)[1:]
}

func TestIsValidCommitMessage(t *testing.T) {
	testData := []struct {
		Msg   string
		Valid bool
	}{
		{
			Msg:   "feat: allow provided config object to extend other configs",
			Valid: true,
		},
		{
			Msg:   "FEAT: allow provided config object to extend other configs",
			Valid: true,
		},
		{
			Msg:   "feat(lang): add polish language",
			Valid: true,
		},
		{
			Msg: dedentMsg(`
				chore!: drop Node 6 from testing matrix

				BREAKING CHANGE: dropping Node 6 which hits end of life in April
			`),
			Valid: true,
		},
		{
			Msg:   "fixed a couple bugs with the thing",
			Valid: false,
		},
	}
	for idx, tc := range testData {
		t.Run(fmt.Sprint(idx), func(t *testing.T) {
			tagbot := New(Config{})
			valid := tagbot.isValidCommitMessage(tc.Msg)
			require.Equal(t, tc.Valid, valid, tc.Msg)
		})
	}
}

func TestGetVersionBumpForCommits(t *testing.T) {
	testData := []struct {
		Commits []string
		Bump    VersionBump
	}{
		{
			Commits: []string{
				"fix: fix a thing",
				"feat: do a thing",
				dedentMsg(`
					fix: fix bug

					BREAKING CHANGE stuff works different
				`),
			},
			Bump: VersionBumpMajor,
		},
		{
			Commits: []string{
				"fix: fix a thing",
				"feat: do a thing",
			},
			Bump: VersionBumpMinor,
		},
		{
			Commits: []string{
				"fix: fix a thing",
				"docs: document a thing",
			},
			Bump: VersionBumpPatch,
		},
		{
			Commits: []string{
				"docs: document a thing",
				"chore: chore-thing",
			},
			Bump: VersionBumpNone,
		},
		{
			Commits: []string{
				"feat: do a thing",
				"refactor!: make a breaking refactor",
			},
			Bump: VersionBumpMajor,
		},
		{
			Commits: []string{
				"feat: do a thing",
				"refactor(config)!: make a breaking refactor",
			},
			Bump: VersionBumpMajor,
		},
		{
			Commits: []string{},
			Bump: VersionBumpNone,
		},
	}
	for idx, tc := range testData {
		t.Run(fmt.Sprint(idx), func(t *testing.T) {
			tagbot := New(Config{})
			bump, err := tagbot.getVersionBumpForCommits(tc.Commits)
			require.NoError(t, err)
			require.Equal(t, tc.Bump, bump)
		})
	}
}

func TestIncrement(t *testing.T) {
	var commitMessages = []string{
		"feat: did a thing",
		"fix: fixed a thing",
	}

	t.Run("no latest", func(t *testing.T) {
		repo := gitMock.NewRepo(t)
		repo.EXPECT().LatestTag().Return(&git.Tag{Tag: semver.MustParse("v1.0.0"), Hash: hashStr}, nil)
		repo.EXPECT().CommitsSinceHash(hashStr).Return(commitMessages, nil)
		repo.EXPECT().MakeTagHead("v1.1.0").Return(nil)
		repo.EXPECT().PushTags().Return(nil)

		tagbot := New(Config{
			Repo: repo,
		})
		require.NoError(t, tagbot.Increment())
	})

	t.Run("latest", func(t *testing.T) {
		repo := gitMock.NewRepo(t)

		repo.EXPECT().LatestTag().Return(&git.Tag{Tag: semver.MustParse("v1.0.0"), Hash: hashStr}, nil)
		repo.EXPECT().CommitsSinceHash(hashStr).Return(commitMessages, nil)
		repo.EXPECT().MakeTagHead("v1.1.0").Return(nil)
		repo.EXPECT().PushTags().Return(nil)
		repo.EXPECT().RemakeTagHead("latest").Return(nil)
		repo.EXPECT().ForcePushTags().Return(nil)

		tagbot := New(Config{
			Repo: repo,
			Latest: true,
		})
		require.NoError(t, tagbot.Increment())
	})

	t.Run("always patch", func(t *testing.T) {
		noBumpCommits := []string{
			"docs: update the docs",
			"ci: fix the CI",
		}
		repo := gitMock.NewRepo(t)

		repo.EXPECT().LatestTag().Return(&git.Tag{Tag: semver.MustParse("v1.0.0"), Hash: hashStr}, nil)
		repo.EXPECT().CommitsSinceHash(hashStr).Return(noBumpCommits, nil)
		repo.EXPECT().MakeTagHead("v1.0.1").Return(nil)
		repo.EXPECT().PushTags().Return(nil)

		tagbot := New(Config{
			Repo: repo,
			AlwaysPatch: true,
		})
		require.NoError(t, tagbot.Increment())
	})
}

func TestNext(t *testing.T) {
	// It's an info only command, so just make sure it doesnt explode. It uses most of the same code
	// as increment so it should be pretty well tested
	t.Run("smokes", func(t *testing.T) {
		repo := gitMock.NewRepo(t)
		repo.EXPECT().LatestTag().Return(&git.Tag{Tag: semver.MustParse("v1.0.0"), Hash: hashStr}, nil)
		repo.EXPECT().CommitsSinceHash(hashStr).Return([]string{"feat: do a thing"}, nil)

		tagbot := New(Config{
			Repo: repo,
		})
		require.NoError(t, tagbot.Next())
	})
}
