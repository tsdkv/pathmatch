package pathmatch

import (
	"errors"
	"fmt"
	"strings"

	"github.com/tsdkv/pathmatch/pathmatchpb"
)

// Template = "/" Segments;
// Segments = Segment { "/" Segment } ;
// Segment  = "*" | "**" | LITERAL | Variable ;
// Variable = "{" FieldPath [ "=" Segments ] "}" ;
// FieldPath = IDENT { "." IDENT } ; // unsupported in this version

type TokenType int

const (
	TokenUnknown    TokenType = iota
	TokenSlash                // '/'
	TokenStar                 // '*'
	TokenDoubleStar           // '**'
	TokenLiteral              // LITERAL
	TokenLeftBrace            // '{'
	TokenRightBrace           // '}'
	TokenEq                   // '='
	TokenEOF
)

var tokenTypeNames = map[TokenType]string{
	TokenUnknown:    "Unknown",
	TokenSlash:      "/",
	TokenStar:       "*",
	TokenDoubleStar: "**",
	TokenLiteral:    "LITERAL",
	TokenLeftBrace:  "{",
	TokenRightBrace: "}",
	TokenEq:         "=",
	TokenEOF:        "TokenEOF",
}

type Token struct {
	Type  TokenType
	Value string
}

func (t Token) String() string {
	if name, ok := tokenTypeNames[t.Type]; ok {
		if t.Value != "" {
			return fmt.Sprintf("%s(%s)", name, t.Value)
		}
		return name
	}
	return fmt.Sprintf("TokenUnknown(%d)", t.Type)
}

type lexer struct {
	input          string
	curr           Token
	prev           Token
	pos            int
	meetDoubleStar bool // Indicates if the lexer has encountered a '**' token
}

func NewLexer(s string) *lexer {
	lex := &lexer{input: s, pos: 0, meetDoubleStar: false}
	if len(lex.input) == 0 {
		lex.curr = Token{Type: TokenEOF}
	} else {
		lex.curr = lex.nextToken()
	}

	return lex
}

func (l *lexer) advance() {
	if l.pos < len(l.input) {
		l.pos++
	}
}

// Returns current token
func (l *lexer) Peek() Token {
	return l.curr
}

// Returns previous token
func (l *lexer) Prev() Token {
	return l.prev
}

func (l *lexer) Rest() string {
	if l.pos >= len(l.input) {
		return ""
	}
	return string(l.input[l.pos:])
}

// If current token matches the given type,
// it advances to the next token and returns true.
func (l *lexer) Match(tok TokenType) bool {
	if l.curr.Type != tok {
		return false
	}
	if tok == TokenDoubleStar {
		l.meetDoubleStar = true // Mark that we have encountered a '**' token
	}
	l.prev = l.curr
	l.curr = l.nextToken()
	return true
}

func (l *lexer) MeetDoubleStar() bool {
	return l.meetDoubleStar
}

func (l *lexer) nextToken() Token {
	if l.pos >= len(l.input) {
		return Token{Type: TokenEOF}
	}

	ch := l.input[l.pos]
	switch ch {
	case '/':
		l.advance()
		return Token{Type: TokenSlash}
	case '*':
		if l.pos+1 < len(l.input) && l.input[l.pos+1] == '*' {
			l.advance()
			l.advance()
			return Token{Type: TokenDoubleStar}
		}
		l.advance()
		return Token{Type: TokenStar}
	case '{':
		l.advance()
		return Token{Type: TokenLeftBrace}
	case '}':
		l.advance()
		return Token{Type: TokenRightBrace}
	case '=':
		l.advance()
		return Token{Type: TokenEq}
	default:
		start := l.pos
		end := strings.IndexAny(l.input[l.pos:], "/*{}=")
		if end == -1 {
			end = len(l.input)
		} else {
			end += l.pos
		}
		value := l.input[start:end]
		l.pos = end
		return Token{Type: TokenLiteral, Value: value}
	}
}

var (
	ErrUnexpectedEndOfInput = errors.New("unexpected end of input")
	ErrUnexpectedDoubleStar = errors.New("unexpected '**' token in the middle of the path")
	ErrUnexpectedToken      = errors.New("unexpected token")
	ErrSubVariable          = errors.New("sub variables are not allowed in thix context")
)

// ParseTemplate parses a path template string and returns a PathMatch object
// or an error if the template is invalid.
func ParseTemplate(s string) (*pathmatchpb.PathTemplate, error) {
	lex := NewLexer(s)
	if lex == nil {
		return nil, fmt.Errorf("failed to create lexer for input: %s", s)
	}

	if !lex.Match(TokenSlash) {
		return nil, fmt.Errorf("expected leading '/', got: %s", lex.Peek())
	}

	return parseSegments(lex)
}

func parseSegments(lex *lexer) (*pathmatchpb.PathTemplate, error) {
	segments := make([]*pathmatchpb.Segment, 0)

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

	return &pathmatchpb.PathTemplate{Segments: segments}, nil
}

// parseSegment parses a single segment of the path template.
// It can be a literal, a wildcard ('*'), a double wildcard ('**'), or a variable.
// If expectVar is true, it expects a variable segment and will parse it accordingly.
// If expectVar is false, it will not parse a variable and will return an error if it encounters one.
func parseSegment(lex *lexer, expectVar bool) (*pathmatchpb.Segment, error) {
	if !lex.MeetDoubleStar() && lex.Match(TokenDoubleStar) {
		return &pathmatchpb.Segment{Segment: &pathmatchpb.Segment_DoubleStar{DoubleStar: &pathmatchpb.DoubleStar{}}}, nil
	}
	// If we encounter '**' in the middle of the path, it's an error
	if lex.MeetDoubleStar() {
		return nil, ErrUnexpectedDoubleStar
	}
	if lex.Match(TokenStar) {
		return &pathmatchpb.Segment{Segment: &pathmatchpb.Segment_Star{}}, nil
	}
	if lex.Match(TokenLiteral) {
		return &pathmatchpb.Segment{
			Segment: &pathmatchpb.Segment_Literal{
				Literal: &pathmatchpb.Literal{
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
		return nil, fmt.Errorf("%w: got %q", ErrSubVariable, seg.Segment.(*pathmatchpb.Segment_Variable).Variable.Name)
	}
	return nil, err
}

func parseVariable(lex *lexer) (*pathmatchpb.Segment, error) {
	if !lex.Match(TokenLeftBrace) {
		return nil, fmt.Errorf("unexpected token: %s", lex.Peek())
	}
	if !lex.Match(TokenLiteral) {
		return nil, fmt.Errorf("expected variable name after '{', got: %s", lex.Peek())
	}
	varName := lex.Prev().Value

	if lex.Match(TokenRightBrace) {
		return &pathmatchpb.Segment{
			Segment: &pathmatchpb.Segment_Variable{
				Variable: &pathmatchpb.Variable{
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

	var segments []*pathmatchpb.Segment
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
	return &pathmatchpb.Segment{
		Segment: &pathmatchpb.Segment_Variable{
			Variable: &pathmatchpb.Variable{
				Name:     varName,
				Segments: segments,
			},
		},
	}, nil
}
