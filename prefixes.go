package main

import (
	"fmt"
	"regexp"
	"github.com/samber/lo"
)

//go:generate go-enum -f $GOFILE -marshal -names

/*
ENUM(
fix
feat
chore
docs
style
refactor
perf
test
ci
improve
)
*/
type CommitPrefix string

var PrefixRegexes = lo.FromEntries(lo.Map(CommitPrefixNames(), func(p string, _ int) lo.Entry[CommitPrefix, *regexp.Regexp] {
	return lo.Entry[CommitPrefix, *regexp.Regexp]{
		// Safe to cast here since we're iterating over the generated names, we can't get a bad one
		Key: CommitPrefix(p),
		Value: regexp.MustCompile(fmt.Sprintf(`(?i)^%v(\(.*\))?!?: .*`, p)),
	}
}))

const (
	BreakingChange = "BREAKING CHANGE"
)

/*
ENUM(
none
patch
minor
major
)
*/
type VersionBump int

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
	AuthMethodToken: RemoteTypeHttps,
	AuthMethodPublicKey: RemoteTypeSsh,
}