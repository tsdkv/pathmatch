package pathmatch

import (
	"github.com/tsdkv/pathmatch/internal/match"
	"github.com/tsdkv/pathmatch/pathmatchpb"
)

// Matches path to a parsed template path
// path cant contain wildcards or variables, only literal segments
//
// /path/*/to matches /path/any/to
// /path/{var} matches /path/to and returns map[string]string{"var": "to"}
// /path/{var=**} matches /path/to/with/more and returns map[string]string{"var": "to/with/more"}
func Match(template *pathmatchpb.PathTemplate, path string) (matched bool, vars map[string]string, err error) {
	pathSegments := Split(path)

	pathIdx := 0
	matched, pathIdx, vars, err = match.Match(template, pathSegments)

	// If we matched the template, check if we consumed all path segments
	matched = matched && pathIdx == len(pathSegments)

	if !matched {
		vars = nil // Clear vars if not matched
	}

	return
}
