package bot

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/nicjohnson145/hlp"
	"github.com/rs/zerolog"
)

const (
	BreakingChange = "BREAKING CHANGE"
)

var (
	InitialTag = hlp.Must(semver.NewVersion("v0.0.1"))
	ErrInvalidMessageError = errors.New("invalid commit message")
)

//go:generate go-enum -f $GOFILE -marshal -names

/*
ENUM(
irrelevant
none
patch
minor
major
)
*/
type VersionBump int

func (v VersionBump) Greater(other VersionBump) bool {
	return int(v) > int(other)
}

var prefixMap = map[string]VersionBump{
	"nop":      VersionBumpNone,
	"fix":      VersionBumpPatch,
	"feat":     VersionBumpMinor,
	"chore":    VersionBumpNone,
	"docs":     VersionBumpNone,
	"style":    VersionBumpNone,
	"refactor": VersionBumpPatch,
	"perf":     VersionBumpPatch,
	"test":     VersionBumpNone,
	"ci":       VersionBumpNone,
}

var messagesRegex = regexp.MustCompile(fmt.Sprintf(`(?i)^(?P<prefix>%v)(\(.*\))?(?P<breaking>!?): .*`, strings.Join(hlp.Keys(prefixMap), "|")))

func VersionBumpFromCommitMessage(ctx context.Context, message string) VersionBump {
	bump, _ := EnsureValidCommitMessage(ctx, message)
	return bump
}

func EnsureValidCommitMessage(ctx context.Context, message string) (VersionBump, error) {
	// If its got the breaking change string in it, nothing else matters, its a breaking change
	if strings.Contains(message, BreakingChange) {
		return VersionBumpMajor, nil
	}

	log := zerolog.Ctx(ctx)

	// Otherwise try to match it up
	parts := hlp.ExtractNamedMatches(messagesRegex, messagesRegex.FindStringSubmatch(message))
	if len(parts) == 0 {
		log.Trace().Msgf("commit message '%v' does not conform to regex, marking as no bump", message)
		return VersionBumpNone, ErrInvalidMessageError
	}

	bump, ok := prefixMap[parts["prefix"]]
	if !ok {
		return VersionBumpNone, nil
	}

	if parts["breaking"] != "" {
		return VersionBumpMajor, nil
	}

	return bump, nil
}
