package pathmatch_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"github.com/tsdkv/pathmatch"
	"github.com/tsdkv/pathmatch/pathmatchpb"

	"google.golang.org/protobuf/testing/protocmp"
)

func TestLexerSimple(t *testing.T) {
	lex := pathmatch.NewLexer("/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t/u/v/w/x/y/z")
	if lex == nil {
		t.Fatal("lexer should not be nil")
	}

	for i := range 25 {
		if !lex.Match(pathmatch.TokenSlash) {
			t.Fatalf("expected token type %v, got %v", pathmatch.TokenSlash, lex.Peek().Type)
		}
		require.Equal(t, string(byte('a'+i)), lex.Peek().Value, "expected token value to match the expected character")
		if !lex.Match(pathmatch.TokenLiteral) {
			t.Fatalf("expected token type %v, got %v", pathmatch.TokenLiteral, lex.Peek().Type)
		}
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		input    string
		expected pathmatchpb.PathTemplate
	}{
		{
			input: "/a/b/c/d/e/f/g",
			expected: pathmatchpb.PathTemplate{
				Segments: []*pathmatchpb.Segment{
					{Segment: &pathmatchpb.Segment_Literal{Literal: &pathmatchpb.Literal{Value: "a"}}},
					{Segment: &pathmatchpb.Segment_Literal{Literal: &pathmatchpb.Literal{Value: "b"}}},
					{Segment: &pathmatchpb.Segment_Literal{Literal: &pathmatchpb.Literal{Value: "c"}}},
					{Segment: &pathmatchpb.Segment_Literal{Literal: &pathmatchpb.Literal{Value: "d"}}},
					{Segment: &pathmatchpb.Segment_Literal{Literal: &pathmatchpb.Literal{Value: "e"}}},
					{Segment: &pathmatchpb.Segment_Literal{Literal: &pathmatchpb.Literal{Value: "f"}}},
					{Segment: &pathmatchpb.Segment_Literal{Literal: &pathmatchpb.Literal{Value: "g"}}},
				},
			},
		},
		{
			input: "/with/wildcard/*",
			expected: pathmatchpb.PathTemplate{
				Segments: []*pathmatchpb.Segment{
					{Segment: &pathmatchpb.Segment_Literal{Literal: &pathmatchpb.Literal{Value: "with"}}},
					{Segment: &pathmatchpb.Segment_Literal{Literal: &pathmatchpb.Literal{Value: "wildcard"}}},
					{Segment: &pathmatchpb.Segment_Star{}},
				},
			},
		},
		{
			input: "/with/double/wildcard/**",
			expected: pathmatchpb.PathTemplate{
				Segments: []*pathmatchpb.Segment{
					{Segment: &pathmatchpb.Segment_Literal{Literal: &pathmatchpb.Literal{Value: "with"}}},
					{Segment: &pathmatchpb.Segment_Literal{Literal: &pathmatchpb.Literal{Value: "double"}}},
					{Segment: &pathmatchpb.Segment_Literal{Literal: &pathmatchpb.Literal{Value: "wildcard"}}},
					{Segment: &pathmatchpb.Segment_DoubleStar{DoubleStar: &pathmatchpb.DoubleStar{}}},
				},
			},
		},
		{
			input: "/with/double/wildcard/{varame=path/**}",
			expected: pathmatchpb.PathTemplate{
				Segments: []*pathmatchpb.Segment{
					{Segment: &pathmatchpb.Segment_Literal{Literal: &pathmatchpb.Literal{Value: "with"}}},
					{Segment: &pathmatchpb.Segment_Literal{Literal: &pathmatchpb.Literal{Value: "double"}}},
					{Segment: &pathmatchpb.Segment_Literal{Literal: &pathmatchpb.Literal{Value: "wildcard"}}},
					{
						Segment: &pathmatchpb.Segment_Variable{
							Variable: &pathmatchpb.Variable{
								Name: "varame",
								Segments: []*pathmatchpb.Segment{
									{Segment: &pathmatchpb.Segment_Literal{Literal: &pathmatchpb.Literal{Value: "path"}}},
									{Segment: &pathmatchpb.Segment_DoubleStar{DoubleStar: &pathmatchpb.DoubleStar{}}},
								},
							},
						},
					},
				},
			},
		},
		{
			input: "/with/variable/{name}",
			expected: pathmatchpb.PathTemplate{
				Segments: []*pathmatchpb.Segment{
					{Segment: &pathmatchpb.Segment_Literal{Literal: &pathmatchpb.Literal{Value: "with"}}},
					{Segment: &pathmatchpb.Segment_Literal{Literal: &pathmatchpb.Literal{Value: "variable"}}},
					{Segment: &pathmatchpb.Segment_Variable{Variable: &pathmatchpb.Variable{Name: "name"}}},
				},
			},
		},
		{
			input: "/with/variable/{name=/some/other/path}",
			expected: pathmatchpb.PathTemplate{
				Segments: []*pathmatchpb.Segment{
					{Segment: &pathmatchpb.Segment_Literal{Literal: &pathmatchpb.Literal{Value: "with"}}},
					{Segment: &pathmatchpb.Segment_Literal{Literal: &pathmatchpb.Literal{Value: "variable"}}},
					{
						Segment: &pathmatchpb.Segment_Variable{
							Variable: &pathmatchpb.Variable{
								Name: "name",
								Segments: []*pathmatchpb.Segment{
									{Segment: &pathmatchpb.Segment_Literal{Literal: &pathmatchpb.Literal{Value: "some"}}},
									{Segment: &pathmatchpb.Segment_Literal{Literal: &pathmatchpb.Literal{Value: "other"}}},
									{Segment: &pathmatchpb.Segment_Literal{Literal: &pathmatchpb.Literal{Value: "path"}}},
								},
							},
						},
					},
				},
			},
		},
		{
			input: "/with/variable/{name=/some/other/path}/and/more",
			expected: pathmatchpb.PathTemplate{
				Segments: []*pathmatchpb.Segment{
					{Segment: &pathmatchpb.Segment_Literal{Literal: &pathmatchpb.Literal{Value: "with"}}},
					{Segment: &pathmatchpb.Segment_Literal{Literal: &pathmatchpb.Literal{Value: "variable"}}},
					{
						Segment: &pathmatchpb.Segment_Variable{
							Variable: &pathmatchpb.Variable{
								Name: "name",
								Segments: []*pathmatchpb.Segment{
									{Segment: &pathmatchpb.Segment_Literal{Literal: &pathmatchpb.Literal{Value: "some"}}},
									{Segment: &pathmatchpb.Segment_Literal{Literal: &pathmatchpb.Literal{Value: "other"}}},
									{Segment: &pathmatchpb.Segment_Literal{Literal: &pathmatchpb.Literal{Value: "path"}}},
								},
							},
						},
					},
					{Segment: &pathmatchpb.Segment_Literal{Literal: &pathmatchpb.Literal{Value: "and"}}},
					{Segment: &pathmatchpb.Segment_Literal{Literal: &pathmatchpb.Literal{Value: "more"}}},
				},
			},
		},
	}

	for i := range tests {
		t.Run(tests[i].input, func(t *testing.T) {
			result, err := pathmatch.ParseTemplate(tests[i].input)
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
			err:   pathmatch.ErrUnexpectedEndOfInput,
		},
		{
			input: "/with/variable/{name=/some/other/path",
			err:   pathmatch.ErrUnexpectedEndOfInput,
		},
		{
			input: "/with/variable/**/after",
			err:   pathmatch.ErrUnexpectedDoubleStar,
		},
		{
			input: "/with/variable/{name=/some/other/**}/and/more",
			err:   pathmatch.ErrUnexpectedDoubleStar,
		},
		{
			input: "/with/variable/{name=/some/other/**/path}/and/more",
			err:   pathmatch.ErrUnexpectedDoubleStar,
		},
		{
			input: "/with/double/wildcard/**/**/",
			err:   pathmatch.ErrUnexpectedDoubleStar,
		},
	}

	for i := range tests {
		t.Run(tests[i].input, func(t *testing.T) {
			_, err := pathmatch.ParseTemplate(tests[i].input)
			require.Error(t, err, "Parse should return an error")
			require.ErrorIs(t, err, tests[i].err, "Expected error should match")
		})
	}
}

func BenchmarkParse(b *testing.B) {
	input := "/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t/u/v/w/x/y/z"
	for b.Loop() {
		_, err := pathmatch.ParseTemplate(input)
		if err != nil {
			b.Fatalf("Parse failed: %v", err)
		}
	}
}

func BenchmarkParseWithWildcards(b *testing.B) {
	input := "/with/wildcard/*/and/double/wildcard/and/variable/{name=**}"
	for b.Loop() {
		_, err := pathmatch.ParseTemplate(input)
		if err != nil {
			b.Fatalf("Parse failed: %v", err)
		}
	}
}
