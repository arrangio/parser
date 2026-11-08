package parser

import (
	"strconv"

	"github.com/arrangio/parser/ast"
	"github.com/arrangio/parser/token"
)

func (p *Parser) parseInt() int64 {
	sign := int64(1)
	if p.match(token.MINUS) {
		sign = -1
	}
	lit := p.expect(token.INT)
	val, _ := strconv.ParseInt(lit, 10, 64)
	return sign * val
}

func (p *Parser) parseVec3() ast.Vec3 {
	tok := p.tok
	p.expect(token.LBRACKET)
	x := p.parseInt()
	p.match(token.COMMA)
	y := p.parseInt()
	p.match(token.COMMA)
	z := p.parseInt()
	p.expect(token.RBRACKET)
	return ast.Vec3{Token: tok, X: x, Y: y, Z: z}
}

func (p *Parser) parseScene() *ast.SceneDecl {
	tok := p.tok
	p.next() // skip 'scene'

	if p.match(token.BOUNDS) {
		minVec := p.parseVec3()
		p.expect(token.TO)
		maxVec := p.parseVec3()
		size := ast.Vec3{Token: minVec.Token, X: maxVec.X - minVec.X, Y: maxVec.Y - minVec.Y, Z: maxVec.Z - minVec.Z}
		return &ast.SceneDecl{Token: tok, Min: minVec, Max: maxVec, Size: size, Anchor: &minVec}
	}

	size := p.parseVec3()
	var anchor *ast.Vec3
	minVec := ast.Vec3{Token: size.Token}
	if p.match(token.AT) {
		a := p.parseVec3()
		anchor, minVec = &a, a
	}
	maxVec := ast.Vec3{Token: size.Token, X: minVec.X + size.X, Y: minVec.Y + size.Y, Z: minVec.Z + size.Z}
	return &ast.SceneDecl{Token: tok, Min: minVec, Max: maxVec, Size: size, Anchor: anchor}
}

func (p *Parser) parseZone() *ast.ZoneDecl {
	tok := p.tok
	p.next() // skip 'zone'
	name, size := p.expect(token.IDENT), p.parseVec3()

	var anchor *ast.Vec3
	if p.match(token.AT) {
		a := p.parseVec3()
		anchor = &a
	}
	var tags []string
	for p.match(token.HASH) {
		tags = append(tags, p.expect(token.IDENT))
	}
	return &ast.ZoneDecl{Token: tok, Name: name, Size: size, Anchor: anchor, Tags: tags}
}

func (p *Parser) parseItem() *ast.ItemDecl {
	tok := p.tok
	p.next() // skip 'item'
	name, size := p.expect(token.IDENT), p.parseVec3()

	var isStatic, canRotate bool
	var tags []string
	for {
		switch {
		case p.match(token.STATIC):
			isStatic = true
		case p.match(token.CAN_ROTATE):
			canRotate = true
		case p.match(token.HASH):
			tags = append(tags, p.expect(token.IDENT))
		default:
			goto itemDone
		}
	}
itemDone:
	var facings []ast.Facing
	if p.match(token.LBRACE) {
		for !p.is(token.RBRACE) && !p.is(token.EOF) {
			fTok, marker := p.tok, p.expect(token.IDENT)
			p.expect(token.ARROW)
			facings = append(facings, ast.Facing{Token: fTok, Marker: marker, Vector: p.parseVec3()})
		}
		p.expect(token.RBRACE)
	}
	return &ast.ItemDecl{Token: tok, Name: name, Size: size, IsStatic: isStatic, CanRotate: canRotate, Tags: tags, Facings: facings}
}

func (p *Parser) parseSpawn() *ast.SpawnDecl {
	tok := p.tok
	p.next() // skip 'spawn'
	name := p.expect(token.IDENT)

	var id *uint64
	var anchor *ast.Vec3
	count := 1

	for !p.isAtTopLevel() {
		switch {
		case p.match(token.ID):
			p.expect(token.ASSIGN)
			val := uint64(p.parseInt())
			id = &val
		case p.is(token.INT):
			count = int(p.parseInt())
		case p.match(token.AT):
			a := p.parseVec3()
			anchor = &a
		default:
			goto spawnDone
		}
	}
spawnDone:
	return &ast.SpawnDecl{Token: tok, ItemName: name, ID: id, Count: count, Anchor: anchor}
}

func (p *Parser) parseSelector() ast.Selector {
	tok := p.tok
	switch {
	case p.match(token.ASTERISK):
		return ast.Selector{Token: tok, Kind: ast.SelectorWildcard}
	case p.match(token.HASH):
		return ast.Selector{Token: tok, Kind: ast.SelectorTag, Value: p.expect(token.IDENT)}
	case p.is(token.IDENT):
		return ast.Selector{Token: tok, Kind: ast.SelectorItem, Value: p.expect(token.IDENT)}
	default:
		p.errorf("expected selector (*, #tag, or name), got %s (%q)", p.tok.Type, p.tok.Literal)
		p.next()
		return ast.Selector{Token: tok}
	}
}

func (p *Parser) parseParamValue() any {
	switch {
	case p.is(token.LBRACKET):
		return p.parseVec3()
	case p.is(token.MINUS), p.is(token.INT):
		return p.parseInt()
	case p.is(token.IDENT):
		return p.expect(token.IDENT)
	default:
		p.errorf("unexpected parameter value: %s (%q)", p.tok.Type, p.tok.Literal)
		p.next()
		return nil
	}
}
