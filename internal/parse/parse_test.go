package parse_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/testing/protocmp"

	pmpb "github.com/tsdkv/pathmatch/pathmatchpb/v1"

	"github.com/tsdkv/pathmatch/internal/parse"
)

func TestLexerSimple(t *testing.T) {
	lex := parse.NewLexer("/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t/u/v/w/x/y/z")
	if lex == nil {
		t.Fatal("lexer should not be nil")
	}

	for i := range 25 {
		if !lex.Match(parse.TokenSlash) {
			t.Fatalf("expected token type %v, got %v", parse.TokenSlash, lex.Peek().Type)
		}
		require.Equal(t, string(byte('a'+i)), lex.Peek().Value, "expected token value to match the expected character")
		if !lex.Match(parse.TokenLiteral) {
			t.Fatalf("expected token type %v, got %v", parse.TokenLiteral, lex.Peek().Type)
		}
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		input    string
		expected pmpb.PathTemplate
	}{
		{
			input: "/a/b/c/d/e/f/g",
			expected: pmpb.PathTemplate{
				Segments: []*pmpb.Segment{
					{Segment: &pmpb.Segment_Literal{Literal: &pmpb.Literal{Value: "a"}}},
					{Segment: &pmpb.Segment_Literal{Literal: &pmpb.Literal{Value: "b"}}},
					{Segment: &pmpb.Segment_Literal{Literal: &pmpb.Literal{Value: "c"}}},
					{Segment: &pmpb.Segment_Literal{Literal: &pmpb.Literal{Value: "d"}}},
					{Segment: &pmpb.Segment_Literal{Literal: &pmpb.Literal{Value: "e"}}},
					{Segment: &pmpb.Segment_Literal{Literal: &pmpb.Literal{Value: "f"}}},
					{Segment: &pmpb.Segment_Literal{Literal: &pmpb.Literal{Value: "g"}}},
				},
			},
		},
		{
			input: "/with/wildcard/*",
			expected: pmpb.PathTemplate{
				Segments: []*pmpb.Segment{
					{Segment: &pmpb.Segment_Literal{Literal: &pmpb.Literal{Value: "with"}}},
					{Segment: &pmpb.Segment_Literal{Literal: &pmpb.Literal{Value: "wildcard"}}},
					{Segment: &pmpb.Segment_Star{}},
				},
			},
		},
		{
			input: "/with/double/wildcard/**",
			expected: pmpb.PathTemplate{
				Segments: []*pmpb.Segment{
					{Segment: &pmpb.Segment_Literal{Literal: &pmpb.Literal{Value: "with"}}},
					{Segment: &pmpb.Segment_Literal{Literal: &pmpb.Literal{Value: "double"}}},
					{Segment: &pmpb.Segment_Literal{Literal: &pmpb.Literal{Value: "wildcard"}}},
					{Segment: &pmpb.Segment_DoubleStar{DoubleStar: &pmpb.DoubleStar{}}},
				},
			},
		},
		{
			input: "/with/double/wildcard/{varame=path/**}",
			expected: pmpb.PathTemplate{
				Segments: []*pmpb.Segment{
					{Segment: &pmpb.Segment_Literal{Literal: &pmpb.Literal{Value: "with"}}},
					{Segment: &pmpb.Segment_Literal{Literal: &pmpb.Literal{Value: "double"}}},
					{Segment: &pmpb.Segment_Literal{Literal: &pmpb.Literal{Value: "wildcard"}}},
					{
						Segment: &pmpb.Segment_Variable{
							Variable: &pmpb.Variable{
								Name: "varame",
								Segments: []*pmpb.Segment{
									{Segment: &pmpb.Segment_Literal{Literal: &pmpb.Literal{Value: "path"}}},
									{Segment: &pmpb.Segment_DoubleStar{DoubleStar: &pmpb.DoubleStar{}}},
								},
							},
						},
					},
				},
			},
		},
		{
			input: "/with/variable/{name}",
			expected: pmpb.PathTemplate{
				Segments: []*pmpb.Segment{
					{Segment: &pmpb.Segment_Literal{Literal: &pmpb.Literal{Value: "with"}}},
					{Segment: &pmpb.Segment_Literal{Literal: &pmpb.Literal{Value: "variable"}}},
					{Segment: &pmpb.Segment_Variable{Variable: &pmpb.Variable{Name: "name"}}},
				},
			},
		},
		{
			input: "/with/variable/{name=/some/other/path}",
			expected: pmpb.PathTemplate{
				Segments: []*pmpb.Segment{
					{Segment: &pmpb.Segment_Literal{Literal: &pmpb.Literal{Value: "with"}}},
					{Segment: &pmpb.Segment_Literal{Literal: &pmpb.Literal{Value: "variable"}}},
					{
						Segment: &pmpb.Segment_Variable{
							Variable: &pmpb.Variable{
								Name: "name",
								Segments: []*pmpb.Segment{
									{Segment: &pmpb.Segment_Literal{Literal: &pmpb.Literal{Value: "some"}}},
									{Segment: &pmpb.Segment_Literal{Literal: &pmpb.Literal{Value: "other"}}},
									{Segment: &pmpb.Segment_Literal{Literal: &pmpb.Literal{Value: "path"}}},
								},
							},
						},
					},
				},
			},
		},
		{
			input: "/with/variable/{name=/some/other/path}/and/more",
			expected: pmpb.PathTemplate{
				Segments: []*pmpb.Segment{
					{Segment: &pmpb.Segment_Literal{Literal: &pmpb.Literal{Value: "with"}}},
					{Segment: &pmpb.Segment_Literal{Literal: &pmpb.Literal{Value: "variable"}}},
					{
						Segment: &pmpb.Segment_Variable{
							Variable: &pmpb.Variable{
								Name: "name",
								Segments: []*pmpb.Segment{
									{Segment: &pmpb.Segment_Literal{Literal: &pmpb.Literal{Value: "some"}}},
									{Segment: &pmpb.Segment_Literal{Literal: &pmpb.Literal{Value: "other"}}},
									{Segment: &pmpb.Segment_Literal{Literal: &pmpb.Literal{Value: "path"}}},
								},
							},
						},
					},
					{Segment: &pmpb.Segment_Literal{Literal: &pmpb.Literal{Value: "and"}}},
					{Segment: &pmpb.Segment_Literal{Literal: &pmpb.Literal{Value: "more"}}},
				},
			},
		},
	}

	for i := range tests {
		t.Run(tests[i].input, func(t *testing.T) {
			result, err := parse.ParseTemplate(tests[i].input)
			require.NoError(t, err, "Parse should not return an error")
			diff := cmp.Diff(&tests[i].expected, result, protocmp.Transform())
			if diff != "" {
				t.Errorf("Parse result mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseError(t *testing.T) {
	tests := []struct {
		input string
		err   error
	}{
		{
			input: "/with/variable/{name",
			err:   parse.ErrUnexpectedEndOfInput,
		},
		{
			input: "/with/variable/{name=/some/other/path",
			err:   parse.ErrUnexpectedEndOfInput,
		},
		{
			input: "/with/variable/**/after",
			err:   parse.ErrUnexpectedDoubleStar,
		},
		{
			input: "/with/variable/{name=/some/other/**}/and/more",
			err:   parse.ErrUnexpectedDoubleStar,
		},
		{
			input: "/with/variable/{name=/some/other/**/path}/and/more",
			err:   parse.ErrUnexpectedDoubleStar,
		},
		{
			input: "/with/double/wildcard/**/**/",
			err:   parse.ErrUnexpectedDoubleStar,
		},
		{
			input: "/with/sub/variable/{var=/sub/{variable=value}}",
			err:   parse.ErrSubVariable,
		},
	}

	for i := range tests {
		t.Run(tests[i].input, func(t *testing.T) {
			_, err := parse.ParseTemplate(tests[i].input)
			require.Error(t, err, "Parse should return an error")
			require.ErrorIs(t, err, tests[i].err, "Expected error should match")
		})
	}
}

func BenchmarkParse(b *testing.B) {
	input := "/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t/u/v/w/x/y/z"
	for b.Loop() {
		_, err := parse.ParseTemplate(input)
		if err != nil {
			b.Fatalf("Parse failed: %v", err)
		}
	}
}

func BenchmarkParseWithWildcards(b *testing.B) {
	input := "/with/wildcard/*/and/double/wildcard/and/variable/{name=**}"
	for b.Loop() {
		_, err := parse.ParseTemplate(input)
		if err != nil {
			b.Fatalf("Parse failed: %v", err)
		}
	}
}
