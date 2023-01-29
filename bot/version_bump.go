package bot

//go:generate go-enum -f $GOFILE -marshal -names

/*
ENUM(
none
patch
minor
major
)
*/
type VersionBump int
