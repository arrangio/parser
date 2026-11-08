package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Col     int
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	IDENT = "IDENT" // desk, room, axis, x, y, min
	INT   = "INT"   // 100, 5000, 0

	HASH     = "#"  // for tags: #furniture
	AT       = "@"  // for positions: @ [0, 0, 0]
	DOT      = "."  // for faces: seat.front
	ASSIGN   = "="  // for parameters: min=120
	ARROW    = "->" // for directions: front -> [0, 1, 0]
	MINUS    = "-"  // for negative values
	ASTERISK = "*"  // match any
	COMMA    = ","  // divide things

	LBRACKET = "["
	RBRACKET = "]"
	LBRACE   = "{"
	RBRACE   = "}"

	// keywords
	SCENE      = "scene"
	BOUNDS     = "bounds"
	TO         = "to"
	ZONE       = "zone"
	ITEM       = "item"
	SPAWN      = "spawn"
	RULE       = "rule"
	STATIC     = "static"
	CAN_ROTATE = "can_rotate"
	ID         = "id"
)

// distinguish keywords from identifiers
var keywords = map[string]TokenType{
	"scene":      SCENE,
	"bounds":     BOUNDS,
	"to":         TO,
	"zone":       ZONE,
	"item":       ITEM,
	"spawn":      SPAWN,
	"rule":       RULE,
	"static":     STATIC,
	"can_rotate": CAN_ROTATE,
	"id":         ID,
}

// `LookupIndent` checks whether the given identifier is a keyword or not
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
