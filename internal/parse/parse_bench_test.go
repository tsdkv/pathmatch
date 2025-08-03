package parse_test

import (
	"fmt"
	"testing"

	"github.com/tsdkv/pathmatch/internal/parse"
)

// Benchmark different path template patterns
func BenchmarkParseTemplate(b *testing.B) {
	testCases := []struct {
		name     string
		template string
	}{
		{"simple", "/users/123"},
		{"with_wildcard", "/users/*/posts"},
		{"with_variable", "/users/{id}"},
		{"complex_variable", "/users/{id=*/posts/*}"},
		{"double_star", "/users/**"},
		{"long_path", "/api/v1/users/123/posts/456/comments/789/replies"},
		{"many_variables", "/users/{user_id}/posts/{post_id}/comments/{comment_id}"},
		{"nested_variable", "/users/{user_id=users/*/profile}/settings"},
		{"many_literals", "/api/v1/users/123/posts/456/comments/789/replies/101112"},
		{"many_variables", "/users/{a}/posts/{b}/comments/{c}/replies/{d}/likes/{e}"},
		{"mixed_complex", "/api/{version=v*/data/*}/users/{id=users/*/profile}/posts/**"},
		{"deep_nesting", "/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t/u/v/w/x/y/z"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for b.Loop() {
				_, err := parse.ParseTemplate(tc.template)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Benchmark memory allocations
func BenchmarkParseTemplateAllocs(b *testing.B) {
	template := "/users/{user_id}/posts/{post_id=posts/*/comments/*}/replies/**"

	b.ReportAllocs()
	for b.Loop() {
		_, err := parse.ParseTemplate(template)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark with varying input sizes
func BenchmarkParseTemplateBySize(b *testing.B) {
	// Generate templates of different sizes
	sizes := []int{10, 50, 100, 500, 2000}

	for _, size := range sizes {
		template := generateTemplate(size)
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			b.ResetTimer()
			for b.Loop() {
				_, err := parse.ParseTemplate(template)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func generateTemplate(approxLength int) string {
	template := "/api/v1"
	patterns := []string{"/users/{id}", "/posts/*", "/comments/**", "/data/literal"}

	for len(template) < approxLength {
		template += patterns[len(template)%len(patterns)]
	}
	return template
}
