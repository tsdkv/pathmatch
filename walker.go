package pathmatch

import (
	"path"

	"github.com/tsdkv/pathmatch/pathmatchpb"
)

// Walker facilitates step-by-step traversal and matching of a concrete path
// against a series of path templates. It is designed for scenarios such as
// evaluating hierarchical configurations or security rules, where a specific path
// (e.g., "/users/alice/settings/profile") is incrementally matched against
// templates (e.g., "/users/{userID}", then "/settings/{section}").
//
// The Walker maintains its current position within the concrete path, accumulates
// variables extracted from successful template matches, and tracks the depth of
// traversal. It supports stepping forward through matches, stepping back to
// previous states, and resetting to the initial state.
//
// A Walker is typically initialized with a full concrete path using NewWalker.
// Subsequent calls to Step attempt to consume parts of this path according
// to the provided PathTemplates.
type Walker struct {
	// Current path segments being traversed
	pathSegments []string

	// Current depth in the tree (0 is root)
	currDepth int

	// Current segment index in the path
	segmentIdx int

	// Stack of segment indices for backtracking
	segIdsCheckpoints []int

	// Stack of variable maps for each level
	vars []map[string]string
}

// NewWalker creates and initializes a new Walker for the given concretePath.
// The walker starts at the beginning of the path with no variables captured
// and a depth of 0.
//
// Example:
//
//	walker := NewWalker("/users/alice/settings/profile")
func NewWalker(path string) *Walker {
	segments := Split(path)
	return &Walker{
		pathSegments:      segments,
		currDepth:         0,
		segmentIdx:        0,
		segIdsCheckpoints: []int{0},
		vars:              []map[string]string{{}},
	}
}

// Step attempts to match the provided PathTemplate against the current
// beginning of the Remaining path.
//
// If the template matches:
//   - The walker's internal position advances past the matched segment(s).
//   - Variables captured by this specific template match are returned in stepVars.
//   - These stepVars are also merged into the walker's total Variables().
//   - The walker's Depth is incremented.
//   - matched is true.
//
// If the template does not match:
//   - The walker's state remains unchanged.
//   - stepVars is nil.
//   - matched is false.
//
// Example:
//
//	walker := NewWalker("/users/alice/settings/profile")
//	userTemplate, _ := pathmatch.ParseTemplate("/users/{id}")
//	vars, ok := walker.Step(userTemplate)
//	// vars: map[string]string{"id": "alice"}, ok: true
//	// walker.Remaining(): "/settings/profile"
//	// walker.Variables(): map[string]string{"id": "alice"}
//	// walker.Depth(): 1
func (w *Walker) Step(template *pathmatchpb.PathTemplate) (stepVars map[string]string, matched bool) {
	return
}

// StepBack reverts the Walker to the state it was in before the last successful
// Step operation. This effectively "undoes" the last match.
//
// If a StepBack is possible (i.e., Depth > 0):
//   - The walker's Remaining path, Variables, and Depth are restored.
//   - It returns true.
//
// If no previous steps exist (Depth == 0), the state is unchanged, and it
// returns false.
//
// Example:
//
//	walker.Step(template1) // depth becomes 1
//	walker.Step(template2) // depth becomes 2
//	ok := walker.StepBack() // ok: true, depth becomes 1
//	ok = walker.StepBack() // ok: true, depth becomes 0
//	ok = walker.StepBack() // ok: false, depth remains 0
func (w *Walker) StepBack() bool {
	return false
}

// Reset returns the Walker to its initial state, as if it were newly created
// with the original concretePath. All captured variables are cleared,
// the remaining path is reset to the full concrete path, and depth is set to 0.
// The history for StepBack is also cleared.
func (w *Walker) Reset() {
	return
}

// IsComplete checks if the entire concretePath has been consumed by Step
// operations. It is a convenience method equivalent to checking if
// Remaining() returns an empty string.
func (w *Walker) IsComplete() bool {
	return w.segmentIdx == len(w.pathSegments)
}

// Depth returns the number of successful Step operations performed,
// effectively the current "level" of matching within the path.
// It starts at 0 and increments with each successful Step.
func (w *Walker) Depth() int {
	return w.currDepth
}

// Remaining returns the portion of the original concretePath that has not yet
// been consumed by successful Step operations.
//
// Example:
//
//	walker := NewWalker("/a/b/c")
//	walker.Step(templateForA) // Assuming templateForA matches "/a"
//	fmt.Println(walker.Remaining()) // Output: "/b/c"
func (w *Walker) Remaining() string {
	if w.segmentIdx >= len(w.pathSegments) {
		return ""
	}
	return path.Join(w.pathSegments[w.segmentIdx:]...)
}

// Variables returns a map of all variables accumulated from all successful
// Step operations up to the current point. The keys are variable names from
// the path templates, and values are the matched segments from the concrete path.
// The returned map is a copy; modifications to it will not affect the walker's internal state.
// TODO: check same variable names at different levels
func (w *Walker) Variables() map[string]string {
	vars := make(map[string]string)
	for _, v := range w.vars {
		for k, val := range v {
			if existing, ok := vars[k]; !ok || existing == "" {
				vars[k] = val
			}
		}
	}
	return vars
}
