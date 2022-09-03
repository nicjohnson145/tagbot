package main

import (
	"fmt"
	"github.com/samber/lo"
	"github.com/Masterminds/semver"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/object"
	"sort"
	"errors"
)

var errStopIteration = errors.New("stop iteration")

var InitialVersion = lo.Must(semver.NewVersion("v0.0.1"))


func NewGitRepo(path string) (*GitRepo, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("error opening git repo: %w", err)
	}

	return &GitRepo{
		repo: repo,
	}, nil
}

type GitRepo struct {
	repo *git.Repository
}

type Tag struct {
	Hash plumbing.Hash
	Tag *semver.Version
}

func (g *GitRepo) LatestTag() (*Tag, error) {
	iter, err := g.repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("error getting repo tags: %w", err)
	}

	tags := []Tag{}
	err = iter.ForEach(func(tag *plumbing.Reference) error {
		v, err := semver.NewVersion(tag.Name().Short())
		if err != nil {
			return nil
		}

		tags = append(tags, Tag{
			Tag: v,
			Hash: tag.Hash(),
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error iterating tags: %w", err)
	}

	if len(tags) == 0 {
		return nil, nil
	}

	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Tag.LessThan(tags[j].Tag)
	})
	return &tags[len(tags)-1], nil
}

func (g *GitRepo) CommitsSinceHash(hash *plumbing.Hash) ([]string, error) {
	iter, err := g.repo.Log(&git.LogOptions{
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting commit list: %v", err)
	}

	var appendAndContinue func(*object.Commit) bool
	if hash == nil {
		// If we don't get a tag, that just means we consider all commits forever
		appendAndContinue = func(*object.Commit) bool { return true }
	} else {
		// Otherwise we stop once we hit the commit with the same hash as the tag
		appendAndContinue = func(c *object.Commit) bool { return c.Hash != *hash }
	}

	commits := []string{}
	err = iter.ForEach(func(c *object.Commit) error {
		if appendAndContinue(c) {
			commits = append(commits, c.Message)
			return nil
		}
		return errStopIteration
	})
	if err != nil && !errors.Is(err, errStopIteration) {
		return nil, fmt.Errorf("error iterating commits: %w", err)
	}

	// Reverse it, so we can iterate in the order things were committed
	reverseArray(commits)
	return commits, nil
}

func (g *GitRepo) MakeTagHead(name string) error {
	h, err := g.repo.Head()
	if err != nil {
		return fmt.Errorf("error getting repo head: %w", err)
	}

	return g.MakeTag(name, h.Hash())
}

func (g *GitRepo) MakeTag(name string, hash plumbing.Hash) error {
	_, err := g.repo.CreateTag(name, hash, &git.CreateTagOptions{Message: "created by TagBot"})
	if err != nil {
		return fmt.Errorf("error creating tag: %w", err)
	}
	return nil
}

func (g *GitRepo) PushTags() error {
	err := g.repo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs: []config.RefSpec{config.RefSpec("refs/tags/*:refs/tags/*")},
		Auth: &http.BasicAuth{
			Username: "TagBot",
			Password: "",
		},
	})
	if err != nil {
		return fmt.Errorf("error pushing tags: %w", err)
	}
	return nil
}

func reverseArray[T any](a []T) []T {
	for i, j := 0, len(a) - 1; j > i; i, j = i + 1, j - 1{
		a[i], a[j] = a[j], a[i]
	}
	return a
}
