package bot

import (
	"context"

	"github.com/Masterminds/semver"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Tag struct {
	Hash    string
	Tag     *semver.Version
	TagName string
}

type IRepo interface {
	GetLatestTag(ctx context.Context, prefix string) (*Tag, error)
	MakeTagsAtHead(ctx context.Context, tags ...string) error
	PushTags(ctx context.Context) error
	IsTagbotDisabled() (bool, error)
	ProcessLogWhere(ctx context.Context, stopFunc func(commit *object.Commit) bool, processFunc CommitProcessFunc) error
}
