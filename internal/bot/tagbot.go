package bot

import (
	"context"
	"fmt"
	"sort"

	"github.com/Masterminds/semver"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/nicjohnson145/hlp"
	"github.com/nicjohnson145/hlp/set"
	"github.com/nicjohnson145/tagbot/internal/config"
	"github.com/rs/zerolog"
)

type TagbotConfig struct {
	MonorepoConfig *config.MonoRepoConfig
	Repo           IRepo
	DryRun         bool
}

func NewTagbot(conf TagbotConfig) *Tagbot {
	return &Tagbot{
		monorepoConfig: conf.MonorepoConfig,
		repo:           conf.Repo,
		dryRun:         conf.DryRun,
	}
}

type Tagbot struct {
	monorepoConfig *config.MonoRepoConfig
	repo           IRepo
	dryRun         bool
}

func (t *Tagbot) Run(ctx context.Context) error {
	log := zerolog.Ctx(ctx)

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

	// get the latest tag for each component, so we only have to walk the commit tree once
	log.Info().Msg("getting latest tags by prefix")
	latestTags := map[string]*Tag{}
	for _, key := range keys {
		component := t.monorepoConfig.Components[key]
		log.Info().Msgf("getting latest tag for %v", component.Name)
		prefix := getPrefix(&component)
		log.Debug().Msgf("using prefix '%v'", prefix)
		mostRecent, err := t.repo.GetLatestTag(ctx, prefix)
		if err != nil {
			return fmt.Errorf("error getting most recent tag: %w", err)
		}
		if mostRecent != nil {
			log.Debug().Msgf("most recent tag is %v:%v", mostRecent.TagName, mostRecent.Hash)
		} else {
			log.Debug().Msg("no tag found")
		}
		latestTags[key] = mostRecent
	}

	// now walk the commits and figure out what our bumps should be, stopping once we've processed every prefix
	log.Info().Msg("walking commit history")
	bumpMap := hlp.MapFromSlice(keys, func(key string, _ int) (string, VersionBump) {
		return key, VersionBumpIrrelevant
	})
	activeKeys := set.New(keys...)
	err := t.repo.ProcessLogWhere(
		ctx,
		func(_ *object.Commit) bool {
			return activeKeys.Count() == 0
		},
		func(ctx context.Context, commit *Commit) (bool, error) {
			log.Trace().Msgf("processing commit %v", commit.ShortHash)

			for _, key := range activeKeys.AsSlice() {
				latestTag := latestTags[key]
				component := t.monorepoConfig.Components[key]

				// if the key has no latest tag, immediately process it (i.e on the latest commit) and remove it from the map
				if latestTag == nil {
					relevant, err := t.commitRelevantToComponent(&component, commit)
					if err != nil {
						return false, err
					}
					if relevant {
						bumpMap[key] = VersionBumpMinor
					}
					activeKeys.Remove(key)
					continue
				}

				// if we've reached the commit that the latest tag corrresponds to, we're done processing commits and the key should be removed from consideration
				if commit.Hash == latestTag.Hash {
					activeKeys.Remove(key)
					continue
				}

				// otherwise we havent gotten to its tag yet, so processthe commit
				oldBump := bumpMap[key]
				newBump, err := t.processCommit(ctx, &component, commit)
				if err != nil {
					return false, err
				}
				if newBump.Greater(oldBump) {
					bumpMap[key] = newBump
				}
			}
			return true, nil
		},
	)
	if err != nil {
		return fmt.Errorf("error walking commit list: %w", err)
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

	// now that we've got all our bumps, walk them again and do our "always patch" logic, log, and make our tags
	for _, key := range keys {
		bump := bumpMap[key]
		component := t.monorepoConfig.Components[key]
		mostRecent := latestTags[key]

		if bump != VersionBumpIrrelevant && VersionBumpPatch.Greater(bump) && component.AlwaysPatch != nil && *component.AlwaysPatch {
			bump = VersionBumpPatch
		}

		log.Info().Msgf("decision for %v is %v", key, bumpMap[key])
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

			wantTags := []string{
				makeTagString(&component, newTag.String()),
			}
			if component.MaintainLatest != nil && *component.MaintainLatest {
				wantTags = append(wantTags, makeLatest(&component))
			}

			if t.dryRun {
				log.Info().Msgf("DRYRUN: would create %v", wantTags)
			} else {
				log.Info().Msgf("creating %v", wantTags)
				if err := t.repo.MakeTagsAtHead(ctx, wantTags...); err != nil {
					return fmt.Errorf("error creating tag: %w", err)
				}
			}
		}
	}

	if t.dryRun {
		log.Info().Msg("DRYRUN: would push tags")
	} else {
		log.Info().Msgf("pushing tags")
		if err := t.repo.PushTags(ctx); err != nil {
			return fmt.Errorf("error pushing tags: %w", err)
		}
	}

	return nil
}

func (t *Tagbot) processCommit(ctx context.Context, component *config.MonoRepoComponent, commit *Commit) (VersionBump, error) {
	log := zerolog.Ctx(ctx)

	relevant, err := t.commitRelevantToComponent(component, commit)
	if err != nil {
		return VersionBump(-1), err
	}

	if !relevant {
		log.Trace().Msgf("commit determined not relevant to %v", component.Name)
		return VersionBumpIrrelevant, nil
	}

	return VersionBumpFromCommitMessage(ctx, commit.Message), nil
}

func (t *Tagbot) commitRelevantToComponent(component *config.MonoRepoComponent, commit *Commit) (bool, error) {
	for _, file := range commit.Files {
		for _, glob := range component.ChangeSetGlobs {
			match, err := doublestar.Match(glob, file)
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
