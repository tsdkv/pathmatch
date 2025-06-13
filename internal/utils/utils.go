package utils

import "strings"

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
