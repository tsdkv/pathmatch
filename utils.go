package pathmatch

import (
	"github.com/tsdkv/pathmatch/internal/utils"
)

// Split splits a path string into its segments.
// It handles leading/trailing slashes and multiple slashes between segments.
// For example, Split("/users//alice/") returns ["users", "alice"].
// An empty path or a path consisting only of slashes results in an empty slice.
func Split(path string) []string {
	return utils.Split(path)
}

// Join combines path segments into a single path string.
// It ensures segments are joined by a single slash and prepends a leading slash.
// If no segments are provided, it returns "/".
// Example: Join("users", "alice", "profile") returns "/users/alice/profile".
func Join(segments ...string) string {
	return utils.Join(segments...)
}

// Clean normalizes a path string by removing redundant slashes and
// any trailing slash (unless it's the root path "/").
// For example, Clean("/users//alice///") returns "/users/alice".
// Clean("/") returns "/".
func Clean(path string) string {
	return utils.Clean(path)
}

// CompileAndMatch parses the templatePattern string and then matches it against the given path.
// It's a convenience wrapper around ParseTemplate and Match.
func CompileAndMatch(templatePattern string, path string) (matched bool, vars map[string]string, err error) {
	tmpl, err := ParseTemplate(templatePattern)
	if err != nil {
		return false, nil, err
	}
	return Match(tmpl, path)
}
