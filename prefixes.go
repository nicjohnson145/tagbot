package main

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

const (
	BreakingChange = "BREAKING CHANGE"
)
