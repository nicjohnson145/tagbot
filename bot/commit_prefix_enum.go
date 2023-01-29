// Code generated by go-enum DO NOT EDIT.
// Version: 0.5.4
// Revision: 9793817a5b65af692253b8bc6081fe69a4b6985f
// Build Date: 2022-12-21T19:29:50Z
// Built By: goreleaser

package bot

import (
	"fmt"
	"strings"
)

const (
	// CommitPrefixNop is a CommitPrefix of type nop.
	CommitPrefixNop CommitPrefix = "nop"
	// CommitPrefixFix is a CommitPrefix of type fix.
	CommitPrefixFix CommitPrefix = "fix"
	// CommitPrefixFeat is a CommitPrefix of type feat.
	CommitPrefixFeat CommitPrefix = "feat"
	// CommitPrefixChore is a CommitPrefix of type chore.
	CommitPrefixChore CommitPrefix = "chore"
	// CommitPrefixDocs is a CommitPrefix of type docs.
	CommitPrefixDocs CommitPrefix = "docs"
	// CommitPrefixStyle is a CommitPrefix of type style.
	CommitPrefixStyle CommitPrefix = "style"
	// CommitPrefixRefactor is a CommitPrefix of type refactor.
	CommitPrefixRefactor CommitPrefix = "refactor"
	// CommitPrefixPerf is a CommitPrefix of type perf.
	CommitPrefixPerf CommitPrefix = "perf"
	// CommitPrefixTest is a CommitPrefix of type test.
	CommitPrefixTest CommitPrefix = "test"
	// CommitPrefixCi is a CommitPrefix of type ci.
	CommitPrefixCi CommitPrefix = "ci"
	// CommitPrefixImprove is a CommitPrefix of type improve.
	CommitPrefixImprove CommitPrefix = "improve"
)

var ErrInvalidCommitPrefix = fmt.Errorf("not a valid CommitPrefix, try [%s]", strings.Join(_CommitPrefixNames, ", "))

var _CommitPrefixNames = []string{
	string(CommitPrefixNop),
	string(CommitPrefixFix),
	string(CommitPrefixFeat),
	string(CommitPrefixChore),
	string(CommitPrefixDocs),
	string(CommitPrefixStyle),
	string(CommitPrefixRefactor),
	string(CommitPrefixPerf),
	string(CommitPrefixTest),
	string(CommitPrefixCi),
	string(CommitPrefixImprove),
}

// CommitPrefixNames returns a list of possible string values of CommitPrefix.
func CommitPrefixNames() []string {
	tmp := make([]string, len(_CommitPrefixNames))
	copy(tmp, _CommitPrefixNames)
	return tmp
}

// String implements the Stringer interface.
func (x CommitPrefix) String() string {
	return string(x)
}

// String implements the Stringer interface.
func (x CommitPrefix) IsValid() bool {
	_, err := ParseCommitPrefix(string(x))
	return err == nil
}

var _CommitPrefixValue = map[string]CommitPrefix{
	"nop":      CommitPrefixNop,
	"fix":      CommitPrefixFix,
	"feat":     CommitPrefixFeat,
	"chore":    CommitPrefixChore,
	"docs":     CommitPrefixDocs,
	"style":    CommitPrefixStyle,
	"refactor": CommitPrefixRefactor,
	"perf":     CommitPrefixPerf,
	"test":     CommitPrefixTest,
	"ci":       CommitPrefixCi,
	"improve":  CommitPrefixImprove,
}

// ParseCommitPrefix attempts to convert a string to a CommitPrefix.
func ParseCommitPrefix(name string) (CommitPrefix, error) {
	if x, ok := _CommitPrefixValue[name]; ok {
		return x, nil
	}
	return CommitPrefix(""), fmt.Errorf("%s is %w", name, ErrInvalidCommitPrefix)
}

// MarshalText implements the text marshaller method.
func (x CommitPrefix) MarshalText() ([]byte, error) {
	return []byte(string(x)), nil
}

// UnmarshalText implements the text unmarshaller method.
func (x *CommitPrefix) UnmarshalText(text []byte) error {
	tmp, err := ParseCommitPrefix(string(text))
	if err != nil {
		return err
	}
	*x = tmp
	return nil
}
