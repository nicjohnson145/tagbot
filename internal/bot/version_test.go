package bot

import (
	"context"
	"testing"

	"github.com/go-jose/go-jose/v4/testutils/require"
	"github.com/lithammer/dedent"
)

func TestVersionBumpFromCommitMessage(t *testing.T) {
	t.Parallel()

	testData := []struct {
		name     string
		message  string
		expected VersionBump
	}{
		{
			name: "basic feature message",
			message: dedent.Dedent(`
				feat: do a thing

				with a sub message with an ! in it
			`[1:]),
			expected: VersionBumpMinor,
		},
		{
			name: "basic fix message",
			message: dedent.Dedent(`
				fix: fix a thing
			`[1:]),
			expected: VersionBumpPatch,
		},
		{
			name: "breaking feature",
			message: dedent.Dedent(`
				feat!: break a thing
			`[1:]),
			expected: VersionBumpMajor,
		},
		{
			name: "breaking feature in description",
			message: dedent.Dedent(`
				feat: break a thing

				BREAKING CHANGE: this is a big one
			`[1:]),
			expected: VersionBumpMajor,
		},
		{
			name: "non confirming",
			message: dedent.Dedent(`
				random commit message

				changes go here
			`[1:]),
			expected: VersionBumpNone,
		},
		{
			name: "non standard",
			message: dedent.Dedent(`
				random!: commit message

				changes go here
			`[1:]),
			expected: VersionBumpNone,
		},
		{
			name: "non standard breaking description",
			message: dedent.Dedent(`
				random commit message

				BREAKING CHANGE changes go here
			`[1:]),
			expected: VersionBumpMajor,
		},
		{
			name: "sub type",
			message: dedent.Dedent(`
				fix(ci): correct go version
			`[1:]),
			expected: VersionBumpPatch,
		},
		{
			name: "sub type breaking",
			message: dedent.Dedent(`
				feat(api)!: breaking api feature
			`[1:]),
			expected: VersionBumpMajor,
		},
	}
	for _, tc := range testData {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := VersionBumpFromCommitMessage(context.Background(), tc.message)
			require.Equal(t, got, tc.expected)
		})
	}
}
