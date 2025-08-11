package common

const ZeroCommit = "0000000000000000000000000000000000000000"
const EmptyTree = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

// <type>(<scope>): <description> (#<issue-number>)
const Pattern = `^(feat|fix|chore|docs|style|refactor|test|build|ci|perf)(?:!\([a-z0-9\-]+\)|\([a-z0-9\-]+\)!|\([a-z0-9\-]+\)):\s.+ \(#[0-9]+\)$`
const Example = `feat(ui): add button (#123)`

const (
	FlagConfig          = "config"
	FlagSrcDir          = "src-dir"
	FlagFunctionsSubDir = "functions-subdir"
	FlagServicesSubDir  = "services-subdir"
)
