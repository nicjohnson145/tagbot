package main

import (
	"fmt"
	"github.com/samber/lo"
	"github.com/Masterminds/semver"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/object"
	"sort"
	"os"
	"github.com/mitchellh/go-homedir"
	"strings"
	"errors"
)

var errStopIteration = errors.New("stop iteration")

var InitialVersion = lo.Must(semver.NewVersion("v0.0.1"))

const (
	EnvVarRemoteName = "TAGBOT_REMOTE_NAME"
	defaultRemoteName = "origin"

	EnvVarAuthMethod = "TAGBOT_AUTH_METHOD"

	EnvVarToken = "TAGBOT_TOKEN"
	EnvVarKeyPath = "TAGBOT_KEY_PATH"

	sshPrefix = "git@"
	httpsPrefix = "https://"
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

		var hash plumbing.Hash

		obj, err := g.repo.TagObject(tag.Hash())
		switch err {
		case nil:
			hash = obj.Target
		case plumbing.ErrObjectNotFound:
			hash = tag.Hash()
		default:
			return err
		}

		tags = append(tags, Tag{
			Tag: v,
			Hash: hash,
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
	auth, err := g.getAuth()
	if err != nil {
		return err
	}

	err = g.repo.Push(&git.PushOptions{
		RemoteName: g.remoteName(),
		RefSpecs: []config.RefSpec{config.RefSpec("refs/tags/*:refs/tags/*")},
		Auth: auth,
	})
	if err != nil {
		return fmt.Errorf("error pushing tags: %w", err)
	}
	return nil
}

func (g *GitRepo) getAuth() (transport.AuthMethod, error) {
	remoteType, err := g.getRemoteType()
	if err != nil {
		return nil, err
	}

	switch remoteType {
	case RemoteTypeSsh:
		return g.sshAuth()
	case RemoteTypeHttps:
		return g.httpsAuth()
	default:
		return nil, fmt.Errorf("unhandled remote type of %v", remoteType)
	}
}

func (g *GitRepo) getRemoteType() (RemoteType, error) {
	allRemotes, err := g.repo.Remotes()
	if err != nil {
		return RemoteType(""), fmt.Errorf("error listing remotes")
	}

	remoteName := g.remoteName()
	remote, ok := lo.Find(allRemotes, func(r *git.Remote) bool {
		return r.Config().Name == remoteName
	})
	if !ok {
		return RemoteType(""), fmt.Errorf("unable to find remote %v, use %v to specify remote name", remoteName, EnvVarRemoteName)
	}

	if val := os.Getenv(EnvVarAuthMethod); val != "" {
		authMethod, err := ParseAuthMethod(val)
		if err != nil {
			return RemoteType(""), err
		}

		return AuthToRemoteMap[authMethod], nil
	}

	url := remote.Config().URLs[0]
	if strings.HasPrefix(url, sshPrefix) {
		return RemoteTypeSsh, nil
	} else if strings.HasPrefix(url, httpsPrefix) {
		return RemoteTypeHttps, nil
	} else {
		return RemoteType(""), fmt.Errorf(
			"remote %v does not have either %v or %v as a prefix, cannot determine auth method. Use %v to set auth method",
			remoteName,
			sshPrefix,
			httpsPrefix,
			EnvVarAuthMethod,
		)
	}
}

func (g *GitRepo) remoteName() string {
	if val, ok := os.LookupEnv(EnvVarRemoteName); ok {
		return val
	} else {
		return defaultRemoteName
	}
}

func (g *GitRepo) sshAuth() (*ssh.PublicKeys, error) {
	paths := []string{
		"~/.ssh/id_rsa",
		"~/.ssh/id_ecdsa",
	}
	if val := os.Getenv(EnvVarKeyPath); val != "" {
		paths = []string{val}
	}
	for _, p := range paths {
		path, err := homedir.Expand(p)
		if err != nil {
			return nil, fmt.Errorf("error expanding ~: %w", err)
		}
		if _, err := os.Stat(path); err == nil {
			pubKey, err := ssh.NewPublicKeysFromFile("git", path, "")
			if err != nil {
				return nil, fmt.Errorf("error handling public key: %w", err)
			}
			return pubKey, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("error checking existance of %v: %w", path, err)
		}
	}
	return nil, fmt.Errorf("no matching keys found, use %v to specify key path", EnvVarKeyPath)
}

func (g *GitRepo) httpsAuth()  (*http.BasicAuth, error) {
	if val := os.Getenv(EnvVarToken); val != "" {
		return &http.BasicAuth{
			Username: "TagBot",
			Password: val,
		}, nil
	}
	return nil, fmt.Errorf("https auth requires %v to be set", EnvVarToken)
}


func reverseArray[T any](a []T) []T {
	for i, j := 0, len(a) - 1; j > i; i, j = i + 1, j - 1{
		a[i], a[j] = a[j], a[i]
	}
	return a
}
