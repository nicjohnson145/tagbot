package main

import (
	"strings"
	"errors"
	"regexp"
)

var ErrInvalidMessage = errors.New("invalid commit message")
var mergeRegex = regexp.MustCompile(`^Merge branch`)

func IsValidCommitMessage(msg string) bool {
	_, err := GetCommitType(msg)
	if err != nil {
		return false
	}
	return true
}

func GetCommitType(msg string) (CommitPrefix, error) {
	if mergeRegex.MatchString(msg) {
		return CommitPrefixNop, nil
	}

	for prefix, regex := range PrefixRegexes {
		if regex.MatchString(msg) {
			return prefix, nil
		}
	}
	return CommitPrefix(""), ErrInvalidMessage
}

func IsBreakingChange(msg string, prefix CommitPrefix) bool {
	if strings.Contains(msg, BreakingChange) {
		return true
	}

	if BreakingPrefixes[prefix].MatchString(msg) {
		return true
	}

	return false
}

func CommitMessageToVersionBump(msg string) (VersionBump, error) {
	prefix, err := GetCommitType(msg)
	if err != nil {
		return VersionBump(-1), err
	}

	if IsBreakingChange(msg, prefix) {
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

func GetVersionBumpForCommits(commits []string) (VersionBump, error) {
	bumpType := VersionBumpNone
	for _, commit := range commits {
		bump, err := CommitMessageToVersionBump(commit)
		// If we encounter an invalid message, just move past it and keep checking
		if err != nil && !errors.Is(err, ErrInvalidMessage){
			return VersionBumpNone, err
		}
		if int(bump) > int(bumpType) {
			bumpType = bump
		}
	}

	return bumpType, nil
}
