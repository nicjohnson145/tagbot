package bot

import (
	"fmt"
	"regexp"

	"github.com/samber/lo"
)

//go:generate go-enum -f $GOFILE -marshal -names

/*
ENUM(
nop
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

var prefixRegexes = lo.FromEntries(lo.Map(
	lo.Filter(CommitPrefixNames(), func(p string, _ int) bool {
		return p != CommitPrefixNop.String()
	}),
	func(p string, _ int) lo.Entry[CommitPrefix, *regexp.Regexp] {
		return lo.Entry[CommitPrefix, *regexp.Regexp]{
			// Safe to cast here since we're iterating over the generated names, we can't get a bad one
			Key:   CommitPrefix(p),
			Value: regexp.MustCompile(fmt.Sprintf(`(?i)^%v(\(.*\))?!?: .*`, p)),
		}
	},
))

var breakingPrefixes = lo.FromEntries(lo.Map(
	lo.Filter(CommitPrefixNames(), func(p string, _ int) bool {
		return p != CommitPrefixNop.String()
	}),
	func(p string, _ int) lo.Entry[CommitPrefix, *regexp.Regexp] {
		return lo.Entry[CommitPrefix, *regexp.Regexp]{
			// Safe to cast here since we're iterating over the generated names, we can't get a bad one
			Key:   CommitPrefix(p),
			Value: regexp.MustCompile(fmt.Sprintf(`(?i)^%v(\(.*\))?!: .*`, p)),
		}
	},
))
