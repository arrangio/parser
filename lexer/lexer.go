package lexer

type Lexer struct {
	input        string
	position     int  // points to the currect `ch`
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
				l.col = 1
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
