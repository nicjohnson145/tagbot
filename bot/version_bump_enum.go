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
	// VersionBumpNone is a VersionBump of type None.
	VersionBumpNone VersionBump = iota
	// VersionBumpPatch is a VersionBump of type Patch.
	VersionBumpPatch
	// VersionBumpMinor is a VersionBump of type Minor.
	VersionBumpMinor
	// VersionBumpMajor is a VersionBump of type Major.
	VersionBumpMajor
)

var ErrInvalidVersionBump = fmt.Errorf("not a valid VersionBump, try [%s]", strings.Join(_VersionBumpNames, ", "))

const _VersionBumpName = "nonepatchminormajor"

var _VersionBumpNames = []string{
	_VersionBumpName[0:4],
	_VersionBumpName[4:9],
	_VersionBumpName[9:14],
	_VersionBumpName[14:19],
}

// VersionBumpNames returns a list of possible string values of VersionBump.
func VersionBumpNames() []string {
	tmp := make([]string, len(_VersionBumpNames))
	copy(tmp, _VersionBumpNames)
	return tmp
}

var _VersionBumpMap = map[VersionBump]string{
	VersionBumpNone:  _VersionBumpName[0:4],
	VersionBumpPatch: _VersionBumpName[4:9],
	VersionBumpMinor: _VersionBumpName[9:14],
	VersionBumpMajor: _VersionBumpName[14:19],
}

// String implements the Stringer interface.
func (x VersionBump) String() string {
	if str, ok := _VersionBumpMap[x]; ok {
		return str
	}
	return fmt.Sprintf("VersionBump(%d)", x)
}

var _VersionBumpValue = map[string]VersionBump{
	_VersionBumpName[0:4]:   VersionBumpNone,
	_VersionBumpName[4:9]:   VersionBumpPatch,
	_VersionBumpName[9:14]:  VersionBumpMinor,
	_VersionBumpName[14:19]: VersionBumpMajor,
}

// ParseVersionBump attempts to convert a string to a VersionBump.
func ParseVersionBump(name string) (VersionBump, error) {
	if x, ok := _VersionBumpValue[name]; ok {
		return x, nil
	}
	return VersionBump(0), fmt.Errorf("%s is %w", name, ErrInvalidVersionBump)
}

// MarshalText implements the text marshaller method.
func (x VersionBump) MarshalText() ([]byte, error) {
	return []byte(x.String()), nil
}

// UnmarshalText implements the text unmarshaller method.
func (x *VersionBump) UnmarshalText(text []byte) error {
	name := string(text)
	tmp, err := ParseVersionBump(name)
	if err != nil {
		return err
	}
	*x = tmp
	return nil
}