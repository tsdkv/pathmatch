package parse

import "strings"

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
