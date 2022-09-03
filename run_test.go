package main

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

func TestGetNewTag(t *testing.T) {
	testData := []struct {
		name            string
		commits         []testcommit
		expectedVersion semver.Version
		expectedUpdate  bool

		optForcePatch   bool
	}{
		{
			name: "normal_repo",
			commits: []testcommit{
				{Msg: "feat: blarg", Tag: "v0.1.0"},
				{Msg: "feat: another blarg"},
			},
			expectedVersion: *lo.Must(semver.NewVersion("v0.2.0")),
			expectedUpdate: true,
		},
		{
			name:    "no_recent_commits",
			commits: []testcommit{
				{Msg: "feat: blarg", Tag: "v0.1.0"},
				{Msg: "feat: another blarg", Tag: "v0.2.0"},
			},
			expectedVersion: semver.Version{},
			expectedUpdate: false,
		},
		{
			name:    "no_recent_commits_patch_forced",
			commits: []testcommit{
				{Msg: "feat: blarg", Tag: "v0.1.0"},
				{Msg: "feat: another blarg", Tag: "v0.2.0"},
			},
			expectedVersion: *lo.Must(semver.NewVersion("v0.2.1")),
			expectedUpdate: true,
			optForcePatch: true,
		},
		{
			name: "initial_untagged_repo",
			commits: []testcommit{
				{Msg: "Initial commit"},
				{Msg: "feat: another blarg"},
			},
			expectedVersion: *lo.Must(semver.NewVersion("v0.0.1")),
			expectedUpdate: true,
		},
	}
	for _, tc := range testData {
		t.Run(tc.name, func(t *testing.T) {
			repo := newDiskTestRepo(t, tc.commits...)
			gitRepo, err := NewGitRepo(repo.Dir)
			require.NoError(t, err)
			version, update, err := getNewTag(gitRepo, repo.Dir, tc.optForcePatch)
			require.NoError(t, err)
			require.Equal(t, tc.expectedUpdate, update)
			require.Equal(t, tc.expectedVersion, version)
		})
	}
}
