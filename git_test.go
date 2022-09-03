package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLatestTag(t *testing.T) {
	repo := newMemoryTestRepo(
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
		repo: repo.Repo,
	}

	tag, err := g.LatestTag()
	require.NoError(t, err, "getting latest tag")
	require.NotNil(t, tag)
	require.Equal(t, "v0.1.1", tag.Tag.Original())
}

func TestCommitsSinceTag(t *testing.T) {
	repo := newMemoryTestRepo(t)
	tagHash := repo.MakeCommit(t, testcommit{
		Msg: "foobar1",
		Tag: "v0.1.0",
	})
	repo.MakeCommit(t, testcommit{
		Msg: "some commit 1",
	})
	repo.MakeCommit(t, testcommit{
		Msg: "some commit 2",
	})

	g := GitRepo{
		repo: repo.Repo,
	}

	commitsSince, err := g.CommitsSinceHash(&tagHash)
	require.NoError(t, err)
	require.Equal(t, []string{"some commit 1", "some commit 2"}, commitsSince)
}

func TestReverseArray(t *testing.T) {
	testData := []struct {
		name   string
		input  []int
		output []int
	}{
		{
			name: "multiple_even",
			input: []int{1, 2, 3, 4},
			output: []int{4, 3, 2, 1},
		},
		{
			name: "multiple_odd",
			input: []int{1, 2, 3},
			output: []int{3, 2, 1},
		},
		{
			name: "one",
			input: []int{1},
			output: []int{1},
		},
		{
			name: "zero",
			input: []int{},
			output: []int{},
		},
	}
	for _, tc := range testData {
		t.Run(tc.name, func(t *testing.T) {
			out := reverseArray(tc.input)
			require.Equal(t, tc.output, out)
		})
	}
}
