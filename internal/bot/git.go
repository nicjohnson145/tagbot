package bot

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	gogit "github.com/go-git/go-git/v5"
	gogitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/utils/merkletrie"
	"github.com/go-git/go-git/v5/utils/merkletrie/noder"
	"github.com/nicjohnson145/tagbot/internal/config"
	"github.com/rs/zerolog"
)

type GitRepoConfig struct {
	Path              string
	Remote            string
	AuthMethod        string
	AuthKeyPath       string
	AuthToken         string
	AuthTokenUsername string
}

func NewGitRepo(conf GitRepoConfig) (*GitRepo, error) {
	repo, err := gogit.PlainOpen(conf.Path)
	if err != nil {
		return nil, fmt.Errorf("error opening git repo: %w", err)
	}

	gr := &GitRepo{
		remote: conf.Remote,
		repo:   repo,
	}

	if err := gr.initializeAuth(conf); err != nil {
		return nil, fmt.Errorf("error initializing authorization: %w", err)
	}

	return gr, nil
}

var _ IRepo = (*GitRepo)(nil)

type GitRepo struct {
	remote string

	repo *gogit.Repository
	auth transport.AuthMethod

	tagsByPrefix map[string][]Tag
}

func (g *GitRepo) initializeAuth(conf GitRepoConfig) error {
	var authMethod config.AuthKind

	// if our auth method is explicitly configured, then use that
	if conf.AuthMethod != "" {
		meth, err := config.ParseAuthKind(conf.AuthMethod)
		if err != nil {
			return err
		}
		authMethod = meth
	} else { // otherwise try and detect it from the configured origin of the repo
		meth, err := g.authMethodFromRemote(conf)
		if err != nil {
			return fmt.Errorf("error detecting auth method: %w", err)
		}
		authMethod = meth
	}

	switch authMethod {
	case config.AuthKindPublicKey:
		a, err := g.sshAuth(conf)
		if err != nil {
			return fmt.Errorf("error establishing ssh auth: %w", err)
		}
		g.auth = a
	case config.AuthKindToken:
		a, err := g.tokenAuth(conf)
		if err != nil {
			return fmt.Errorf("error establishing token auth: %w", err)
		}
		g.auth = a
	default:
		return fmt.Errorf("unhandled auth kind %v", authMethod)
	}

	return nil
}

func (g *GitRepo) authMethodFromRemote(conf GitRepoConfig) (config.AuthKind, error) {
	remotes, err := g.repo.Remotes()
	if err != nil {
		return config.AuthKind(""), fmt.Errorf("error listing remotes: %w", err)
	}

	idx := slices.IndexFunc(remotes, func(r *gogit.Remote) bool {
		return r.Config().Name == conf.Remote
	})
	if idx == -1 {
		return config.AuthKind(""), fmt.Errorf("unable to find remote named '%v'", conf.Remote)
	}

	url := remotes[idx].Config().URLs[0]
	if strings.HasPrefix(url, "git@") {
		return config.AuthKindPublicKey, nil
	}
	if strings.HasPrefix(url, "https://") {
		return config.AuthKindToken, nil
	}

	return config.AuthKind(""), fmt.Errorf("unable to auto determine auth method, please explictly configure")
}

func (g *GitRepo) sshAuth(conf GitRepoConfig) (*ssh.PublicKeys, error) {
	if conf.AuthKeyPath == "" {
		return nil, fmt.Errorf("ssh auth detected, but no key path configured")
	}

	key, err := ssh.NewPublicKeysFromFile("git", conf.AuthKeyPath, "")
	if err != nil {
		return nil, fmt.Errorf("error loading public key: %w", err)
	}

	return key, nil
}

func (g *GitRepo) tokenAuth(conf GitRepoConfig) (*http.BasicAuth, error) {
	if conf.AuthToken == "" {
		return nil, fmt.Errorf("token auth detected, but no token configured")
	}

	return &http.BasicAuth{
		Username: conf.AuthTokenUsername,
		Password: conf.AuthToken,
	}, nil
}

func (g *GitRepo) GetLatestTag(ctx context.Context, prefix string) (*Tag, error) {
	if g.tagsByPrefix == nil {
		zerolog.Ctx(ctx).Debug().Msg("tag map is nil, populating tag cache")
		if err := g.constructTagsByPrefixMap(ctx); err != nil {
			return nil, fmt.Errorf("error constructing tag map: %w", err)
		}
	}

	tagList, ok := g.tagsByPrefix[prefix]
	if !ok {
		return nil, nil
	}

	return &tagList[0], nil
}

func (g *GitRepo) constructTagsByPrefixMap(ctx context.Context) error {
	log := zerolog.Ctx(ctx)

	tagIter, err := g.repo.Tags()
	if err != nil {
		return fmt.Errorf("error getting tag iterator: %w", err)
	}

	tagMap := map[string][]Tag{}
	err = tagIter.ForEach(func(tag *plumbing.Reference) error {
		log.Trace().Msgf("processing tag %v", tag.Name().Short())
		prefix := ""
		tagName := tag.Name().Short()

		// Check if its got a prefix, if so separate the two
		if strings.Contains(tagName, "/") {
			log.Trace().Msg("tag contains '/', attempting processing as prefixed tag")
			parts := strings.Split(tagName, "/")
			if len(parts) != 2 {
				log.Debug().Msgf("tag ref %v not in format '<prefix>/<semver>', dropping", tagName)
				return nil
			}

			prefix = parts[0]
			tagName = parts[1]
		}

		// parse the tag as semver
		ver, err := semver.NewVersion(tagName)
		if err != nil {
			log.Debug().Err(err).Msgf("error parsing tag '%v' as semver, dropping", tagName)
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
			log.Err(err).Msg("converting to tag object to hash")
			return err
		}

		log.Trace().Str("prefix", prefix).Str("name", tagName).Str("ref", hash.String()).Msg("constructed tag reference")

		tagList, ok := tagMap[prefix]
		if !ok {
			tagList = []Tag{}
		}

		tagList = append(tagList, Tag{
			TagName: tagName,
			Tag:     ver,
			Hash:    hash.String(),
		})
		tagMap[prefix] = tagList

		return nil
	})
	if err != nil {
		return fmt.Errorf("error iterating tags: %w", err)
	}

	// Once we have our map, sort each list in latest-first order
	for prefix := range tagMap {
		list := tagMap[prefix]
		slices.SortFunc(list, func(a Tag, b Tag) int {
			// Sort in descending order by inverting the comparison
			return -1 * a.Tag.Compare(b.Tag)
		})
		tagMap[prefix] = list
	}

	// finally set our cache
	g.tagsByPrefix = tagMap

	return nil
}

type CommitProcessFunc func(ctx context.Context, commit *Commit) (bool, error)

type Commit struct {
	Hash      string
	ShortHash string
	Message   string
	Files     []string
}

func (g *GitRepo) MakeTagsAtHead(ctx context.Context, tags ...string) error {
	head, err := g.repo.Head()
	if err != nil {
		// huehuehue
		return fmt.Errorf("error getting head: %w", err)
	}

	for _, tag := range tags {
		if err := g.repo.DeleteTag(tag); err != nil && !errors.Is(err, gogit.ErrTagNotFound) {
			return fmt.Errorf("error deleting old tag: %w", err)
		}
		_, err = g.repo.CreateTag(tag, head.Hash(), &gogit.CreateTagOptions{
			Message: "Created By TagBot",
			Tagger: &object.Signature{
				Name:  "TagBot",
				Email: "tagbot@example.com",
				When:  time.Now(),
			},
		})
		if err != nil {
			return fmt.Errorf("error creating tag: %w", err)
		}
	}

	return nil
}

func (g *GitRepo) PushTags(ctx context.Context) error {
	err := g.repo.Push(&gogit.PushOptions{
		RemoteName: g.remote,
		RefSpecs:   []gogitconfig.RefSpec{gogitconfig.RefSpec("refs/tags/*:refs/tags/*")},
		Auth:       g.auth,
		Force:      true, // needed, as we may be overwriting a "latest" tag
	})
	if err != nil {
		return fmt.Errorf("error pushing: %w", err)
	}
	return nil
}

func (g *GitRepo) ProcessLogWhere(ctx context.Context, stopFunc func(commit *object.Commit) bool, processFunc CommitProcessFunc) error {
	iter, err := g.repo.Log(&gogit.LogOptions{
		Order: gogit.LogOrderCommitterTime,
	})
	if err != nil {
		return fmt.Errorf("error constructing iterator: %w", err)
	}

	stopIterationErr := errors.New("__internal_stop_iteration_error")
	err = iter.ForEach(func(commit *object.Commit) error {
		if stopFunc(commit) {
			return stopIterationErr
		}
		files, err := g.getFilesForCommit(commit)
		if err != nil {
			return fmt.Errorf("error getting files: %w", err)
		}

		c := &Commit{
			Hash:      commit.Hash.String(),
			ShortHash: commit.Hash.String()[:10],
			Message:   commit.Message,
			Files:     files,
		}

		shouldContinue, err := processFunc(ctx, c)
		if err != nil {
			return err
		}
		if !shouldContinue {
			return stopIterationErr
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, stopIterationErr) {
			return nil
		}
		return fmt.Errorf("error iterating commits: %w", err)
	}
	return nil
}

func (g *GitRepo) getFilesForCommit(commit *object.Commit) ([]string, error) {
	// "WTF is this?!"
	// So apparently https://github.com/go-git/go-git/issues/307 is like...how this is supposed to work? which is
	// insane to me.
	// https://github.com/metio/terraform-provider-git/commit/fbedb640bc03c4d3ebda8ab75653d18dc1d32277 introduced
	// the idea of computing patches, but the default version tried to also get patch _content_, which for commits
	// that had a large number of files, or large files (cough PF gamp-config), it would balloon memory like crazy.
	// We dont actually need the file content, only the changed file paths and the message. So this is basically the guts
	// of `patch.FilePatches().Files()` but with the content gathering ripped out.
	files := []string{}

	if commit.NumParents() == 0 {
		fileIter, err := commit.Files()
		if err != nil {
			return nil, fmt.Errorf("error creating file iter: %w", err)
		}
		err = fileIter.ForEach(func(f *object.File) error {
			files = append(files, f.Name)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error iterating files: %w", err)
		}
		return files, nil
	}

	parent, err := commit.Parent(0)
	if err != nil {
		return nil, fmt.Errorf("error getting parent: %w", err)
	}

	parentTree, err := parent.Tree()
	if err != nil {
		return nil, fmt.Errorf("error getting parent tree: %w", err)
	}
	commitTree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("error getting commit tree: %w", err)
	}

	parentObj := object.NewTreeRootNode(parentTree)
	commitObj := object.NewTreeRootNode(commitTree)

	hashEqual := func(a, b noder.Hasher) bool {
		return bytes.Equal(a.Hash(), b.Hash())
	}

	changes, err := merkletrie.DiffTree(parentObj, commitObj, hashEqual)
	if err != nil {
		return nil, fmt.Errorf("error computing diff: %w", err)
	}

	for _, c := range changes {
		if c.From != nil {
			files = append(files, c.From.String())
		} else if c.To != nil {
			files = append(files, c.To.String())
		}
	}

	return files, nil
}

func (g *GitRepo) IsTagbotDisabled() (bool, error) {
	conf, err := g.repo.Config()
	if err != nil {
		return false, fmt.Errorf("error fetching git config: %w", err)
	}

	if !conf.Raw.HasSection("tagbot") {
		return false, nil
	}

	return conf.Raw.Section("tagbot").Option("disable") == "true", nil
}
