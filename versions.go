package main

import (
	"github.com/Masterminds/semver"
)

func NewTagForVersionBump(tag *semver.Version, bump VersionBump) semver.Version {
	switch bump {
	case VersionBumpMajor:
		return tag.IncMajor()
	case VersionBumpMinor:
		return tag.IncMinor()
	case VersionBumpPatch:
		return tag.IncPatch()
	default:
		return *tag
	}
}
