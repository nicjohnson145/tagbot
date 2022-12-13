package main

import (
	"github.com/Masterminds/semver"
	"github.com/apex/log"
	"os"
	"fmt"
)

type PathSetter interface {
	SetPath(string)
}

type IncrementOpts struct {
	Path        string
	AlwaysPatch bool
}

func (o *IncrementOpts) SetPath(s string) {
	o.Path = s
}

type NextOpts struct {
	Path string
}

func (o *NextOpts) SetPath(s string) {
	o.Path = s
}

type PullRequestOpts struct {
	Path string
}

func (o *PullRequestOpts) SetPath(s string) {
	o.Path = s
}

func IncrementTag(opts IncrementOpts) error {
	repo, err := NewGitRepo(opts.Path)
	if err != nil {
		return err
	}

	version, updated, err := getNewTag(repo, opts.Path, opts.AlwaysPatch)
	if err != nil {
		return err
	}

	if !updated {
		log.Debug("no new tag needed")
		return nil
	}

	log.Debugf("creating new tag: %v", version.Original())
	if err := repo.MakeTagHead(version.Original()); err != nil {
		return err
	}
	if err := repo.PushTags(); err != nil {
		return err
	}

	if os.Getenv("GITHUB_ACTIONS") == "true" {
		fmt.Printf("::set-output name=tag::%v\n", version.Original())
	}

	return nil
}

func Next(opts NextOpts) error {
	repo, err := NewGitRepo(opts.Path)
	if err != nil {
		return err
	}
	newTag, update, err := getNewTag(repo, opts.Path, false)
	if err != nil {
		return err
	}

	if update {
		log.Info(newTag.Original())
	} else {
		log.Info("up to date")
	}
	return nil
}

func getNewTag(repo *GitRepo, path string, alwaysPatch bool) (semver.Version, bool, error) {
	errResp := func(err error) (semver.Version, bool, error) {
		return semver.Version{}, false, err
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

func CommitMessage(path string) error {
	// Per git documentation:
	// Before Git invokes a hook, it changes its working directory to either $GIT_DIR in a bare
	// repository or the root of the working tree in a non-bare repository
	repo, err := NewGitRepo(".")
	if err != nil {
		return err
	}

	disabled, err := repo.IsTagbotDisabled()
	if err != nil {
		return err
	}

	if disabled {
		return nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	if !IsValidCommitMessage(string(content)) {
		return fmt.Errorf("commit message does not conform to tagbot conventions")
	}
	return nil
}

func PullRequest(opts PullRequestOpts) error {
	repo, err := NewGitRepo(opts.Path)
	if err != nil {
		return err
	}

	ref := os.Getenv("GITHUB_BASE_REF")
	if ref == "" {
		return fmt.Errorf("GITHUB_BASE_REF not set, are you running in a pull request context?")
	}

	hash, err := repo.GetHashForBranch(ref)
	if err != nil {
		return err
	}

	commits, err := repo.CommitsSinceHash(hash)
	if err != nil {
		return err
	}

	for _, c := range commits {
		if !IsValidCommitMessage(c) {
			return fmt.Errorf("commit does not comform to tagbot conventions\n\n%v", c)
		}
	}
	return nil
}
