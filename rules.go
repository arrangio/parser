package parser

import (
	"github.com/arrangio/parser/ast"
	"github.com/arrangio/parser/token"
)

func isLegacyRuleName(name string) bool {
	switch name {
	case "no_collision", "separation", "proximity", "orientation",
		"containment", "clearance", "alignment", "zone_exclusion", "axis_restriction":
		return true
	}
	return false
}

func (p *Parser) parseRule() *ast.RuleDecl {
	tok := p.tok
	p.next() // skip 'rule'

	kind := p.tok.Literal
	if p.is(token.NO_COLLISION) || p.is(token.IDENT) {
		p.next()
	} else {
		kind = p.expect(token.IDENT)
	}

	rule := &ast.RuleDecl{Token: tok, Kind: kind, Params: make(map[string]any)}

	if p.isSelectorStart() {
		rule.Target = p.parseSelector()
		if p.match(token.DOT) {
			rule.Marker = p.expect(token.IDENT)
		}
	}

	for !p.isAtTopLevel() {
		switch {
		case p.match(token.ARROW) || p.is(token.LBRACKET):
			rule.Vectors = append(rule.Vectors, p.parseVec3())
		case p.is(token.IDENT) && p.peekIs(token.ASSIGN):
			k := p.expect(token.IDENT)
			p.expect(token.ASSIGN)
			rule.Params[k] = p.parseParamValue()
		case p.match(token.TO), rule.Secondary == nil && p.isSelectorStart() && p.tok.Literal != "closest":
			sec := p.parseSelector()
			rule.Secondary = &sec
		case p.is(token.IDENT):
			rule.Flags = append(rule.Flags, p.expect(token.IDENT))
		default:
			p.errorf("unexpected token in rule: %s (%q)", p.tok.Type, p.tok.Literal)
			p.next()
		}
	}
	return rule
}

func (p *Parser) parseRuleDef() *ast.RuleDefDecl {
	tok := p.tok
	p.next() // 'rule_def'
	name := p.expect(token.IDENT)
	p.expect(token.LPAREN)

	var params []ast.RuleParam
	for !p.is(token.RPAREN) && !p.is(token.EOF) {
		pName := p.expect(token.IDENT)
		var defVal ast.Expr
		if p.match(token.ASSIGN) {
			defVal = p.parseExpression(LOWEST)
		}
		params = append(params, ast.RuleParam{Name: pName, DefaultValue: defVal})
		p.match(token.COMMA)
	}
	p.expect(token.RPAREN)
	p.expect(token.LBRACE)

	var body []ast.Node
	for !p.is(token.RBRACE) && !p.is(token.EOF) {
		switch {
		case p.is(token.RULE):
			if p.peek.Type == token.NO_COLLISION || isLegacyRuleName(p.peek.Literal) {
				body = append(body, p.parseRule())
			} else {
				body = append(body, p.parseRuleInvocation())
			}
		case p.isExpressionStart():
			if c := p.parseConstraint(); c != nil {
				body = append(body, c)
			}
		default:
			p.errorf("unexpected token in rule_def body: %s (%q)", p.tok.Type, p.tok.Literal)
			p.next()
		}
	}
	p.expect(token.RBRACE)
	return &ast.RuleDefDecl{Token: tok, Name: name, Params: params, Body: body}
}

func (p *Parser) parseRuleInvocation() *ast.RuleInvocationStmt {
	tok := p.tok
	p.next() // 'rule'
	inv := &ast.RuleInvocationStmt{Token: tok, Name: p.expect(token.IDENT), Params: make(map[string]ast.Expr)}

	parseArg := func() {
		if p.is(token.IDENT) && p.peekIs(token.ASSIGN) {
			k := p.expect(token.IDENT)
			p.expect(token.ASSIGN)
			inv.Params[k] = p.parseExpression(LOWEST)
		} else if p.isExpressionStart() {
			inv.Args = append(inv.Args, p.parseExpression(LOWEST))
		}
	}

	if p.match(token.LPAREN) {
		for !p.is(token.RPAREN) && !p.is(token.EOF) {
			parseArg()
			p.match(token.COMMA)
		}
		p.expect(token.RPAREN)
		return inv
	}

	for !p.isAtTopLevel() && (p.isExpressionStart() || (p.is(token.IDENT) && p.peekIs(token.ASSIGN))) {
		parseArg()
	}
	return inv
}
