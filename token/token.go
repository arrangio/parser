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
	MINUS    = "-"  // for negative values and subtraction
	PLUS     = "+"  // addition
	ASTERISK = "*"  // multiplication or match any
	SLASH    = "/"  // division
	COMMA    = ","  // divide things

	// Comparison operators
	EQ     = "=="
	NOT_EQ = "!="
	LT     = "<"
	LTE    = "<="
	GT     = ">"
	GTE    = ">="
	BANG   = "!"
	DOTDOT = ".."

	LBRACKET = "["
	RBRACKET = "]"
	LBRACE   = "{"
	RBRACE   = "}"
	LPAREN   = "("
	RPAREN   = ")"

	// keywords
	SCENE        = "scene"
	BOUNDS       = "bounds"
	TO           = "to"
	ZONE         = "zone"
	ITEM         = "item"
	SPAWN        = "spawn"
	RULE         = "rule"
	RULE_DEF     = "rule_def"
	STATIC       = "static"
	CAN_ROTATE   = "can_rotate"
	ID           = "id"
	DIST         = "dist"
	IN           = "in"
	NOT          = "not"
	NO_COLLISION = "no_collision"
)

// LookupIdent checks whether the given identifier is a keyword or not
func LookupIdent(ident string) TokenType {
	switch ident {
	case SCENE, BOUNDS, TO, ZONE, ITEM, SPAWN, RULE, RULE_DEF,
		STATIC, CAN_ROTATE, ID, DIST, IN, NOT, NO_COLLISION:
		return TokenType(ident)
	default:
		return IDENT
	}
}
