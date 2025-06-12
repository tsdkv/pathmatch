package pathmatch_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tsdkv/pathmatch"
)

func equalVars(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || v != bv {
			return false
		}
	}
	return true
}

func TestMatch(t *testing.T) {
	tests := []struct {
		templateStr   string
		path          string
		expectedMatch bool
		expectedVars  map[string]string
	}{
		{
			templateStr:   "/a",
			path:          "/a",
			expectedMatch: true,
		},
		{
			templateStr:   "/path/to/resource",
			path:          "/path/to/resource",
			expectedMatch: true,
			expectedVars:  nil,
		},
		{
			templateStr:   "/path/to/resource",
			path:          "/path/to/another/resource",
			expectedMatch: false,
		},
		{
			templateStr:   "/path/*/to",
			path:          "/path/to/another/to",
			expectedMatch: false,
		},
		{
			templateStr:   "/path/{var}",
			path:          "/path/to/another",
			expectedMatch: false,
		},
		{
			templateStr:   "/path/*/to",
			path:          "/path/any/to",
			expectedMatch: true,
		},
		{
			templateStr:   "/{var}",
			path:          "/value",
			expectedMatch: true,
			expectedVars:  map[string]string{"var": "value"},
		},
		{
			templateStr:   "/path/{var}",
			path:          "/path/to",
			expectedMatch: true,
			expectedVars:  map[string]string{"var": "to"},
		},
		{
			templateStr:   "/**",
			path:          "/path/to/resource",
			expectedMatch: true,
		},
		{
			templateStr:   "/path/**",
			path:          "/path/to/with/more",
			expectedMatch: true,
		},
		{
			templateStr:   "/path/{var=**}",
			path:          "/path/to/with/more",
			expectedMatch: true,
			expectedVars:  map[string]string{"var": "to/with/more"},
		},
		{
			templateStr:   "/path/{var1}/{var2=/hello/*}/world",
			path:          "/path/value1/hello/value2/world",
			expectedMatch: true,
			expectedVars:  map[string]string{"var1": "value1", "var2": "hello/value2"},
		},
		{
			templateStr:   "/path/with/var/{var1=sub/path/**}",
			path:          "/path/with/var/sub/path/to/resource",
			expectedMatch: true,
			expectedVars:  map[string]string{"var1": "sub/path/to/resource"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.templateStr+"_"+tt.path, func(t *testing.T) {
			template, err := pathmatch.ParseTemplate(tt.templateStr)
			require.NoError(t, err, "failed to parse template: %v", err)

			match, vars, err := pathmatch.Match(template, tt.path)
			require.NoError(t, err, "failed to match path: %v", err)
			require.Equal(t, tt.expectedMatch, match, "expected match to be %v", tt.expectedMatch)
			require.True(t, equalVars(vars, tt.expectedVars), "expected vars to be %v, got %v", tt.expectedVars, vars)
		})
	}
}

func BenchmarkMatch(b *testing.B) {
	template, err := pathmatch.ParseTemplate("/path/{var=**}")
	if err != nil {
		b.Fatalf("failed to parse template: %v", err)
	}

	for b.Loop() {
		_, _, err := pathmatch.Match(template, "/path/to/with/more/segments")
		if err != nil {
			b.Fatalf("failed to match path: %v", err)
		}
	}
}
