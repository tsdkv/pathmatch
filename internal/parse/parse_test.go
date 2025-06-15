package parse_test

import (
	"testing"

	"github.com/stretchr/testify/require"

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
