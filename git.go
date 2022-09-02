package main

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/Masterminds/semver"
	"fmt"
	"github.com/apex/log"
	"sort"
)

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

func (g *GitRepo) LatestTag() (*semver.Version, error) {
	iter, err := g.repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("error getting repo tags: %w", err)
	}

	tags := []*semver.Version{}
	err = iter.ForEach(func(tag *plumbing.Reference) error {
		v, err := semver.NewVersion(tag.Name().Short())
		if err != nil {
			log.WithField("name", tag.String()).Warn("error parsing as semver, not considering")
			return nil
		}

		tags = append(tags, v)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error iterating tags: %w", err)
	}

	if len(tags) == 0 {
		log.Debug("unable to find any valid tags")
		return nil, nil
	}

	sort.Sort(semver.Collection(tags))
	return tags[len(tags) - 1], nil
}
