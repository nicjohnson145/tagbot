package main

import (
	"github.com/apex/log"
	"github.com/Masterminds/semver"
)

type IncrementOpts struct {
	Path        string
	AlwaysPatch bool
}

func IncrementTag(opts IncrementOpts) error {
	version, updated, err := getNewTag(opts.Path, opts.AlwaysPatch)
	if err != nil {
		return err
	}

	if !updated {
		log.Info("no new tag needed")
		return nil
	}

	log.Infof("new tag: %v", version.Original())
	return nil
}

func getNewTag(path string, alwaysPatch bool) (semver.Version, bool, error) {
	errResp := func(err error) (semver.Version, bool, error) {
		return semver.Version{}, false, err
	}

	repo, err := NewGitRepo(path)
	if err != nil {
		return errResp(err)
	}

	latestTag, err := repo.LatestTag()
	if err != nil {
		return errResp(err)
	}

	if latestTag == nil {
		return *InitialVersion, true, nil
	}

	commits, err := repo.CommitsSinceHash(&latestTag.Hash)
	if err != nil {
		return errResp(err)
	}

	bump, err := GetVersionBumpForCommits(commits)
	if err != nil {
		return errResp(err)
	}

	if bump == VersionBumpNone && alwaysPatch {
		bump = VersionBumpPatch
	}

	if bump == VersionBumpNone {
		return semver.Version{}, false, nil
	}

	newTag := NewTagForVersionBump(latestTag.Tag, bump)
	return newTag, true, nil
}
