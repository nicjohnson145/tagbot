package config

import (
	"os"
	"testing"

	"github.com/lithammer/dedent"
	"github.com/nicjohnson145/hlp"
	"github.com/stretchr/testify/require"
)

func TestParseMonoRepoConfig(t *testing.T) {
	t.Run("smokes test", func(t *testing.T) {
		dir := t.TempDir()
		content := dedent.Dedent(`
			components:
			  foo:
			    change-set-globs:
			    - 'foo/*'
			  bar:
			    change-set-globs:
			    - 'bar/*'
		`[1:])
		require.NoError(t, os.WriteFile(dir + "/file.yaml", []byte(content), 0644))

		got, err := ParseMonoRepoConfig(dir + "/file.yaml")
		require.NoError(t, err)
		require.Equal(
			t,
			&MonoRepoConfig{
				Components: map[string]MonoRepoComponent{
					"foo": {
						Name: "foo",
						ChangeSetGlobs: []string{
							"foo/*",
						},
						MaintainLatest: hlp.Ptr(false),
						LatestName: hlp.Ptr(""),
						NoV: hlp.Ptr(false),
						AlwaysPatch: hlp.Ptr(false),
					},
					"bar": {
						Name: "bar",
						ChangeSetGlobs: []string{
							"bar/*",
						},
						MaintainLatest: hlp.Ptr(false),
						LatestName: hlp.Ptr(""),
						NoV: hlp.Ptr(false),
						AlwaysPatch: hlp.Ptr(false),
					},
				},
			},
			got,
		)
	})
}
