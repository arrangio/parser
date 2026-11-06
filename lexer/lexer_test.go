package lexer_test

import (
	"testing"

	"github.com/arrangio/parser/lexer"
	"github.com/arrangio/parser/token"
)

func TestLexer(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []token.Token
	}{
		{
			name:  "empty input",
			input: "",
			expected: []token.Token{
				{Type: token.EOF, Literal: "", Line: 1, Col: 1},
			},
		},
		{
			name:  "single character symbols",
			input: "[ ] { } # @ . = * ,",
			expected: []token.Token{
				{Type: token.LBRACKET, Literal: "[", Line: 1, Col: 1},
				{Type: token.RBRACKET, Literal: "]", Line: 1, Col: 3},
				{Type: token.LBRACE, Literal: "{", Line: 1, Col: 5},
				{Type: token.RBRACE, Literal: "}", Line: 1, Col: 7},
				{Type: token.HASH, Literal: "#", Line: 1, Col: 9},
				{Type: token.AT, Literal: "@", Line: 1, Col: 11},
				{Type: token.DOT, Literal: ".", Line: 1, Col: 13},
				{Type: token.ASSIGN, Literal: "=", Line: 1, Col: 15},
				{Type: token.ASTERISK, Literal: "*", Line: 1, Col: 17},
				{Type: token.COMMA, Literal: ",", Line: 1, Col: 19},
				{Type: token.EOF, Literal: "", Line: 1, Col: 20},
			},
		},
		{
			name:  "arrow vs minus disambiguation",
			input: "-> - ->",
			expected: []token.Token{
				{Type: token.ARROW, Literal: "->", Line: 1, Col: 1},
				{Type: token.MINUS, Literal: "-", Line: 1, Col: 4},
				{Type: token.ARROW, Literal: "->", Line: 1, Col: 6},
				{Type: token.EOF, Literal: "", Line: 1, Col: 8},
			},
		},
		{
			name:  "keywords and identifiers",
			input: "zone item spawn rule static can_rotate desk chair_1",
			expected: []token.Token{
				{Type: token.ZONE, Literal: "zone", Line: 1, Col: 1},
				{Type: token.ITEM, Literal: "item", Line: 1, Col: 6},
				{Type: token.SPAWN, Literal: "spawn", Line: 1, Col: 11},
				{Type: token.RULE, Literal: "rule", Line: 1, Col: 17},
				{Type: token.STATIC, Literal: "static", Line: 1, Col: 22},
				{Type: token.CAN_ROTATE, Literal: "can_rotate", Line: 1, Col: 29},
				{Type: token.IDENT, Literal: "desk", Line: 1, Col: 40},
				{Type: token.IDENT, Literal: "chair_1", Line: 1, Col: 45},
				{Type: token.EOF, Literal: "", Line: 1, Col: 52},
			},
		},
		{
			name:  "numbers",
			input: "0 42 1000",
			expected: []token.Token{
				{Type: token.INT, Literal: "0", Line: 1, Col: 1},
				{Type: token.INT, Literal: "42", Line: 1, Col: 3},
				{Type: token.INT, Literal: "1000", Line: 1, Col: 6},
				{Type: token.EOF, Literal: "", Line: 1, Col: 10},
			},
		},
		{
			name: "comments and newlines",
			input: `// header comment
zone // inline comment
item`,
			expected: []token.Token{
				{Type: token.ZONE, Literal: "zone", Line: 2, Col: 1},
				{Type: token.ITEM, Literal: "item", Line: 3, Col: 1},
				{Type: token.EOF, Literal: "", Line: 3, Col: 5},
			},
		},
		{
			name:  "vector with negative numbers",
			input: "front -> [0, -1, 0]",
			expected: []token.Token{
				{Type: token.IDENT, Literal: "front", Line: 1, Col: 1},
				{Type: token.ARROW, Literal: "->", Line: 1, Col: 7},
				{Type: token.LBRACKET, Literal: "[", Line: 1, Col: 10},
				{Type: token.INT, Literal: "0", Line: 1, Col: 11},
				{Type: token.COMMA, Literal: ",", Line: 1, Col: 12},
				{Type: token.MINUS, Literal: "-", Line: 1, Col: 14},
				{Type: token.INT, Literal: "1", Line: 1, Col: 15},
				{Type: token.COMMA, Literal: ",", Line: 1, Col: 16},
				{Type: token.INT, Literal: "0", Line: 1, Col: 18},
				{Type: token.RBRACKET, Literal: "]", Line: 1, Col: 19},
				{Type: token.EOF, Literal: "", Line: 1, Col: 20},
			},
		},
		{
			name:  "rule declaration with params",
			input: "rule separation chair table min=10",
			expected: []token.Token{
				{Type: token.RULE, Literal: "rule", Line: 1, Col: 1},
				{Type: token.IDENT, Literal: "separation", Line: 1, Col: 6},
				{Type: token.IDENT, Literal: "chair", Line: 1, Col: 17},
				{Type: token.IDENT, Literal: "table", Line: 1, Col: 23},
				{Type: token.IDENT, Literal: "min", Line: 1, Col: 29},
				{Type: token.ASSIGN, Literal: "=", Line: 1, Col: 32},
				{Type: token.INT, Literal: "10", Line: 1, Col: 33},
				{Type: token.EOF, Literal: "", Line: 1, Col: 35},
			},
		},
		{
			name:  "match all wildcard rule",
			input: "rule no_collision * *",
			expected: []token.Token{
				{Type: token.RULE, Literal: "rule", Line: 1, Col: 1},
				{Type: token.IDENT, Literal: "no_collision", Line: 1, Col: 6},
				{Type: token.ASTERISK, Literal: "*", Line: 1, Col: 19},
				{Type: token.ASTERISK, Literal: "*", Line: 1, Col: 21},
				{Type: token.EOF, Literal: "", Line: 1, Col: 22},
			},
		},
		{
			name:  "illegal characters",
			input: "? $",
			expected: []token.Token{
				{Type: token.ILLEGAL, Literal: "?", Line: 1, Col: 1},
				{Type: token.ILLEGAL, Literal: "$", Line: 1, Col: 3},
				{Type: token.EOF, Literal: "", Line: 1, Col: 4},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			l := lexer.New(tc.input)

			for i, exp := range tc.expected {
				tok := l.NextToken()

				if tok.Type != exp.Type {
					t.Errorf("[%d] type mismatch: expected %q, got %q", i, exp.Type, tok.Type)
				}
				if tok.Literal != exp.Literal {
					t.Errorf("[%d] literal mismatch: expected %q, got %q", i, exp.Literal, tok.Literal)
				}
				if tok.Line != exp.Line {
					t.Errorf("[%d] line mismatch: expected %d, got %d", i, exp.Line, tok.Line)
				}
				if tok.Col != exp.Col {
					t.Errorf("[%d] col mismatch: expected %d, got %d", i, exp.Col, tok.Col)
				}
			}
		})
	}
}
