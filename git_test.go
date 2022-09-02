package main

import (
	"github.com/stretchr/testify/require"
	"testing"
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
	require.Equal(t, "v0.1.1", tag.Tag.Original())
}

func TestCommitsSinceTag(t *testing.T) {
	repo := makeTestRepo(
		t,
		testcommit{
			Msg: "foobar1",
			Tag: "v0.1.0",
		},
		testcommit{
			Msg: "some commit 1",
		},
		testcommit{
			Msg: "some commit 2",
		},
	)

	g := GitRepo{
		repo: repo,
	}

	tag, err := g.LatestTag()
	require.NoError(t, err)
	require.NotNil(t, tag)
	require.Equal(t, "v0.1.0", tag.Tag.Original())

	commitsSince, err := g.CommitsSinceTag(tag)
	require.NoError(t, err)
	require.Equal(t, []string{"some commit 1", "some commit 2"}, commitsSince)
}
