package bot

import (
	"context"

	"github.com/Masterminds/semver"
)

type Tag struct {
	Hash string
	Tag  *semver.Version
}

type IRepo interface {
	GetLatestTag(ctx context.Context, prefix string) (*Tag, error)
	ProcessCommitsUntil(ctx context.Context, hashStr string, processFunc CommitProcessFunc) error
	ProcessLatestCommit(ctx context.Context, processFunc CommitProcessFunc) error
	MakeTagsAtHead(ctx context.Context, tags ...string) error
	PushTags(ctx context.Context) error
	IsTagbotDisabled() (bool, error)
}
