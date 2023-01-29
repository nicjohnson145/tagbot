package bot

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/nicjohnson145/tagbot/git"
	gitMock "github.com/nicjohnson145/tagbot/mocks/git"
	"github.com/stretchr/testify/require"
)

func TestIncrement(t *testing.T) {
	const hashStr = "3f90193ad5ec9cdd6ee2af363d20b4205405d4bf"
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
