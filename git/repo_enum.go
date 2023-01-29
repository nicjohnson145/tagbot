// Code generated by go-enum DO NOT EDIT.
// Version: 0.5.4
// Revision: 9793817a5b65af692253b8bc6081fe69a4b6985f
// Build Date: 2022-12-21T19:29:50Z
// Built By: goreleaser

package git

import (
	"fmt"
	"strings"
)

const (
	// AuthMethodPublicKey is a AuthMethod of type public-key.
	AuthMethodPublicKey AuthMethod = "public-key"
	// AuthMethodToken is a AuthMethod of type token.
	AuthMethodToken AuthMethod = "token"
)

var ErrInvalidAuthMethod = fmt.Errorf("not a valid AuthMethod, try [%s]", strings.Join(_AuthMethodNames, ", "))

var _AuthMethodNames = []string{
	string(AuthMethodPublicKey),
	string(AuthMethodToken),
}

// AuthMethodNames returns a list of possible string values of AuthMethod.
func AuthMethodNames() []string {
	tmp := make([]string, len(_AuthMethodNames))
	copy(tmp, _AuthMethodNames)
	return tmp
}

// String implements the Stringer interface.
func (x AuthMethod) String() string {
	return string(x)
}

// String implements the Stringer interface.
func (x AuthMethod) IsValid() bool {
	_, err := ParseAuthMethod(string(x))
	return err == nil
}

var _AuthMethodValue = map[string]AuthMethod{
	"public-key": AuthMethodPublicKey,
	"token":      AuthMethodToken,
}

// ParseAuthMethod attempts to convert a string to a AuthMethod.
func ParseAuthMethod(name string) (AuthMethod, error) {
	if x, ok := _AuthMethodValue[name]; ok {
		return x, nil
	}
	return AuthMethod(""), fmt.Errorf("%s is %w", name, ErrInvalidAuthMethod)
}

// MarshalText implements the text marshaller method.
func (x AuthMethod) MarshalText() ([]byte, error) {
	return []byte(string(x)), nil
}

// UnmarshalText implements the text unmarshaller method.
func (x *AuthMethod) UnmarshalText(text []byte) error {
	tmp, err := ParseAuthMethod(string(text))
	if err != nil {
		return err
	}
	*x = tmp
	return nil
}

const (
	// RemoteTypeSsh is a RemoteType of type ssh.
	RemoteTypeSsh RemoteType = "ssh"
	// RemoteTypeHttps is a RemoteType of type https.
	RemoteTypeHttps RemoteType = "https"
)

var ErrInvalidRemoteType = fmt.Errorf("not a valid RemoteType, try [%s]", strings.Join(_RemoteTypeNames, ", "))

var _RemoteTypeNames = []string{
	string(RemoteTypeSsh),
	string(RemoteTypeHttps),
}

// RemoteTypeNames returns a list of possible string values of RemoteType.
func RemoteTypeNames() []string {
	tmp := make([]string, len(_RemoteTypeNames))
	copy(tmp, _RemoteTypeNames)
	return tmp
}

// String implements the Stringer interface.
func (x RemoteType) String() string {
	return string(x)
}

// String implements the Stringer interface.
func (x RemoteType) IsValid() bool {
	_, err := ParseRemoteType(string(x))
	return err == nil
}

var _RemoteTypeValue = map[string]RemoteType{
	"ssh":   RemoteTypeSsh,
	"https": RemoteTypeHttps,
}

// ParseRemoteType attempts to convert a string to a RemoteType.
func ParseRemoteType(name string) (RemoteType, error) {
	if x, ok := _RemoteTypeValue[name]; ok {
		return x, nil
	}
	return RemoteType(""), fmt.Errorf("%s is %w", name, ErrInvalidRemoteType)
}

// MarshalText implements the text marshaller method.
func (x RemoteType) MarshalText() ([]byte, error) {
	return []byte(string(x)), nil
}

// UnmarshalText implements the text unmarshaller method.
func (x *RemoteType) UnmarshalText(text []byte) error {
	tmp, err := ParseRemoteType(string(text))
	if err != nil {
		return err
	}
	*x = tmp
	return nil
}