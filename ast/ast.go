package ast

import (
	"fmt"

	"github.com/arrangio/parser/token"
)

// `Node` represents any node in the Abstract Syntax Tree
type Node interface {
	Pos() (line, col int)
}

// `Vec3` represents a 3D vector or dimensions [X, Y, Z]
type Vec3 struct {
	Token token.Token
	X     int64
	Y     int64
	Z     int64
}

func (v Vec3) Pos() (int, int) {
	return v.Token.Line, v.Token.Col
}

func (v Vec3) String() string {
	return fmt.Sprintf("[%d, %d, %d]", v.X, v.Y, v.Z)
}

// `SelectorKind` defines what kind of target a Selector points to
type SelectorKind int

const (
	SelectorWildcard SelectorKind = iota // matches all entities (*)
	SelectorTag                           // matches entities by tag (#furniture)
	SelectorItem                          // matches entities by template name (chair)
)

// `Selector` represents a target or obstacle selector
type Selector struct {
	Token token.Token
	Kind  SelectorKind
	Value string // tag name without '#' (e.g. "furniture") or item name (e.g. "chair")
}

func (s Selector) Pos() (int, int) {
	return s.Token.Line, s.Token.Col
}

func (s Selector) IsWildcard() bool {
	return s.Kind == SelectorWildcard
}

func (s Selector) IsTag() bool {
	return s.Kind == SelectorTag
}

func (s Selector) IsItem() bool {
	return s.Kind == SelectorItem
}

func (s Selector) String() string {
	switch s.Kind {
	case SelectorWildcard:
		return "*"
	case SelectorTag:
		return "#" + s.Value
	case SelectorItem:
		return s.Value
	default:
		return "<unknown selector>"
	}
}

// `Facing` represents a semantic facing direction vector, e.g. `front -> [0, -1, 0]`
type Facing struct {
	Token  token.Token
	Marker string // e.g. "front", "back", "seat"
	Vector Vec3
}

func (f Facing) Pos() (int, int) {
	return f.Token.Line, f.Token.Col
}

// `ZoneDecl` represents a `zone` declaration.
// Example: zone room [1000, 800, 300] @ [0, 0, 0] #indoor
type ZoneDecl struct {
	Token  token.Token
	Name   string
	Size   Vec3
	Anchor *Vec3    // optional, nil if omitted
	Tags   []string // tag names without '#'
}

func (z *ZoneDecl) Pos() (int, int) {
	return z.Token.Line, z.Token.Col
}

// `ItemDecl` represents an `item` template declaration
// Example: item chair [50, 50, 90] can_rotate #furniture { front -> [0, -1, 0] }
type ItemDecl struct {
	Token     token.Token
	Name      string
	Size      Vec3
	IsStatic  bool
	CanRotate bool
	Tags      []string // tag names without '#'
	Facings   []Facing // semantic edge orientations
}

func (i *ItemDecl) Pos() (int, int) {
	return i.Token.Line, i.Token.Col
}

// `SpawnDecl` represents a `spawn` directive.
// Example: spawn chair 4 @ [100, 100, 0]
type SpawnDecl struct {
	Token    token.Token
	ItemName string
	Count    int   // defaults to 1 if omitted
	Anchor   *Vec3 // optional initial placement anchor
}

func (s *SpawnDecl) Pos() (int, int) {
	return s.Token.Line, s.Token.Col
}

// `RuleDecl` represents a `rule` declaration.
// Example: rule separation chair table min=10
// Example: rule orientation chair.front -> [0, 1, 0]
type RuleDecl struct {
	Token     token.Token
	Kind      string         // "no_collision", "separation", "proximity", "alignment", etc.
	Target    Selector       // primary selector
	Marker    string         // optional facing marker (e.g. "front" in "chair.front")
	Secondary *Selector      // optional secondary selector (obstacle, zone, "to" target)
	Params    map[string]any // named parameters: min=10, max=100, axis="x", radius=50, padding=Vec3
	Vectors   []Vec3         // positional vectors (containment bounds, orientation target vector)
	Flags     []string       // flag words (e.g. "closest")
}

func (r *RuleDecl) Pos() (int, int) {
	return r.Token.Line, r.Token.Col
}

// `Program` is the root node of the AST representing a complete DSL "program"
type Program struct {
	Zones  []*ZoneDecl
	Items  []*ItemDecl
	Spawns []*SpawnDecl
	Rules  []*RuleDecl
}

func (p *Program) Pos() (int, int) {
	if len(p.Zones) > 0 {
		return p.Zones[0].Pos()
	}
	if len(p.Items) > 0 {
		return p.Items[0].Pos()
	}
	if len(p.Spawns) > 0 {
		return p.Spawns[0].Pos()
	}
	if len(p.Rules) > 0 {
		return p.Rules[0].Pos()
	}
	return 1, 1
}
