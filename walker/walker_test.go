package walker_test

import (
	"testing"

	"github.com/tsdkv/pathmatch/internal/parse"
	pmpb "github.com/tsdkv/pathmatch/pathmatchpb/v1"
	pwalker "github.com/tsdkv/pathmatch/walker"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustParseTemplate(t *testing.T, pattern string) *pmpb.PathTemplate {
	t.Helper()
	tmpl, err := parse.ParseTemplate(pattern)
	require.NoError(t, err)
	return tmpl
}

func TestWalker_Step(t *testing.T) {
	templateUser := mustParseTemplate(t, "/users/{id}")
	templateSettings := mustParseTemplate(t, "/settings/{section}")
	templateInvalid := mustParseTemplate(t, "/does/not/match")

	t.Run("SingleSuccessfulStep", func(t *testing.T) {
		walker := pwalker.NewWalker("/users/alice/settings/profile")
		stepVars, matched, err := walker.Step(templateUser)

		require.NoError(t, err)
		assert.True(t, matched)
		assert.Equal(t, map[string]string{"id": "alice"}, stepVars)
		assert.Equal(t, 1, walker.Depth())
		assert.Equal(t, map[string]string{"id": "alice"}, walker.Variables())
		assert.Equal(t, "/settings/profile", walker.Remaining())
	})

	t.Run("MultipleSuccessfulSteps", func(t *testing.T) {
		walker := pwalker.NewWalker("/users/alice/settings/profile")
		_, _, _ = walker.Step(templateUser)
		stepVars, matched, err := walker.Step(templateSettings)

		require.NoError(t, err)
		assert.True(t, matched)
		assert.Equal(t, map[string]string{"section": "profile"}, stepVars)
		assert.Equal(t, 2, walker.Depth())
		assert.Equal(t, map[string]string{"id": "alice", "section": "profile"}, walker.Variables())
		assert.Equal(t, "", walker.Remaining())
		assert.True(t, walker.IsComplete())
	})

	t.Run("UnsuccessfulStep", func(t *testing.T) {
		walker := pwalker.NewWalker("/users/alice")
		initialDepth := walker.Depth()
		initialVars := walker.Variables()
		initialRemaining := walker.Remaining()

		stepVars, matched, err := walker.Step(templateInvalid)

		require.NoError(t, err)
		assert.False(t, matched)
		assert.Nil(t, stepVars)
		assert.Equal(t, initialDepth, walker.Depth())
		assert.Equal(t, initialVars, walker.Variables())
		assert.Equal(t, initialRemaining, walker.Remaining())
	})

	t.Run("StepWithDoubleStar", func(t *testing.T) {
		walker := pwalker.NewWalker("/a/b/c/d/e")
		template := mustParseTemplate(t, "/{first}/{rest=**}")
		stepVars, matched, err := walker.Step(template)

		require.NoError(t, err)
		assert.True(t, matched)
		assert.Equal(t, map[string]string{"first": "a", "rest": "/b/c/d/e"}, stepVars)
		assert.True(t, walker.IsComplete())
		assert.Equal(t, "", walker.Remaining())
	})
}

func TestWalker_StepBack(t *testing.T) {
	templateUser := mustParseTemplate(t, "/users/{id}")
	templateSettings := mustParseTemplate(t, "/settings/{section}")

	t.Run("StepBackAfterOneStep", func(t *testing.T) {
		walker := pwalker.NewWalker("/users/alice/settings/profile")
		_, _, _ = walker.Step(templateUser) // Depth 1, vars {"id":"alice"}, remaining "/settings/profile"

		require.Equal(t, 1, walker.Depth())

		steppedBack := walker.StepBack()
		assert.True(t, steppedBack)
		assert.Equal(t, 0, walker.Depth())
		assert.Empty(t, walker.Variables())
		assert.Equal(t, "/users/alice/settings/profile", walker.Remaining())
	})

	t.Run("StepBackAfterMultipleSteps", func(t *testing.T) {
		walker := pwalker.NewWalker("/users/alice/settings/profile")
		_, _, _ = walker.Step(templateUser) // Depth 1
		varsAfterFirstStep := walker.Variables()
		depthAfterFirstStep := walker.Depth()
		remainingAfterFirstStep := walker.Remaining()

		_, _, _ = walker.Step(templateSettings) // Depth 2

		require.Equal(t, 2, walker.Depth())

		steppedBack := walker.StepBack() // Back to Depth 1
		assert.True(t, steppedBack)
		assert.Equal(t, depthAfterFirstStep, walker.Depth())
		assert.Equal(t, varsAfterFirstStep, walker.Variables())
		assert.Equal(t, remainingAfterFirstStep, walker.Remaining())

		steppedBack = walker.StepBack() // Back to Depth 0
		assert.True(t, steppedBack)
		assert.Equal(t, 0, walker.Depth())
		assert.Empty(t, walker.Variables())
		assert.Equal(t, "/users/alice/settings/profile", walker.Remaining())
	})

	t.Run("StepBackAtDepthZero", func(t *testing.T) {
		walker := pwalker.NewWalker("/users/alice")
		steppedBack := walker.StepBack()
		assert.False(t, steppedBack)
		assert.Equal(t, 0, walker.Depth())
		assert.Equal(t, "/users/alice", walker.Remaining())
	})
}

func TestWalker_Reset(t *testing.T) {
	originalPath := "/users/alice/settings/profile"
	walker := pwalker.NewWalker(originalPath)
	templateUser := mustParseTemplate(t, "/users/{id}")
	templateSettings := mustParseTemplate(t, "/settings/{section}")

	_, _, _ = walker.Step(templateUser)
	_, _, _ = walker.Step(templateSettings)

	require.True(t, walker.IsComplete())

	walker.Reset()

	assert.Equal(t, 0, walker.Depth())
	assert.Empty(t, walker.Variables())
	assert.Equal(t, originalPath, walker.Remaining())
	assert.False(t, walker.IsComplete())
}

func TestWalker_IsComplete(t *testing.T) {
	walker := pwalker.NewWalker("/a/b")
	templateA := mustParseTemplate(t, "/a")
	templateB := mustParseTemplate(t, "/b")

	assert.False(t, walker.IsComplete())
	_, _, _ = walker.Step(templateA)
	assert.False(t, walker.IsComplete())
	_, _, _ = walker.Step(templateB)
	assert.True(t, walker.IsComplete())
	walker.StepBack()
	assert.False(t, walker.IsComplete())

	emptyWalker := pwalker.NewWalker("/") // Split("/") -> []
	assert.True(t, emptyWalker.IsComplete())
	assert.Equal(t, "", emptyWalker.Remaining())

	// TODO: maybe we should fail on paths without leading slash?
	emptyWalker2 := pwalker.NewWalker("") // Split("") -> []
	assert.True(t, emptyWalker2.IsComplete())
	assert.Equal(t, "", emptyWalker2.Remaining())
}

func TestWalker_Depth(t *testing.T) {
	walker := pwalker.NewWalker("/a/b/c")
	template1 := mustParseTemplate(t, "/a")
	template2 := mustParseTemplate(t, "/b")

	assert.Equal(t, 0, walker.Depth())
	_, _, _ = walker.Step(template1)
	assert.Equal(t, 1, walker.Depth())
	_, _, _ = walker.Step(template2)
	assert.Equal(t, 2, walker.Depth())
	walker.StepBack()
	assert.Equal(t, 1, walker.Depth())
	walker.Reset()
	assert.Equal(t, 0, walker.Depth())
}

func TestWalker_Remaining(t *testing.T) {
	path := "/users/alice/settings"
	walker := pwalker.NewWalker(path)
	templateUser := mustParseTemplate(t, "/users/{id}")
	templateSettings := mustParseTemplate(t, "/settings") // Matches only "/settings"

	assert.Equal(t, path, walker.Remaining())

	_, _, _ = walker.Step(templateUser) // Consumes "/users/alice"
	assert.Equal(t, "/settings", walker.Remaining())

	_, _, _ = walker.Step(templateSettings) // Consumes "/settings"
	assert.Equal(t, "", walker.Remaining())
	assert.True(t, walker.IsComplete())

	walker.StepBack() // Back to after templateUser match
	assert.Equal(t, "/settings", walker.Remaining())

	walker.Reset()
	assert.Equal(t, path, walker.Remaining())

	// Test Remaining() for root and empty paths
	rootWalker := pwalker.NewWalker("/")
	assert.True(t, rootWalker.IsComplete()) // Path segments are empty
	assert.Equal(t, "", rootWalker.Remaining())

	emptyWalker := pwalker.NewWalker("")
	assert.True(t, emptyWalker.IsComplete()) // Path segments are empty
	assert.Equal(t, "", emptyWalker.Remaining())
}

// TODO: fix this when we have a better way to handle variables
func TestWalker_Variables(t *testing.T) {
	walker := pwalker.NewWalker("/users/alice/settings/profile")
	templateUser := mustParseTemplate(t, "/users/{id}")
	templateSettings := mustParseTemplate(t, "/settings/{section}")
	// Template to test variable accumulation logic: first non-empty value for a key is kept.
	_ = mustParseTemplate(t, "/users/{id=groupX}")

	assert.Empty(t, walker.Variables())

	_, _, _ = walker.Step(templateUser) // id:alice
	assert.Equal(t, map[string]string{"id": "alice"}, walker.Variables())

	_, _, _ = walker.Step(templateSettings) // section:profile
	assert.Equal(t, map[string]string{"id": "alice", "section": "profile"}, walker.Variables())

	// Test variable accumulation logic (first non-empty value for a key is kept)
	// We need to simulate steps that would lead to this.
	// Scenario:
	// 1. Step with /prefix/{id=foo}/suffix -> id: "foo"
	// 2. Step with /prefix/{id}/suffix -> id: "" (if path segment is empty)
	// The Variables() should still report id: "foo"
	walkerForVarLogic := pwalker.NewWalker("/prefix/val1/mid//suffix") // val1, then empty segment
	tplVal1 := mustParseTemplate(t, "/prefix/{id}/mid")
	tplEmpty := mustParseTemplate(t, "/{id}/suffix") // This will match the empty segment

	_, _, err := walkerForVarLogic.Step(tplVal1) // id: "val1"
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"id": "val1"}, walkerForVarLogic.Variables())

	_, _, err = walkerForVarLogic.Step(tplEmpty) // id: ""
	require.NoError(t, err)
	// if "id" was "val1", and a new step provides "id":"", "val1" should be kept.
	assert.Equal(t, map[string]string{"id": "val1"}, walkerForVarLogic.Variables())

	walker.StepBack() // Back to after user step (id:alice)
	assert.Equal(t, map[string]string{"id": "alice"}, walker.Variables())

	walker.Reset()
	assert.Empty(t, walker.Variables())
}

func TestWalker_Remaining_EdgeCases(t *testing.T) {
	tests := []struct {
		name                   string
		initialPath            string
		steps                  []string // template patterns to step through
		expectedFinalRemaining string
		expectComplete         bool
	}{
		{
			name:                   "Empty path",
			initialPath:            "",
			steps:                  []string{},
			expectedFinalRemaining: "",
			expectComplete:         true,
		},
		{
			name:                   "Root path",
			initialPath:            "/",
			steps:                  []string{},
			expectedFinalRemaining: "",
			expectComplete:         true,
		},
		{
			name:                   "Single segment path, fully consumed",
			initialPath:            "/a",
			steps:                  []string{"/a"},
			expectedFinalRemaining: "",
			expectComplete:         true,
		},
		{
			name:                   "Single segment path, var consumed",
			initialPath:            "/a",
			steps:                  []string{"/{foo}"},
			expectedFinalRemaining: "",
			expectComplete:         true,
		},
		{
			name:                   "Multi-segment path, partially consumed",
			initialPath:            "/a/b/c",
			steps:                  []string{"/a", "/{next}"}, // consumes /a, then /b
			expectedFinalRemaining: "/c",
			expectComplete:         false,
		},
		{
			name:                   "Multi-segment path, fully consumed",
			initialPath:            "/a/b/c",
			steps:                  []string{"/a", "/b", "/c"},
			expectedFinalRemaining: "",
			expectComplete:         true,
		},
		{
			name:                   "Path with double star, fully consumed",
			initialPath:            "/a/b/c/d",
			steps:                  []string{"/a/{rest=**}"},
			expectedFinalRemaining: "",
			expectComplete:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			walker := pwalker.NewWalker(tt.initialPath)
			for i, templatePattern := range tt.steps {
				template := mustParseTemplate(t, templatePattern)
				_, matched, err := walker.Step(template)
				require.NoError(t, err, "Step %d with template '%s' failed", i+1, templatePattern)
				require.True(t, matched, "Step %d with template '%s' should have matched", i+1, templatePattern)
			}

			assert.Equal(t, tt.expectedFinalRemaining, walker.Remaining())
			assert.Equal(t, tt.expectComplete, walker.IsComplete())
		})
	}
}

func TestWalkerBuilder_WithOptions(t *testing.T) {
	t.Run("WithKeepFirstVariable", func(t *testing.T) {
		// By default, last-in-wins
		builder := pwalker.NewWalkerBuilder("/a/b")
		walker, _ := builder.Build()
		_, _, _ = walker.Step(mustParseTemplate(t, "/{id}")) // id="a"
		_, _, _ = walker.Step(mustParseTemplate(t, "/{id}")) // id="b"
		assert.Equal(t, map[string]string{"id": "b"}, walker.Variables())

		// With the option, first-in-wins
		builder = pwalker.NewWalkerBuilder("/a/b")
		walker, _ = builder.WithKeepFirstVariable().Build()
		_, _, _ = walker.Step(mustParseTemplate(t, "/{id}")) // id="a"
		_, _, _ = walker.Step(mustParseTemplate(t, "/{id}")) // id="b"
		assert.Equal(t, map[string]string{"id": "a"}, walker.Variables())
	})

	t.Run("WithCaseIncensitive", func(t *testing.T) {
		builder := pwalker.NewWalkerBuilder("/USERS/ALICE")
		walker, _ := builder.WithCaseIncensitive().Build()
		template := mustParseTemplate(t, "/users/{name}")
		vars, matched, err := walker.Step(template)

		require.NoError(t, err)
		assert.True(t, matched)
		assert.Equal(t, map[string]string{"name": "ALICE"}, vars)
	})
}
