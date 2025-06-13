package pathmatch

import (
	"strings"
)

// Split splits a path string into its segments.
// It handles leading/trailing slashes and multiple slashes between segments.
// For example, Split("/users//alice/") returns ["users", "alice"].
// An empty path or a path consisting only of slashes results in an empty slice.
func Split(path string) []string {
	trimmedPath := strings.Trim(path, "/")
	if trimmedPath == "" {
		return []string{}
	}
	// Split by slash and filter out empty strings resulting from multiple slashes
	rawSegments := strings.Split(trimmedPath, "/")
	segments := make([]string, 0, len(rawSegments))
	for _, s := range rawSegments {
		if s != "" {
			segments = append(segments, s)
		}
	}
	return segments
}

// Join combines path segments into a single path string.
// It ensures segments are joined by a single slash and prepends a leading slash.
// If no segments are provided, it returns "/".
// Example: Join("users", "alice", "profile") returns "/users/alice/profile".
func Join(segments ...string) string {
	if len(segments) == 0 {
		return "/" // Or "" depending on desired behavior for empty join
	}
	// Filter out empty segments to avoid multiple slashes like "/a//b"
	// if an empty string was passed in `segments`.
	validSegments := make([]string, 0, len(segments))
	for _, s := range segments {
		trimmed := strings.Trim(s, "/") // Avoid issues if segments themselves have slashes
		if trimmed != "" {
			validSegments = append(validSegments, trimmed)
		}
	}
	if len(validSegments) == 0 { // If all segments were empty or slashes
		return "/"
	}
	return "/" + strings.Join(validSegments, "/")
}

// Clean normalizes a path string by removing redundant slashes and
// any trailing slash (unless it's the root path "/").
// For example, Clean("/users//alice///") returns "/users/alice".
// Clean("/") returns "/".
func Clean(path string) string {
	if path == "" {
		return "" // Or "." or "/" depending on convention
	}
	if path == "/" {
		return "/"
	}

	// Use Split and Join for a robust cleaning mechanism
	// This inherently handles multiple slashes and leading/trailing ones.
	segments := Split(path)
	if len(segments) == 0 { // Path was like "/" or "///"
		return "/"
	}
	return Join(segments...) // Join will add the leading slash
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
