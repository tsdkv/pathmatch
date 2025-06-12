package pathmatch

import (
	"errors"
	"strings"

	"github.com/tsdkv/pathmatch/pathmatchpb"
)

// Matches path to a parsed template path
// path cant contain wildcards or variables, only literal segments
//
// /path/*/to matches /path/any/to
// /path/{var} matches /path/to and returns map[string]string{"var": "to"}
// /path/{var=**} matches /path/to/with/more and returns map[string]string{"var": "to/with/more"}
func Match(template *pathmatchpb.PathTemplate, path string) (bool, map[string]string, error) {
	if template == nil {
		return false, nil, errors.New("template cannot be nil")
	}

	if len(path) == 0 {
		return len(template.Segments) == 0, nil, nil
	}

	// Remove leading and trailing slashes without allocation
	path = trimSlashes(path)
	if path == "" {
		return len(template.Segments) == 0, nil, nil
	}

	vars := make(map[string]string, len(template.Segments))

	pathSegments := splitPath(path)

	templateIdx := 0
	pathIdx := 0

	for templateIdx < len(template.Segments) && pathIdx < len(pathSegments) {
		segment := template.Segments[templateIdx]
		pathSegment := pathSegments[pathIdx]

		switch s := segment.Segment.(type) {
		case *pathmatchpb.Segment_Literal:
			if s.Literal.Value != pathSegment {
				return false, nil, nil
			}
			templateIdx++
			pathIdx++

		case *pathmatchpb.Segment_Star:
			// Star matches any single segment
			templateIdx++
			pathIdx++

		case *pathmatchpb.Segment_DoubleStar:
			// Double star matches remaining segments
			if templateIdx != len(template.Segments)-1 {
				return false, nil, errors.New("double star must be the last segment")
			}
			// Double star matches all remaining segments
			return true, vars, nil

		case *pathmatchpb.Segment_Variable:
			if s.Variable.Segments == nil {
				// Simple variable: {var}
				vars[s.Variable.Name] = pathSegment
				templateIdx++
				pathIdx++
			} else {
				// Variable with pattern: {var=pattern}
				// Check if remaining path segments match the variable pattern
				varValue := []string{}
				for i := range s.Variable.Segments {
					switch seg := s.Variable.Segments[i].Segment.(type) {
					case *pathmatchpb.Segment_Literal:
						if pathIdx >= len(pathSegments) || seg.Literal.Value != pathSegments[pathIdx] {
							return false, nil, nil
						}
						varValue = append(varValue, seg.Literal.Value)
						pathIdx++
					case *pathmatchpb.Segment_DoubleStar:
						// Double star in variable pattern matches all remaining segments
						if i != len(s.Variable.Segments)-1 && templateIdx != len(template.Segments)-1 {
							return false, nil, errors.New("double star must be the last segment in variable pattern")
						}
						// Collect all remaining segments
						varValue = append(varValue, pathSegments[pathIdx:]...)
						vars[s.Variable.Name] = strings.Join(varValue, "/")
						return true, vars, nil
					case *pathmatchpb.Segment_Star:
						// Star in variable pattern matches any single segment
						if pathIdx < len(pathSegments) {
							varValue = append(varValue, pathSegments[pathIdx])
							pathIdx++
						} else {
							return false, nil, nil
						}
					case *pathmatchpb.Segment_Variable:
						return false, nil, errors.New("nested variables in patterns are not allowed")
					default:
						return false, nil, errors.New("unexpected segment type in variable pattern")
					}

				}
				vars[s.Variable.Name] = strings.Join(varValue, "/")
				templateIdx++
			}
		}
	}

	// Check if we've matched all segments
	if templateIdx != len(template.Segments) || pathIdx != len(pathSegments) {
		return false, nil, nil
	}

	return true, vars, nil
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
