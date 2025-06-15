package pathmatch

import (
	"github.com/tsdkv/pathmatch/internal/parse"
	pmpb "github.com/tsdkv/pathmatch/pathmatchpb"
)

// Template = "/" Segments;
// Segments = Segment { "/" Segment } ;
// Segment  = "*" | "**" | LITERAL | Variable ;
// Variable = "{" FieldPath [ "=" Segments ] "}" ;
// FieldPath = IDENT { "." IDENT } ; // unsupported in this version
func ParseTemplate(s string) (*pmpb.PathTemplate, error) {
	return parse.ParseTemplate(s)
}
