package git

import (
	"github.com/Masterminds/semver"
)

const (
	sshPrefix   = "git@"
	httpsPrefix = "https://"
)

type Tag struct {
	Hash string
	Tag  *semver.Version
}

type Repo interface {
	LatestTag() (*Tag, error)
	CommitsSinceHash(hash string) ([]string, error)
	MakeTagHead(name string) error
	RemakeTagHead(name string) error
	PushTags() error
	ForcePushTags() error
	GetHashForBranch(branch string) (string, error)
	IsTagbotDisabled() (bool, error)
}

//go:generate go-enum -f $GOFILE -marshal -names

/*
ENUM(
ssh
https
)
*/
type RemoteType string

/*
ENUM(
public-key
token
)
*/
type AuthMethod string

var AuthToRemoteMap = map[AuthMethod]RemoteType{
	AuthMethodToken:     RemoteTypeHttps,
	AuthMethodPublicKey: RemoteTypeSsh,
}
