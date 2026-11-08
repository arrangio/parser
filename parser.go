package parser

import (
	"fmt"

	"github.com/arrangio/parser/ast"
	"github.com/arrangio/parser/lexer"
	"github.com/arrangio/parser/token"
)

type Parser struct {
	l      *lexer.Lexer
	tok    token.Token
	peek   token.Token
	errors []string
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}
	p.next()
	p.next()
	return p
}

func (p *Parser) next() {
	p.tok = p.peek
	p.peek = p.l.NextToken()
}

func (p *Parser) is(t token.TokenType) bool     { return p.tok.Type == t }
func (p *Parser) peekIs(t token.TokenType) bool { return p.peek.Type == t }

func (p *Parser) match(t token.TokenType) bool {
	if p.is(t) {
		p.next()
		return true
	}
	return false
}

func (p *Parser) expect(t token.TokenType) string {
	if p.is(t) {
		lit := p.tok.Literal
		p.next()
		return lit
	}
	p.errorf("expected %s, got %s (%q)", t, p.tok.Type, p.tok.Literal)
	return ""
}

func (p *Parser) errorf(format string, args ...any) {
	p.errors = append(p.errors, fmt.Sprintf("%d:%d: %s", p.tok.Line, p.tok.Col, fmt.Sprintf(format, args...)))
}

func (p *Parser) Errors() []string { return p.errors }

func (p *Parser) isAtTopLevel() bool {
	switch p.tok.Type {
	case token.SCENE, token.ZONE, token.ITEM, token.SPAWN, token.RULE, token.RULE_DEF, token.EOF:
		return true
	}
	return false
}

func (p *Parser) isSelectorStart() bool {
	return p.is(token.ASTERISK) || p.is(token.HASH) || p.is(token.IDENT)
}

func (p *Parser) synchronize() {
	p.next()
	for !p.is(token.EOF) && !p.isAtTopLevel() {
		p.next()
	}
}

func (p *Parser) ParseProgram() (*ast.Program, error) {
	prog := &ast.Program{}

	for !p.is(token.EOF) {
		prevErr := len(p.errors)

		switch p.tok.Type {
		case token.SCENE:
			if s := p.parseScene(); s != nil {
				if prog.Scene != nil {
					p.errorf("duplicate scene declaration")
				} else {
					prog.Scene = s
				}
			}
		case token.ZONE:
			if z := p.parseZone(); z != nil {
				prog.Zones = append(prog.Zones, z)
			}
		case token.ITEM:
			if it := p.parseItem(); it != nil {
				prog.Items = append(prog.Items, it)
			}
		case token.SPAWN:
			if sp := p.parseSpawn(); sp != nil {
				prog.Spawns = append(prog.Spawns, sp)
			}
		case token.RULE_DEF:
			if rd := p.parseRuleDef(); rd != nil {
				prog.RuleDefs = append(prog.RuleDefs, rd)
			}
		case token.RULE:
			if p.peek.Type == token.NO_COLLISION || isLegacyRuleName(p.peek.Literal) {
				if r := p.parseRule(); r != nil {
					prog.Rules = append(prog.Rules, r)
				}
			} else if inv := p.parseRuleInvocation(); inv != nil {
				prog.Invocations = append(prog.Invocations, inv)
			}
		default:
			if p.isExpressionStart() {
				if c := p.parseConstraint(); c != nil {
					prog.Constraints = append(prog.Constraints, c)
				}
			} else {
				p.errorf("unexpected token at root: %s (%q)", p.tok.Type, p.tok.Literal)
				p.next()
			}
		}

		if len(p.errors) > prevErr {
			p.synchronize()
		}
	}

	if len(p.errors) > 0 {
		return prog, fmt.Errorf("parsing failed with %d error(s)", len(p.errors))
	}
	return prog, nil
}
