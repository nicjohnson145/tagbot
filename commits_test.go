package main

import (
	"fmt"
	"github.com/lithammer/dedent"
	"github.com/stretchr/testify/require"
	"testing"
)

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
			valid := IsValidCommitMessage(tc.Msg)
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
			bump, err := GetVersionBumpForCommits(tc.Commits)
			require.NoError(t, err)
			require.Equal(t, tc.Bump, bump)
		})
	}
}
