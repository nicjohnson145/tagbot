package git

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"bytes"

	"github.com/Masterminds/semver"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

var errStopIteration = errors.New("stop iteration")

type Config struct {
	Logger            zerolog.Logger
	Path              string
	Remote            string
	AuthMethod        string
	AuthKeyPath       string
	AuthToken         string
	AuthTokenUsername string
}

func NewGitRepo(config Config) (*GitRepo, error) {
	repo, err := git.PlainOpen(config.Path)
	if err != nil {
		config.Logger.Err(err).Str("path", config.Path).Msg("error opening git repo")
		return nil, fmt.Errorf("error opening git repo: %w", err)
	}

	return &GitRepo{
		repo:   repo,
		log:    config.Logger,
		config: config,
	}, nil
}

var _ Repo = (*GitRepo)(nil)

type GitRepo struct {
	repo   *git.Repository
	log    zerolog.Logger
	config Config
}

func (g *GitRepo) IsTagbotDisabled() (bool, error) {
	conf, err := g.repo.Config()
	if err != nil {
		g.log.Err(err).Msg("fetching git config")
		return false, fmt.Errorf("error fetching git config: %w", err)
	}

	if !conf.Raw.HasSection("tagbot") {
		return false, nil
	}

	return conf.Raw.Section("tagbot").Option("disable") == "true", nil
}

func (g *GitRepo) LatestTag() (*Tag, error) {
	iter, err := g.repo.Tags()
	if err != nil {
		g.log.Err(err).Msg("getting repo tags")
		return nil, fmt.Errorf("error getting repo tags: %w", err)
	}

	tags := []Tag{}
	err = iter.ForEach(func(tag *plumbing.Reference) error {
		v, err := semver.NewVersion(tag.Name().Short())
		if err != nil {
			g.log.Debug().Str("error", err.Error()).Str("tag", tag.Name().Short()).Msg("ignoring tag parsing error")
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
			g.log.Err(err).Msg("converting to tag object")
			return err
		}

		tags = append(tags, Tag{
			Tag:  v,
			Hash: hash.String(),
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error iterating tags: %w", err)
	}

	if len(tags) == 0 {
		g.log.Debug().Msg("found zero tags")
		return nil, nil
	}

	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Tag.LessThan(tags[j].Tag)
	})
	return &tags[len(tags)-1], nil
}

func (g *GitRepo) CommitsSinceHash(hashStr string) ([]string, error) {
	hash := plumbing.NewHash(hashStr)

	iter, err := g.repo.Log(&git.LogOptions{
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		g.log.Err(err).Msg("getting commit list")
		return nil, fmt.Errorf("error getting commit list: %v", err)
	}

	commits := []string{}
	err = iter.ForEach(func(c *object.Commit) error {
		if bytes.Compare(c.Hash[:], hash[:]) != 0 {
			commits = append(commits, c.Message)
			return nil
		}
		return errStopIteration
	})
	if err != nil && !errors.Is(err, errStopIteration) {
		g.log.Err(err).Msg("iterating commits")
		return nil, fmt.Errorf("error iterating commits: %w", err)
	}

	// Reverse it, so we can iterate in the order things were committed
	reverseArray(commits)
	return commits, nil
}

func (g *GitRepo) MakeTagHead(name string) error {
	return g.makeTagAtHead(name, false)
}

func (g *GitRepo) RemakeTagHead(name string) error {
	return g.makeTagAtHead(name, true)
}

func (g *GitRepo) makeTagAtHead(name string, deleteFirst bool) error {
	head, err := g.repo.Head()
	if err != nil {
		g.log.Err(err).Msg("getting repo head")
		return fmt.Errorf("error getting repo head: %w", err)
	}
	if deleteFirst {
		if err := g.repo.DeleteTag(name); err != nil && !errors.Is(err, git.ErrTagNotFound) {
			g.log.Err(err).Msg("deleting old tag")
			return fmt.Errorf("error deleting old tag: %w", err)
		}
	}
	_, err = g.repo.CreateTag(name, head.Hash(), &git.CreateTagOptions{
		Message: "created by TagBot",
		Tagger: &object.Signature{
			Name:  "TagBot",
			Email: "tagbot@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		g.log.Err(err).Msg("creating tag")
		return fmt.Errorf("error creating tag: %w", err)
	}
	return nil
}

func (g *GitRepo) PushTags() error {
	return g.pushTags(false)
}

func (g *GitRepo) ForcePushTags() error {
	return g.pushTags(true)
}

func (g *GitRepo) pushTags(force bool) error {
	auth, err := g.getAuth()
	if err != nil {
		g.log.Err(err).Msg("getting authentication")
		return fmt.Errorf("error getting authentication: %w", err)
	}

	g.log.Debug().Str("remote", g.config.Remote).Bool("force", force).Msg("pushing tag")
	err = g.repo.Push(&git.PushOptions{
		RemoteName: g.config.Remote,
		RefSpecs:   []config.RefSpec{config.RefSpec("refs/tags/*:refs/tags/*")},
		Auth:       auth,
		Force:      force,
	})
	if err != nil {
		g.log.Err(err).Msg("pushing tags")
		return fmt.Errorf("error pushing tags: %w", err)
	}
	return nil
}

func (g *GitRepo) getAuth() (transport.AuthMethod, error) {
	if g.config.AuthMethod != "" {
		method, err := ParseAuthMethod(g.config.AuthMethod)
		if err != nil {
			g.log.Err(err).Msg("parsing auth method")
			return nil, fmt.Errorf("error parsing auth method: %w", err)
		}
		switch method {
		case AuthMethodPublicKey:
			return g.sshAuth()
		case AuthMethodToken:
			return g.httpsAuth()
		default:
			return nil, fmt.Errorf("unhandled auth method %v", g.config.AuthMethod)
		}
	}

	// Otherwise try to auto detect it
	remoteType, err := g.getRemoteType()
	if err != nil {
		g.log.Err(err).Msg("getting remote type")
		return nil, fmt.Errorf("error getting remote type: %w", err)
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
		g.log.Err(err).Msg("listing remotes")
		return RemoteType(""), fmt.Errorf("error listing remotes")
	}

	remote, ok := lo.Find(allRemotes, func(r *git.Remote) bool {
		return r.Config().Name == g.config.Remote
	})
	if !ok {
		g.log.Error().Str("remote", g.config.Remote).Msg("unable to find")
		return RemoteType(""), fmt.Errorf("unable to find remote %v", g.config.Remote)
	}

	url := remote.Config().URLs[0]
	if strings.HasPrefix(url, sshPrefix) {
		return RemoteTypeSsh, nil
	} else if strings.HasPrefix(url, httpsPrefix) {
		return RemoteTypeHttps, nil
	} else {
		g.log.Error().Str("remote", g.config.Remote).Msg("cannot auto determine auth method")
		return RemoteType(""), fmt.Errorf(
			"remote %v does not have either %v or %v as a prefix, cannot determine auth method",
			g.config.Remote,
			sshPrefix,
			httpsPrefix,
		)
	}
}

func (g *GitRepo) sshAuth() (*ssh.PublicKeys, error) {
	pubKey, err := ssh.NewPublicKeysFromFile("git", g.config.AuthKeyPath, "")
	if err != nil {
		g.log.Err(err).Str("path", g.config.AuthKeyPath).Msg("reading key")
		return nil, fmt.Errorf("error reading key: %w", err)
	}
	return pubKey, nil
}

func (g *GitRepo) httpsAuth() (*http.BasicAuth, error) {
	return &http.BasicAuth{Username: g.config.AuthTokenUsername, Password: g.config.AuthToken}, nil
}

func (g *GitRepo) GetHashForBranch(branch string) (string, error) {
	h, err := g.repo.ResolveRevision(plumbing.Revision(branch))
	if err != nil {
		g.log.Err(err).Str("branch", branch).Msg("resolving branch")
		return "", fmt.Errorf("error resolving branch %v: %w", branch, err)
	}
	return h.String(), nil
}

func reverseArray[T any](a []T) []T {
	for i, j := 0, len(a)-1; j > i; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
	return a
}
