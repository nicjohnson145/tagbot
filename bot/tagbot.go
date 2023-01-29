package bot

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/nicjohnson145/tagbot/git"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
)


const (
	BreakingChange = "BREAKING CHANGE"
)

var (
	ErrInvalidMessage = errors.New("invalid commit message")
	mergeRegex        = regexp.MustCompile(`^Merge branch`)
	initialVersion    = lo.Must(semver.NewVersion("v0.0.1"))
)

type Config struct {
	Logger      zerolog.Logger
	AlwaysPatch bool
	Latest      bool
	Repo        git.Repo
}

func New(config Config) *Tagbot {
	return &Tagbot{
		log:    config.Logger,
		config: config,
	}
}

type Tagbot struct {
	log    zerolog.Logger
	config Config
}

func (t *Tagbot) Increment() error {
	version, err := t.getNewTag()
	if err != nil {
		t.log.Err(err).Msg("computing new tag")
		return fmt.Errorf("error computing new tag: %w", err)
	}

	if version == nil {
		t.log.Info().Msg("No new tag needed")
		return nil
	}

	t.log.Debug().Str("tag", version.Original()).Msg("starting tag creation")
	if err := t.config.Repo.MakeTagHead(version.Original()); err != nil {
		t.log.Err(err).Msg("making new tag")
		return fmt.Errorf("error making tag: %w", err)
	}
	if err := t.config.Repo.PushTags(); err != nil {
		t.log.Err(err).Msg("pushing new tag")
		return fmt.Errorf("error pushing new tag: %w", err)
	}

	t.log.Info().Msgf("created tag %v", version.Original())

	if t.config.Latest {
		t.log.Debug().Msg("starting 'latest' tag creation")
		if err := t.config.Repo.RemakeTagHead("latest"); err != nil {
			t.log.Err(err).Msg("making 'latest' tag")
			return fmt.Errorf("error making 'latest' tag: %w", err)
		}
		if err := t.config.Repo.ForcePushTags(); err != nil {
			t.log.Err(err).Msg("pushing 'latest' tag")
			return fmt.Errorf("error pushing 'latest' tag: %w", err)
		}

		t.log.Info().Msg("created tag 'latest'")
	}

	return nil
}

func (t *Tagbot) Next() error {
	version, err := t.getNewTag()
	if err != nil {
		t.log.Err(err).Msg("computing new tag")
		return fmt.Errorf("error computing new tag: %w", err)
	}

	if version == nil {
		t.log.Info().Msg("up to date")
		return nil
	}

	t.log.Info().Str("tag", version.Original()).Msg("new tag required")
	return nil
}

func (t *Tagbot) CommitMessage(path string) error {
	disabled, err := t.config.Repo.IsTagbotDisabled()
	if err != nil {
		t.log.Err(err).Msg("checking if tagbot is disabled")
		return fmt.Errorf("error checking if tagbot is disabled: %w", err)
	}

	if disabled {
		t.log.Debug().Msg("skipping validation, tagbot disabled")
		return nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.log.Err(err).Msg("reading commit message file")
		return fmt.Errorf("error reading commit message file: %w", err)
	}

	if !t.isValidCommitMessage(string(content)) {
		return fmt.Errorf("commit message does not conform to tagbot conventions")
	}
	return nil
}

func (t *Tagbot) getNewTag() (*semver.Version, error) {
	latestTag, err := t.config.Repo.LatestTag()
	if err != nil {
		t.log.Err(err).Msg("getting latest tag")
		return nil, fmt.Errorf("error getting latest tag: %w", err)
	}

	if latestTag == nil {
		return initialVersion, nil
	}

	commits, err := t.config.Repo.CommitsSinceHash(latestTag.Hash)
	if err != nil {
		t.log.Err(err).Msg("getting commit list")
		return nil, fmt.Errorf("error getting commit list: %w", err)
	}

	bump, err := t.getVersionBumpForCommits(commits)
	if err != nil {
		t.log.Err(err).Msg("getting version bump for commits")
		return nil, fmt.Errorf("error converting commits to version bump: %w", err)
	}

	if bump == VersionBumpNone && t.config.AlwaysPatch {
		bump = VersionBumpPatch
	}

	newTag := t.newTagForVersionBump(latestTag.Tag, bump)
	return newTag, nil
}

func (t *Tagbot) commitMessageToVersionBump(msg string) (VersionBump, error) {
	prefix, err := t.getCommitType(msg)
	if err != nil {
		return VersionBump(-1), err
	}

	if t.isBreakingChange(msg, prefix) {
		return VersionBumpMajor, nil
	}

	switch prefix {
	case CommitPrefixFeat:
		return VersionBumpMinor, nil
	case CommitPrefixFix:
		return VersionBumpPatch, nil
	default:
		return VersionBumpNone, nil
	}
}

func (t *Tagbot) getVersionBumpForCommits(commits []string) (VersionBump, error) {
	bumpType := VersionBumpNone
	for _, commit := range commits {
		bump, err := t.commitMessageToVersionBump(commit)
		// If we encounter an invalid message, just move past it and keep checking
		if err != nil && !errors.Is(err, ErrInvalidMessage) {
			return VersionBumpNone, err
		}
		if int(bump) > int(bumpType) {
			bumpType = bump
		}
	}

	return bumpType, nil
}

func (t *Tagbot) isValidCommitMessage(msg string) bool {
	_, err := t.getCommitType(msg)
	if err != nil {
		return false
	}
	return true
}

func (t *Tagbot) getCommitType(msg string) (CommitPrefix, error) {
	if mergeRegex.MatchString(msg) {
		return CommitPrefixNop, nil
	}

	for prefix, regex := range prefixRegexes {
		if regex.MatchString(msg) {
			return prefix, nil
		}
	}
	return CommitPrefix(""), ErrInvalidMessage
}

func (t *Tagbot) isBreakingChange(msg string, prefix CommitPrefix) bool {
	if prefix == CommitPrefixNop {
		return false
	}

	if strings.Contains(msg, BreakingChange) {
		return true
	}

	if breakingPrefixes[prefix].MatchString(msg) {
		return true
	}

	return false
}

func (t *Tagbot) newTagForVersionBump(tag *semver.Version, bump VersionBump) *semver.Version {
	var newTag semver.Version
	switch bump {
	case VersionBumpMajor:
		newTag = tag.IncMajor()
	case VersionBumpMinor:
		newTag = tag.IncMinor()
	case VersionBumpPatch:
		newTag = tag.IncPatch()
	default:
		return nil
	}

	return &newTag
}
