package parse

import "fmt"

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
