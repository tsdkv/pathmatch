package match

import (
	"errors"

	"github.com/tsdkv/pathmatch/internal/utils"
	"github.com/tsdkv/pathmatch/pathmatchpb/v1"
)

type MatchOptions struct {
	CaseInsensitive   bool
	KeepFirstVariable bool
}

func StrictMatch(template *pathmatchpb.PathTemplate, path string, opts *MatchOptions) (matched bool, vars map[string]string, err error) {
	pathSegments := utils.Split(path)

	pathIdx := 0
	matched, pathIdx, vars, err = Match(template, pathSegments, opts)

	// If we matched the template, check if we consumed all path segments
	matched = matched && pathIdx == len(pathSegments)

	if !matched {
		vars = nil // Clear vars if not matched
	}

	return
}

func Match(template *pathmatchpb.PathTemplate, pathSegments []string, opts *MatchOptions) (bool, int, map[string]string, error) {
	if template == nil {
		return false, 0, nil, errors.New("template cannot be nil")
	}

	if len(pathSegments) == 0 {
		return len(template.Segments) == 0, 0, nil, nil
	}

	vars := make(map[string]string, len(template.Segments))

	templateIdx := 0
	pathIdx := 0

	for templateIdx < len(template.Segments) && pathIdx < len(pathSegments) {
		segment := template.Segments[templateIdx]
		pathSegment := pathSegments[pathIdx]

		switch s := segment.Segment.(type) {
		case *pathmatchpb.Segment_Literal:
			if !compareStrings(s.Literal.Value, pathSegment, opts.CaseInsensitive) {
				return false, 0, nil, nil
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
				return false, 0, nil, errors.New("double star must be the last segment")
			}
			pathIdx = len(pathSegments) // Move path index to the end
			return true, pathIdx, vars, nil

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
						if pathIdx >= len(pathSegments) || !compareStrings(seg.Literal.Value, pathSegments[pathIdx], opts.CaseInsensitive) {
							return false, 0, nil, nil
						}
						varValue = append(varValue, seg.Literal.Value)
						pathIdx++
					case *pathmatchpb.Segment_DoubleStar:
						// Double star in variable pattern matches all remaining segments
						if i != len(s.Variable.Segments)-1 && templateIdx != len(template.Segments)-1 {
							return false, 0, nil, errors.New("double star must be the last segment in variable pattern")
						}
						// Collect all remaining segments
						varValue = append(varValue, pathSegments[pathIdx:]...)
						vars[s.Variable.Name] = utils.Join(varValue...)
						pathIdx = len(pathSegments) // Move to the end of path segments
						return true, pathIdx, vars, nil
					case *pathmatchpb.Segment_Star:
						// Star in variable pattern matches any single segment
						if pathIdx < len(pathSegments) {
							varValue = append(varValue, pathSegments[pathIdx])
							pathIdx++
						} else {
							return false, 0, nil, nil
						}
					case *pathmatchpb.Segment_Variable:
						return false, 0, nil, errors.New("nested variables in patterns are not allowed")
					default:
						return false, 0, nil, errors.New("unexpected segment type in variable pattern")
					}

				}
				templateIdx++

				_, ok := vars[s.Variable.Name]
				if !ok {
					vars[s.Variable.Name] = utils.Join(varValue...)
				} else if !opts.KeepFirstVariable {
					// If the variable already exists and we're not keeping the first value,
					// overwrite it with the new value.
					vars[s.Variable.Name] = utils.Join(varValue...)
				}
			}
		}
	}

	// Check if we've matched all segments
	if templateIdx != len(template.Segments) {
		return false, 0, nil, nil
	}

	return true, pathIdx, vars, nil
}
