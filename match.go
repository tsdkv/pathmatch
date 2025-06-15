package pathmatch

import (
	"github.com/tsdkv/pathmatch/internal/match"
	"github.com/tsdkv/pathmatch/pathmatchpb"
)

type MatchOption func(*match.MatchOptions)

// WithCaseInsensitive sets the match options to be case-insensitive.
func WithCaseInsensitive() MatchOption {
	return func(opts *match.MatchOptions) {
		opts.CaseInsensitive = true
	}
}

// Matches path to a parsed template path
// path cant contain wildcards or variables, only literal segments
//
// /path/*/to matches /path/any/to
// /path/{var} matches /path/to and returns map[string]string{"var": "to"}
// /path/{var=**} matches /path/to/with/more and returns map[string]string{"var": "to/with/more"}
func Match(template *pathmatchpb.PathTemplate, path string, opts ...MatchOption) (matched bool, vars map[string]string, err error) {
	mopts := &match.MatchOptions{
		CaseInsensitive: false, // Default to case-sensitive matching
	}
	for _, opt := range opts {
		opt(mopts)
	}

	pathSegments := Split(path)

	pathIdx := 0
	matched, pathIdx, vars, err = match.Match(template, pathSegments, mopts)

	// If we matched the template, check if we consumed all path segments
	matched = matched && pathIdx == len(pathSegments)

	if !matched {
		vars = nil // Clear vars if not matched
	}

	return
}
