package parser

import (
	"strconv"

	"github.com/arrangio/parser/ast"
	"github.com/arrangio/parser/token"
)

const (
	_ int = iota
	LOWEST
	ARROW_PREC
	COMPARISON
	RANGE_PREC
	SUM
	PRODUCT
	PREFIX
	CALL
	MEMBER
)

func (p *Parser) curPrecedence() int {
	switch p.tok.Type {
	case token.ARROW:
		return ARROW_PREC
	case token.EQ, token.NOT_EQ, token.LT, token.LTE, token.GT, token.GTE, token.IN, token.NOT:
		return COMPARISON
	case token.DOTDOT:
		return RANGE_PREC
	case token.PLUS, token.MINUS:
		return SUM
	case token.ASTERISK, token.SLASH:
		return PRODUCT
	case token.LPAREN:
		return CALL
	case token.DOT:
		return MEMBER
	default:
		return LOWEST
	}
}

func (p *Parser) isExpressionStart() bool {
	switch p.tok.Type {
	case token.IDENT, token.DIST, token.NO_COLLISION, token.ASTERISK,
		token.HASH, token.INT, token.LPAREN, token.LBRACKET, token.MINUS:
		return true
	}
	return false
}

func (p *Parser) parseExpression(precedence int) ast.Expr {
	left := p.parsePrefix()
	if left == nil {
		p.errorf("expected expression, got %s (%q)", p.tok.Type, p.tok.Literal)
		return nil
	}

	for !p.is(token.EOF) && precedence < p.curPrecedence() {
		if p.is(token.RPAREN) || p.is(token.RBRACKET) || p.is(token.RBRACE) || p.is(token.COMMA) {
			break
		}
		if left = p.parseInfix(left); left == nil {
			break
		}
	}
	return left
}

func (p *Parser) parsePrefix() ast.Expr {
	tok := p.tok
	switch tok.Type {
	case token.INT:
		val, _ := strconv.ParseInt(tok.Literal, 10, 64)
		p.next()
		return &ast.NumberExpr{Token: tok, Value: val}
	case token.IDENT, token.DIST:
		p.next()
		return &ast.IdentExpr{Token: tok, Name: tok.Literal}
	case token.ASTERISK:
		p.next()
		return &ast.IdentExpr{Token: tok, Name: "*"}
	case token.HASH:
		p.next()
		return &ast.IdentExpr{Token: tok, Name: "#" + p.expect(token.IDENT)}
	case token.MINUS, token.BANG:
		p.next()
		return &ast.UnaryExpr{Token: tok, Op: tok.Type, Right: p.parseExpression(PREFIX)}
	case token.LPAREN:
		p.next()
		expr := p.parseExpression(LOWEST)
		p.expect(token.RPAREN)
		return expr
	case token.LBRACKET:
		return p.parseVectorExpr()
	default:
		return nil
	}
}

func (p *Parser) parseInfix(left ast.Expr) ast.Expr {
	tok := p.tok
	switch tok.Type {
	case token.PLUS, token.MINUS, token.ASTERISK, token.SLASH,
		token.EQ, token.NOT_EQ, token.LT, token.LTE, token.GT, token.GTE, token.IN, token.ARROW:
		prec := p.curPrecedence()
		p.next()
		return &ast.BinaryExpr{Token: tok, Left: left, Op: tok.Type, Right: p.parseExpression(prec)}
	case token.NOT:
		p.next()
		p.expect(token.IN)
		return &ast.BinaryExpr{Token: tok, Left: left, Op: token.NOT, Right: p.parseExpression(COMPARISON)}
	case token.DOTDOT:
		p.next()
		return &ast.RangeExpr{Token: tok, From: left, To: p.parseExpression(RANGE_PREC)}
	case token.DOT:
		p.next()
		return &ast.MemberExpr{Token: tok, Object: left, Property: p.expect(token.IDENT)}
	case token.LPAREN:
		p.next()
		var args []ast.Expr
		for !p.is(token.RPAREN) && !p.is(token.EOF) {
			if arg := p.parseExpression(LOWEST); arg != nil {
				args = append(args, arg)
			}
			p.match(token.COMMA)
		}
		p.expect(token.RPAREN)
		name := ""
		if id, ok := left.(*ast.IdentExpr); ok {
			name = id.Name
		}
		return &ast.CallExpr{Token: tok, Function: name, Args: args}
	default:
		return nil
	}
}

func (p *Parser) parseVectorExpr() ast.Expr {
	tok := p.tok
	p.next() // '['
	x := p.parseExpression(LOWEST)
	p.match(token.COMMA)
	y := p.parseExpression(LOWEST)
	p.match(token.COMMA)
	z := p.parseExpression(LOWEST)
	p.expect(token.RBRACKET)
	return &ast.VectorExpr{Token: tok, X: x, Y: y, Z: z}
}

func (p *Parser) parseConstraint() *ast.ConstraintStmt {
	tok := p.tok
	if p.match(token.NO_COLLISION) {
		var args []ast.Expr
		for !p.isAtTopLevel() && p.isSelectorStart() {
			sec := p.parseSelector()
			args = append(args, &ast.IdentExpr{Token: sec.Token, Name: sec.String()})
		}
		return &ast.ConstraintStmt{Token: tok, Expr: &ast.CallExpr{Token: tok, Function: "no_collision", Args: args}}
	}

	expr := p.parseExpression(LOWEST)
	if expr == nil {
		return nil
	}
	switch expr.(type) {
	case *ast.BinaryExpr, *ast.CallExpr:
		return &ast.ConstraintStmt{Token: tok, Expr: expr}
	default:
		p.errorf("unexpected token at root: %s (%q)", tok.Type, tok.Literal)
		return nil
	}
}
