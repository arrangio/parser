package parser

import (
	"fmt"
	"strconv"

	"github.com/arrangio/parser/ast"
	"github.com/arrangio/parser/lexer"
	"github.com/arrangio/parser/token"
)

type Parser struct {
	l      *lexer.Lexer
	tok    token.Token // current token
	peek   token.Token // lookahead token
	errors []string
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: make([]string, 0),
	}
	// populate tok and peek
	p.next()
	p.next()
	return p
}

func (p *Parser) next() {
	p.tok = p.peek
	p.peek = p.l.NextToken()
}

func (p *Parser) is(t token.TokenType) bool {
	return p.tok.Type == t
}

func (p *Parser) peekIs(t token.TokenType) bool {
	return p.peek.Type == t
}

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
	msg := fmt.Sprintf("%d:%d: ", p.tok.Line, p.tok.Col) + fmt.Sprintf(format, args...)
	p.errors = append(p.errors, msg)
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) isAtTopLevel() bool {
	return p.is(token.SCENE) ||
		p.is(token.ZONE) ||
		p.is(token.ITEM) ||
		p.is(token.SPAWN) ||
		p.is(token.RULE) ||
		p.is(token.EOF)
}

func (p *Parser) isSelectorStart() bool {
	return p.is(token.ASTERISK) || p.is(token.HASH) || p.is(token.IDENT)
}

// synchronize skips tokens until the next statement on error
func (p *Parser) synchronize() {
	p.next()
	for !p.is(token.EOF) && !p.isAtTopLevel() {
		p.next()
	}
}

func (p *Parser) ParseProgram() (*ast.Program, error) {
	prog := &ast.Program{
		Zones:  make([]*ast.ZoneDecl, 0),
		Items:  make([]*ast.ItemDecl, 0),
		Spawns: make([]*ast.SpawnDecl, 0),
		Rules:  make([]*ast.RuleDecl, 0),
	}

	for !p.is(token.EOF) {
		prevErrCount := len(p.errors)

		switch p.tok.Type {
		case token.SCENE:
			scene := p.parseScene()
			if scene != nil {
				if prog.Scene != nil {
					p.errorf("duplicate scene declaration")
				} else {
					prog.Scene = scene
				}
			}
		case token.ZONE:
			zone := p.parseZone()
			if zone != nil {
				prog.Zones = append(prog.Zones, zone)
			}
		case token.ITEM:
			item := p.parseItem()
			if item != nil {
				prog.Items = append(prog.Items, item)
			}
		case token.SPAWN:
			spawn := p.parseSpawn()
			if spawn != nil {
				prog.Spawns = append(prog.Spawns, spawn)
			}
		case token.RULE:
			rule := p.parseRule()
			if rule != nil {
				prog.Rules = append(prog.Rules, rule)
			}
		default:
			p.errorf("unexpected token at root: %s (%q)", p.tok.Type, p.tok.Literal)
			p.next()
		}

		// on error, skip garbage until the next declaration
		if len(p.errors) > prevErrCount {
			p.synchronize()
		}
	}

	if len(p.errors) > 0 {
		return prog, fmt.Errorf("parsing failed with %d error(s)", len(p.errors))
	}
	return prog, nil
}

// parseInt parses an integer (with optional unary minus)
func (p *Parser) parseInt() int64 {
	sign := int64(1)
	if p.match(token.MINUS) {
		sign = -1
	}
	lit := p.expect(token.INT)
	val, _ := strconv.ParseInt(lit, 10, 64)
	return sign * val
}

// parseVec3 parses [x, y, z]
func (p *Parser) parseVec3() ast.Vec3 {
	tok := p.tok
	p.expect(token.LBRACKET)

	x := p.parseInt()
	p.match(token.COMMA) // optional comma

	y := p.parseInt()
	p.match(token.COMMA) // optional comma

	z := p.parseInt()
	p.expect(token.RBRACKET)

	return ast.Vec3{Token: tok, X: x, Y: y, Z: z}
}

// scene [w, h, d] (@ [x, y, z])?
// scene bounds [min_x, min_y, min_z] to [max_x, max_y, max_z]
func (p *Parser) parseScene() *ast.SceneDecl {
	tok := p.tok
	p.next() // skip `scene`

	// scene bounds [min] to [max]
	if p.is(token.BOUNDS) || (p.is(token.IDENT) && p.tok.Literal == "bounds") {
		p.next() // skip 'bounds'
		minVec := p.parseVec3()

		if p.is(token.TO) || (p.is(token.IDENT) && p.tok.Literal == "to") {
			p.next() // skip 'to'
		} else {
			p.errorf("expected 'to' after bounds min vector, got %s (%q)", p.tok.Type, p.tok.Literal)
		}
		maxVec := p.parseVec3()

		sizeVec := ast.Vec3{
			Token: minVec.Token,
			X:     maxVec.X - minVec.X,
			Y:     maxVec.Y - minVec.Y,
			Z:     maxVec.Z - minVec.Z,
		}
		anchorVec := minVec

		return &ast.SceneDecl{
			Token:  tok,
			Min:    minVec,
			Max:    maxVec,
			Size:   sizeVec,
			Anchor: &anchorVec,
		}
	}

	// scene [w, h, d] (@ [x, y, z])?
	sizeVec := p.parseVec3()
	var anchor *ast.Vec3
	minVec := ast.Vec3{Token: sizeVec.Token, X: 0, Y: 0, Z: 0}

	if p.match(token.AT) {
		a := p.parseVec3()
		anchor = &a
		minVec = a
	}

	maxVec := ast.Vec3{
		Token: sizeVec.Token,
		X:     minVec.X + sizeVec.X,
		Y:     minVec.Y + sizeVec.Y,
		Z:     minVec.Z + sizeVec.Z,
	}

	return &ast.SceneDecl{
		Token:  tok,
		Min:    minVec,
		Max:    maxVec,
		Size:   sizeVec,
		Anchor: anchor,
	}
}

// zone <name> [w, h, d] (@ [x, y, z])? (#tag)*
func (p *Parser) parseZone() *ast.ZoneDecl {
	tok := p.tok
	p.next() // skip 'zone'

	name := p.expect(token.IDENT)
	size := p.parseVec3()

	var anchor *ast.Vec3
	if p.match(token.AT) {
		a := p.parseVec3()
		anchor = &a
	}

	tags := make([]string, 0)
	for p.match(token.HASH) {
		tags = append(tags, p.expect(token.IDENT))
	}

	return &ast.ZoneDecl{
		Token:  tok,
		Name:   name,
		Size:   size,
		Anchor: anchor,
		Tags:   tags,
	}
}

// item <name> [w, h, d] (static | can_rotate | #tag)* ({ facings })?
func (p *Parser) parseItem() *ast.ItemDecl {
	tok := p.tok
	p.next() // skip 'item'

	name := p.expect(token.IDENT)
	size := p.parseVec3()

	var isStatic, canRotate bool
	tags := make([]string, 0)

	for {
		if p.match(token.STATIC) {
			isStatic = true
		} else if p.match(token.CAN_ROTATE) {
			canRotate = true
		} else if p.match(token.HASH) {
			tags = append(tags, p.expect(token.IDENT))
		} else {
			break
		}
	}

	facings := make([]ast.Facing, 0)
	if p.match(token.LBRACE) {
		for !p.is(token.RBRACE) && !p.is(token.EOF) {
			fTok := p.tok
			marker := p.expect(token.IDENT)
			p.expect(token.ARROW)
			vec := p.parseVec3()
			facings = append(facings, ast.Facing{Token: fTok, Marker: marker, Vector: vec})
		}
		p.expect(token.RBRACE)
	}

	return &ast.ItemDecl{
		Token:     tok,
		Name:      name,
		Size:      size,
		IsStatic:  isStatic,
		CanRotate: canRotate,
		Tags:      tags,
		Facings:   facings,
	}
}

// spawn <name> (id=n | count | @ [x, y, z])*
func (p *Parser) parseSpawn() *ast.SpawnDecl {
	tok := p.tok
	p.next() // skip 'spawn'

	name := p.expect(token.IDENT)

	var id *uint64
	var anchor *ast.Vec3
	count := 1

	for !p.isAtTopLevel() {
		if p.match(token.ID) {
			p.expect(token.ASSIGN)
			val := uint64(p.parseInt())
			id = &val
		} else if p.is(token.INT) {
			count = int(p.parseInt())
		} else if p.match(token.AT) {
			a := p.parseVec3()
			anchor = &a
		} else {
			break
		}
	}

	return &ast.SpawnDecl{
		Token:    tok,
		ItemName: name,
		ID:       id,
		Count:    count,
		Anchor:   anchor,
	}
}

// parseSelector parses *, #tag, or item_name
func (p *Parser) parseSelector() ast.Selector {
	tok := p.tok
	if p.match(token.ASTERISK) {
		return ast.Selector{Token: tok, Kind: ast.SelectorWildcard}
	}
	if p.match(token.HASH) {
		return ast.Selector{Token: tok, Kind: ast.SelectorTag, Value: p.expect(token.IDENT)}
	}
	if p.is(token.IDENT) {
		name := p.expect(token.IDENT)
		return ast.Selector{Token: tok, Kind: ast.SelectorItem, Value: name}
	}
	p.errorf("expected selector (*, #tag, or name), got %s (%q)", p.tok.Type, p.tok.Literal)
	p.next()
	return ast.Selector{Token: tok}
}

func (p *Parser) parseParamValue() any {
	if p.is(token.LBRACKET) {
		return p.parseVec3()
	}
	if p.is(token.MINUS) || p.is(token.INT) {
		return p.parseInt()
	}
	if p.is(token.IDENT) {
		return p.expect(token.IDENT)
	}
	p.errorf("unexpected parameter value: %s (%q)", p.tok.Type, p.tok.Literal)
	p.next()
	return nil
}

// rule <kind> <target>[.marker] (<secondary> | to <secondary>)? (-> [vec] | [vec])* (key=val)* (flags)*
func (p *Parser) parseRule() *ast.RuleDecl {
	tok := p.tok
	p.next() // skip 'rule'

	kind := p.expect(token.IDENT)
	rule := &ast.RuleDecl{
		Token:   tok,
		Kind:    kind,
		Params:  make(map[string]any),
		Vectors: make([]ast.Vec3, 0),
		Flags:   make([]string, 0),
	}

	// 1. Primary target selector
	if p.isSelectorStart() {
		rule.Target = p.parseSelector()
		if p.match(token.DOT) {
			rule.Marker = p.expect(token.IDENT)
		}
	}

	// 2. Remaining rule arguments
	for !p.isAtTopLevel() {
		if p.match(token.ARROW) {
			rule.Vectors = append(rule.Vectors, p.parseVec3())
		} else if p.is(token.LBRACKET) {
			rule.Vectors = append(rule.Vectors, p.parseVec3())
		} else if p.is(token.IDENT) && p.peekIs(token.ASSIGN) {
			key := p.expect(token.IDENT)
			p.expect(token.ASSIGN)
			rule.Params[key] = p.parseParamValue()
		} else if p.is(token.TO) || (p.is(token.IDENT) && p.tok.Literal == "to") {
			p.next() // skip "to"
			sec := p.parseSelector()
			rule.Secondary = &sec
		} else if rule.Secondary == nil && p.isSelectorStart() && p.tok.Literal != "closest" {
			sec := p.parseSelector()
			rule.Secondary = &sec
		} else if p.is(token.IDENT) {
			rule.Flags = append(rule.Flags, p.expect(token.IDENT))
		} else {
			p.errorf("unexpected token in rule: %s (%q)", p.tok.Type, p.tok.Literal)
			p.next()
		}
	}

	return rule
}
