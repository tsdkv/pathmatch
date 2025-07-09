package pathmatch

import (
	"github.com/tsdkv/pathmatch/internal/parse"
	pmpb "github.com/tsdkv/pathmatch/pathmatchpb/v1"
)

// ParseTemplate parses a path template string into a structured PathTemplate object.
//
// The template string must start with a '/' and may contain:
//   - Literal segments (e.g., "/users")
//   - Wildcard segments: '*' matches any single path segment
//   - Double wildcard: '**' matches zero or more segments, but only as a full segment and only at the end
//   - Variables: '{name}' for a single segment, or '{name=pattern}' where pattern is a sequence of segments
func ParseTemplate(s string) (*pmpb.PathTemplate, error) {
	return parse.ParseTemplate(s)
}
