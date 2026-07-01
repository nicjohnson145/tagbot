package bot

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/Masterminds/semver"
	"github.com/nicjohnson145/hlp"
	"github.com/nicjohnson145/tagbot/internal/config"
	"github.com/rs/zerolog"
)

type TagbotConfig struct {
	MonorepoConfig *config.MonoRepoConfig
	Repo           IRepo
}

func NewTagbot(conf TagbotConfig) *Tagbot {
	return &Tagbot{
		monorepoConfig: conf.MonorepoConfig,
		repo:           conf.Repo,
	}
}

type Tagbot struct {
	monorepoConfig *config.MonoRepoConfig
	repo           IRepo
}

func (t *Tagbot) Run(ctx context.Context) error {
	log := zerolog.Ctx(ctx)

	type update struct {
		Component  *config.MonoRepoComponent
		NewVersion *semver.Version
	}

	getPrefix := func(component *config.MonoRepoComponent) string {
		prefix := component.Name
		if component.Prefix != nil {
			prefix = *component.Prefix
		}
		return prefix
	}

	// Process components in alphabetical order to avoid flaky tests
	keys := hlp.Keys(t.monorepoConfig.Components)
	sort.Strings(keys)

	updates := []update{}
	for _, key := range keys {
		component := t.monorepoConfig.Components[key]
		log.Info().Msgf("processing %v", component.Name)

		prefix := getPrefix(&component)
		log.Debug().Msgf("using prefix '%v'", prefix)

		mostRecent, err := t.repo.GetLatestTag(ctx, prefix)
		if err != nil {
			return fmt.Errorf("error getting most recent tag: %w", err)
		}

		if mostRecent != nil {
			log.Debug().Msgf("most recent tag is %v:%v", mostRecent.Tag, mostRecent.Hash)
		} else {
			log.Debug().Msg("no tag found")
		}
		bump := VersionBumpIrrelevant
		if mostRecent != nil { // standard loop if we find an existing tag
			err = t.repo.ProcessCommitsUntil(ctx, mostRecent.Hash, func(ctx context.Context, commit *Commit) (bool, error) {
				commitBump, err := t.processCommit(ctx, &component, commit)
				if err != nil {
					return false, err
				}
				if commitBump.Greater(bump) {
					bump = commitBump
				}

				return true, nil
			})
			if err != nil {
				return fmt.Errorf("error iterating commits: %w", err)
			}
			if bump != VersionBumpIrrelevant && VersionBumpPatch.Greater(bump) && component.AlwaysPatch != nil && *component.AlwaysPatch {
				bump = VersionBumpPatch
			}
		} else {
			// otherwise there's a chance we're making our first tag, which in a monorepo scenario we should only do if the
			// commits are relevant to the change set
			err = t.repo.ProcessLatestCommit(ctx, func(ctx context.Context, commit *Commit) (bool, error) {
				relevant, err := t.commitRelevantToComponent(&component, commit)
				if err != nil {
					return false, err
				}
				if relevant {
					bump = VersionBumpMinor
				}
				return true, nil
			})
			if err != nil {
				return fmt.Errorf("error processing latest commit: %w", err)
			}
		}

		log.Info().Msgf("decided action is %v", bump)
		if bump.Greater(VersionBumpNone) {
			var newTag semver.Version
			if mostRecent == nil {
				log.Debug().Msg("no previous tag, will create initial")
				newTag = *InitialTag
			} else {
				switch bump {
				case VersionBumpMajor:
					newTag = mostRecent.Tag.IncMajor()
				case VersionBumpMinor:
					newTag = mostRecent.Tag.IncMinor()
				case VersionBumpPatch:
					newTag = mostRecent.Tag.IncPatch()
				default:
					return fmt.Errorf("unhandled version bump %v", bump)
				}
			}

			updates = append(updates, update{
				Component:  &component,
				NewVersion: &newTag,
			})
		}
	}

	if len(updates) == 0 {
		log.Info().Msg("no components require tag updates")
		return nil
	}

	makeTagString := func(component *config.MonoRepoComponent, version string) string {
		prefix := ""
		v := "v"
		if compPrefix := getPrefix(component); compPrefix != "" {
			prefix = compPrefix + "/"
		}
		if *component.NoV {
			v = ""
		}
		return prefix + v + version
	}
	makeLatest := func(component *config.MonoRepoComponent) string {
		prefix := ""
		if compPrefix := getPrefix(component); compPrefix != "" {
			prefix = compPrefix + "/"
		}
		return prefix + *component.LatestName
	}

	for _, up := range updates {
		wantTags := []string{
			makeTagString(up.Component, up.NewVersion.String()),
		}
		if up.Component.MaintainLatest != nil && *up.Component.MaintainLatest {
			wantTags = append(wantTags, makeLatest(up.Component))
		}

		log.Info().Msgf("creating %v", wantTags)
		if err := t.repo.MakeTagsAtHead(ctx, wantTags...); err != nil {
			return fmt.Errorf("error creating tag: %w", err)
		}
	}

	log.Info().Msgf("pushing tags")
	if err := t.repo.PushTags(ctx); err != nil {
		return fmt.Errorf("error pushing tags: %w", err)
	}

	return nil
}

func (t *Tagbot) processCommit(ctx context.Context, component *config.MonoRepoComponent, commit *Commit) (VersionBump, error) {
	log := zerolog.Ctx(ctx)

	log.Trace().Msgf("processing commit %v", commit.Hash)
	relevant, err := t.commitRelevantToComponent(component, commit)
	if err != nil {
		return VersionBump(-1), err
	}

	if !relevant {
		log.Trace().Msg("commit determined not relevant to this component")
		return VersionBumpIrrelevant, nil
	}

	return VersionBumpFromCommitMessage(ctx, commit.Message), nil
}

func (t *Tagbot) commitRelevantToComponent(component *config.MonoRepoComponent, commit *Commit) (bool, error) {
	for _, file := range commit.Files {
		for _, glob := range component.ChangeSetGlobs {
			// For some reason, a bare splat doesnt work like I would expect, so special case it
			if glob == "*" {
				return true, nil
			}

			match, err := filepath.Match(glob, file)
			if err != nil {
				return false, fmt.Errorf("error checking glob match: %w", err)
			}
			if match {
				return true, nil
			}
		}
	}
	return false, nil
}

func (t *Tagbot) CommitMessage(ctx context.Context, content string) error {
	disabled, err := t.repo.IsTagbotDisabled()
	if err != nil {
		return fmt.Errorf("error checking if tagbot is disabled: %w", err)
	}

	if disabled {
		log := zerolog.Ctx(ctx)
		log.Debug().Msg("skipping validation, tagbot disabled")
		return nil
	}

	if _, err := EnsureValidCommitMessage(ctx, content); err != nil {
		return err
	}

	return nil
}
