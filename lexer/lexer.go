package lexer

import "github.com/arrangio/parser/token"

type Lexer struct {
	input        string
	position     int  // points to the current `ch`
	readPosition int  // next reading position
	ch           byte // current char under examination
	line         int  // line number, starting from 1
	col          int  // column number, starting from 1
}

func New(input string) *Lexer {
	l := &Lexer{
		input: input,
		line:  1,
		col:   0,
	}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // EOF
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	l.col++
}

func (l *Lexer) skipWhitespacesAndComments() {
	for {
		// skip whitespaces
		for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
			if l.ch == '\n' {
				l.line++
				l.col = 0
			}
			l.readChar()
		}

		// skip comments
		if l.ch == '/' && l.peekChar() == '/' {
			for l.ch != '\n' && l.ch != 0 {
				l.readChar()
			}
			continue
		}

		break
	}
}

// `peekChar` peeks the next character without advancing the lexer
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) NextToken() token.Token {
	l.skipWhitespacesAndComments()

	tokLine := l.line
	tokCol := l.col

	var tok token.Token

	switch l.ch {
	case '[':
		tok = newToken(token.LBRACKET, l.ch, tokLine, tokCol)
	case ']':
		tok = newToken(token.RBRACKET, l.ch, tokLine, tokCol)
	case '{':
		tok = newToken(token.LBRACE, l.ch, tokLine, tokCol)
	case '}':
		tok = newToken(token.RBRACE, l.ch, tokLine, tokCol)
	case '@':
		tok = newToken(token.AT, l.ch, tokLine, tokCol)
	case '#':
		tok = newToken(token.HASH, l.ch, tokLine, tokCol)
	case '.':
		tok = newToken(token.DOT, l.ch, tokLine, tokCol)
	case ',':
		tok = newToken(token.COMMA, l.ch, tokLine, tokCol)
	case '*':
		tok = newToken(token.ASTERISK, l.ch, tokLine, tokCol)
	case '=':
		tok = newToken(token.ASSIGN, l.ch, tokLine, tokCol)
	case '-':
		if l.peekChar() == '>' {
			l.readChar()
			l.readChar()
			return token.Token{
				Type:    token.ARROW,
				Literal: "->",
				Line:    tokLine,
				Col:     tokCol,
			}
		}
		tok = newToken(token.MINUS, l.ch, tokLine, tokCol)
	case 0:
		return token.Token{
			Type:    token.EOF,
			Literal: "",
			Line:    tokLine,
			Col:     tokCol,
		}
	default:
		if isLetter(l.ch) {
			ident := l.readIdentifier()
			return token.Token{
				Type:    token.LookupIdent(ident),
				Literal: ident,
				Line:    tokLine,
				Col:     tokCol,
			}
		} else if isDigit(l.ch) {
			num := l.readNumber()
			return token.Token{
				Type:    token.INT,
				Literal: num,
				Line:    tokLine,
				Col:     tokCol,
			}
		} else {
			tok = newToken(token.ILLEGAL, l.ch, tokLine, tokCol)
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) readIdentifier() string {
	start := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[start:l.position]
}

func (l *Lexer) readNumber() string {
	start := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[start:l.position]
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func newToken(tokenType token.TokenType, ch byte, line, col int) token.Token {
	return token.Token{
		Type:    tokenType,
		Literal: string(ch),
		Line:    line,
		Col:     col,
	}
}
