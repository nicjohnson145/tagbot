package main

import (
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/apex/log"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"sort"
	"errors"
)

var errStopIteration = errors.New("stop iteration")

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
			log.WithField("name", tag.String()).Warn("error parsing as semver, not considering")
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
		log.Debug("unable to find any valid tags")
		return nil, nil
	}

	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Tag.LessThan(tags[j].Tag)
	})
	return &tags[len(tags)-1], nil
}

func (g *GitRepo) CommitsSinceTag(tag *Tag) ([]string, error) {
	iter, err := g.repo.Log(&git.LogOptions{
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting commit list: %v", err)
	}

	var appendAndContinue func(*object.Commit) bool
	if tag == nil {
		// If we don't get a tag, that just means we consider all commits forever
		appendAndContinue = func(*object.Commit) bool { return true }
	} else {
		// Otherwise we stop once we hit the commit with the same hash as the tag
		appendAndContinue = func(c *object.Commit) bool { return c.Hash != tag.Hash }
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
	reversed := make([]string, 0, len(commits))
	for i := len(commits) - 1; i >= 0; i-- {
		reversed = append(reversed, commits[i])
	}

	return reversed, nil
}
