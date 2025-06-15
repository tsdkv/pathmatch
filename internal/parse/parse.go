package parse

import (
	"errors"
	"fmt"

	pmpb "github.com/tsdkv/pathmatch/pathmatchpb"
)

var (
	ErrUnexpectedEndOfInput = errors.New("unexpected end of input")
	ErrUnexpectedDoubleStar = errors.New("unexpected '**' token in the middle of the path")
	ErrUnexpectedToken      = errors.New("unexpected token")
	ErrSubVariable          = errors.New("sub variables are not allowed in thix context")
)

// ParseTemplate parses a path template string and returns a PathMatch object
// or an error if the template is invalid.
func ParseTemplate(s string) (*pmpb.PathTemplate, error) {
	lex := NewLexer(s)
	if lex == nil {
		return nil, fmt.Errorf("failed to create lexer for input: %s", s)
	}

	if !lex.Match(TokenSlash) {
		return nil, fmt.Errorf("expected leading '/', got: %s", lex.Peek())
	}

	return parseSegments(lex)
}

func parseSegments(lex *lexer) (*pmpb.PathTemplate, error) {
	segments := make([]*pmpb.Segment, 0)

	for {
		if lex.Match(TokenEOF) {
			break
		}

		if lex.Match(TokenSlash) {
			continue
		}

		segment, err := parseSegment(lex, true)
		if err != nil {
			return nil, err
		}
		segments = append(segments, segment)
	}

	return &pmpb.PathTemplate{Segments: segments}, nil
}

// parseSegment parses a single segment of the path template.
// It can be a literal, a wildcard ('*'), a double wildcard ('**'), or a variable.
// If expectVar is true, it expects a variable segment and will parse it accordingly.
// If expectVar is false, it will not parse a variable and will return an error if it encounters one.
func parseSegment(lex *lexer, expectVar bool) (*pmpb.Segment, error) {
	if !lex.MeetDoubleStar() && lex.Match(TokenDoubleStar) {
		return &pmpb.Segment{Segment: &pmpb.Segment_DoubleStar{DoubleStar: &pmpb.DoubleStar{}}}, nil
	}
	// If we encounter '**' in the middle of the path, it's an error
	if lex.MeetDoubleStar() {
		return nil, ErrUnexpectedDoubleStar
	}
	if lex.Match(TokenStar) {
		return &pmpb.Segment{Segment: &pmpb.Segment_Star{}}, nil
	}
	if lex.Match(TokenLiteral) {
		return &pmpb.Segment{
			Segment: &pmpb.Segment_Literal{
				Literal: &pmpb.Literal{
					Value: lex.Prev().Value,
				},
			},
		}, nil
	}

	if expectVar {
		return parseVariable(lex)
	}
	// sub variables are not allowed
	seg, err := parseVariable(lex)
	if err == nil {
		return nil, fmt.Errorf("%w: got %q", ErrSubVariable, seg.Segment.(*pmpb.Segment_Variable).Variable.Name)
	}
	return nil, err
}

func parseVariable(lex *lexer) (*pmpb.Segment, error) {
	if !lex.Match(TokenLeftBrace) {
		return nil, fmt.Errorf("unexpected token: %s", lex.Peek())
	}
	if !lex.Match(TokenLiteral) {
		return nil, fmt.Errorf("expected variable name after '{', got: %s", lex.Peek())
	}
	varName := lex.Prev().Value

	if lex.Match(TokenRightBrace) {
		return &pmpb.Segment{
			Segment: &pmpb.Segment_Variable{
				Variable: &pmpb.Variable{
					Name:     varName,
					Segments: nil, // Segments will be filled if there is an '='
				},
			},
		}, nil
	}
	if lex.Match(TokenEOF) {
		return nil, fmt.Errorf("%w: variable '%s' must be closed with '}'", ErrUnexpectedEndOfInput, varName)
	}

	if !lex.Match(TokenEq) {
		return nil, fmt.Errorf("expected '=' or '/' after variable name '%s', got: %s", varName, lex.Peek())
	}

	var segments []*pmpb.Segment
	for !lex.Match(TokenRightBrace) {
		if lex.Match(TokenEOF) {
			return nil, fmt.Errorf("%w: variable '%s'", ErrUnexpectedEndOfInput, varName)
		}

		if lex.Match(TokenSlash) {
			continue
		}

		segment, err := parseSegment(lex, false)
		if err != nil {
			return nil, err
		}
		segments = append(segments, segment)
	}

	if len(segments) == 0 {
		return nil, fmt.Errorf("%w: variable '%s' must have at least one segment after '='", ErrUnexpectedEndOfInput, varName)
	}
	return &pmpb.Segment{
		Segment: &pmpb.Segment_Variable{
			Variable: &pmpb.Variable{
				Name:     varName,
				Segments: segments,
			},
		},
	}, nil
}
