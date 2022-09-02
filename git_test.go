package main

import (
	"testing"
	"github.com/stretchr/testify/require"
)

func TestLatestTag(t *testing.T) {
	repo := makeTestRepo(
		t,
		testcommit{
			Msg: "foobar1",
			Tag: "v0.1.0",
		},
		testcommit{
			Msg: "foobar3",
			Tag: "v0.1.1",
		},
		testcommit{
			Msg: "foobar3",
			Tag: "some-non-semver-tag",
		},
	)

	g := GitRepo{
		repo: repo,
	}

	tag, err := g.LatestTag()
	require.NoError(t, err, "getting latest tag")
	require.NotNil(t, tag)
	require.Equal(t, "v0.1.1", tag.Original())
}
