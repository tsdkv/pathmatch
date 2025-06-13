package pathmatch

import (
	"strings"

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
	path = trimSlashes(path)
	if path == "" {
		return len(template.Segments) == 0, nil, nil
	}

	pathSegments := splitPath(path)

	pathIdx := 0
	matched, pathIdx, vars, err = match.Match(template, pathSegments)

	// If we matched the template, check if we consumed all path segments
	matched = matched && pathIdx == len(pathSegments)

	if !matched {
		vars = nil // Clear vars if not matched
	}

	return
}

// trimSlashes removes leading and trailing slashes without allocation
func trimSlashes(s string) string {
	start := 0
	end := len(s)
	for start < end && s[start] == '/' {
		start++
	}
	for end > start && s[end-1] == '/' {
		end--
	}
	return s[start:end]
}

// splitPath splits a path into segments without allocation
func splitPath(path string) []string {
	if path == "" {
		return nil
	}
	segments := make([]string, 0, strings.Count(path, "/")+1)
	start := 0
	for i := range len(path) {
		if path[i] == '/' {
			if i > start {
				segments = append(segments, path[start:i])
			}
			start = i + 1
		}
	}
	if start < len(path) {
		segments = append(segments, path[start:])
	}
	return segments
}
