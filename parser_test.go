package parser_test

import (
	"testing"

	"github.com/arrangio/parser"
	"github.com/arrangio/parser/ast"
	"github.com/arrangio/parser/lexer"
)

func TestParser_ZoneDeclarations(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedName   string
		expectedSize   ast.Vec3
		expectedAnchor *ast.Vec3
		expectedTags   []string
	}{
		{
			name:         "simple zone",
			input:        "zone room [1000, 800, 300]",
			expectedName: "room",
			expectedSize: ast.Vec3{X: 1000, Y: 800, Z: 300},
			expectedTags: []string{},
		},
		{
			name:           "zone with anchor and tags",
			input:          "zone living_room [500, 600, 250] @ [100, 200, 0] #indoor #lounge",
			expectedName:   "living_room",
			expectedSize:   ast.Vec3{X: 500, Y: 600, Z: 250},
			expectedAnchor: &ast.Vec3{X: 100, Y: 200, Z: 0},
			expectedTags:   []string{"indoor", "lounge"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			l := lexer.New(tc.input)
			p := parser.New(l)
			prog, err := p.ParseProgram()
			if err != nil {
				t.Fatalf("ParseProgram failed: %v, errors=%v", err, p.Errors())
			}
			if len(prog.Zones) != 1 {
				t.Fatalf("expected 1 zone, got %d", len(prog.Zones))
			}

			zone := prog.Zones[0]
			if zone.Name != tc.expectedName {
				t.Errorf("zone name: expected %q, got %q", tc.expectedName, zone.Name)
			}
			if zone.Size.X != tc.expectedSize.X || zone.Size.Y != tc.expectedSize.Y || zone.Size.Z != tc.expectedSize.Z {
				t.Errorf("zone size: expected %v, got %v", tc.expectedSize, zone.Size)
			}
			if tc.expectedAnchor != nil {
				if zone.Anchor == nil {
					t.Fatalf("expected anchor %v, got nil", tc.expectedAnchor)
				}
				if zone.Anchor.X != tc.expectedAnchor.X || zone.Anchor.Y != tc.expectedAnchor.Y || zone.Anchor.Z != tc.expectedAnchor.Z {
					t.Errorf("zone anchor: expected %v, got %v", tc.expectedAnchor, zone.Anchor)
				}
			} else if zone.Anchor != nil {
				t.Errorf("expected nil anchor, got %v", zone.Anchor)
			}

			if len(zone.Tags) != len(tc.expectedTags) {
				t.Fatalf("tags count: expected %d, got %d", len(tc.expectedTags), len(zone.Tags))
			}
			for i, tag := range tc.expectedTags {
				if zone.Tags[i] != tag {
					t.Errorf("tag[%d]: expected %q, got %q", i, tag, zone.Tags[i])
				}
			}
		})
	}
}

func TestParser_ItemDeclarations(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectedName      string
		expectedSize      ast.Vec3
		expectedStatic    bool
		expectedCanRotate bool
		expectedTags      []string
		expectedFacings   []ast.Facing
	}{
		{
			name:              "simple item with modifiers and tags",
			input:             "item desk [120, 60, 75] can_rotate #furniture #wood",
			expectedName:      "desk",
			expectedSize:      ast.Vec3{X: 120, Y: 60, Z: 75},
			expectedCanRotate: true,
			expectedTags:      []string{"furniture", "wood"},
			expectedFacings:   []ast.Facing{},
		},
		{
			name:           "static item without facings",
			input:          "item column [40, 40, 300] static #obstacle",
			expectedName:   "column",
			expectedSize:   ast.Vec3{X: 40, Y: 40, Z: 300},
			expectedStatic: true,
			expectedTags:   []string{"obstacle"},
		},
		{
			name: "item with facings",
			input: `item chair [50, 50, 90] can_rotate #furniture {
				front -> [0, -1, 0]
				top -> [0, 0, 1]
			}`,
			expectedName:      "chair",
			expectedSize:      ast.Vec3{X: 50, Y: 50, Z: 90},
			expectedCanRotate: true,
			expectedTags:      []string{"furniture"},
			expectedFacings: []ast.Facing{
				{Marker: "front", Vector: ast.Vec3{X: 0, Y: -1, Z: 0}},
				{Marker: "top", Vector: ast.Vec3{X: 0, Y: 0, Z: 1}},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			l := lexer.New(tc.input)
			p := parser.New(l)
			prog, err := p.ParseProgram()
			if err != nil {
				t.Fatalf("ParseProgram failed: %v, errors=%v", err, p.Errors())
			}
			if len(prog.Items) != 1 {
				t.Fatalf("expected 1 item, got %d", len(prog.Items))
			}

			item := prog.Items[0]
			if item.Name != tc.expectedName {
				t.Errorf("item name: expected %q, got %q", tc.expectedName, item.Name)
			}
			if item.IsStatic != tc.expectedStatic {
				t.Errorf("item isStatic: expected %v, got %v", tc.expectedStatic, item.IsStatic)
			}
			if item.CanRotate != tc.expectedCanRotate {
				t.Errorf("item canRotate: expected %v, got %v", tc.expectedCanRotate, item.CanRotate)
			}
			if len(item.Facings) != len(tc.expectedFacings) {
				t.Fatalf("facings count: expected %d, got %d", len(tc.expectedFacings), len(item.Facings))
			}
			for i, expF := range tc.expectedFacings {
				gotF := item.Facings[i]
				if gotF.Marker != expF.Marker {
					t.Errorf("facing[%d] marker: expected %q, got %q", i, expF.Marker, gotF.Marker)
				}
				if gotF.Vector.X != expF.Vector.X || gotF.Vector.Y != expF.Vector.Y || gotF.Vector.Z != expF.Vector.Z {
					t.Errorf("facing[%d] vector: expected %v, got %v", i, expF.Vector, gotF.Vector)
				}
			}
		})
	}
}

func TestParser_SpawnDeclarations(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedItem   string
		expectedCount  int
		expectedAnchor *ast.Vec3
	}{
		{
			name:          "spawn single item default",
			input:         "spawn desk",
			expectedItem:  "desk",
			expectedCount: 1,
		},
		{
			name:          "spawn multiple items",
			input:         "spawn chair 4",
			expectedItem:  "chair",
			expectedCount: 4,
		},
		{
			name:           "spawn with count and anchor",
			input:          "spawn column 2 @ [100, 200, 0]",
			expectedItem:   "column",
			expectedCount:  2,
			expectedAnchor: &ast.Vec3{X: 100, Y: 200, Z: 0},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			l := lexer.New(tc.input)
			p := parser.New(l)
			prog, err := p.ParseProgram()
			if err != nil {
				t.Fatalf("ParseProgram failed: %v, errors=%v", err, p.Errors())
			}
			if len(prog.Spawns) != 1 {
				t.Fatalf("expected 1 spawn, got %d", len(prog.Spawns))
			}

			spawn := prog.Spawns[0]
			if spawn.ItemName != tc.expectedItem {
				t.Errorf("spawn item: expected %q, got %q", tc.expectedItem, spawn.ItemName)
			}
			if spawn.Count != tc.expectedCount {
				t.Errorf("spawn count: expected %d, got %d", tc.expectedCount, spawn.Count)
			}
			if tc.expectedAnchor != nil {
				if spawn.Anchor == nil {
					t.Fatalf("expected anchor %v, got nil", tc.expectedAnchor)
				}
				if spawn.Anchor.X != tc.expectedAnchor.X || spawn.Anchor.Y != tc.expectedAnchor.Y || spawn.Anchor.Z != tc.expectedAnchor.Z {
					t.Errorf("spawn anchor: expected %v, got %v", tc.expectedAnchor, spawn.Anchor)
				}
			}
		})
	}
}

func TestParser_RuleDeclarations(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedKind string
		validate     func(t *testing.T, r *ast.RuleDecl)
	}{
		{
			name:         "no_collision wildcard",
			input:        "rule no_collision * *",
			expectedKind: "no_collision",
			validate: func(t *testing.T, r *ast.RuleDecl) {
				if !r.Target.IsWildcard() {
					t.Errorf("expected target wildcard, got %v", r.Target)
				}
				if r.Secondary == nil || !r.Secondary.IsWildcard() {
					t.Errorf("expected secondary wildcard, got %v", r.Secondary)
				}
			},
		},
		{
			name:         "separation with param",
			input:        "rule separation chair table min=10",
			expectedKind: "separation",
			validate: func(t *testing.T, r *ast.RuleDecl) {
				if r.Target.Value != "chair" || !r.Target.IsItem() {
					t.Errorf("expected target chair, got %v", r.Target)
				}
				if r.Secondary == nil || r.Secondary.Value != "table" {
					t.Errorf("expected secondary table, got %v", r.Secondary)
				}
				if minVal, ok := r.Params["min"].(int64); !ok || minVal != 10 {
					t.Errorf("expected min=10, got %v", r.Params["min"])
				}
			},
		},
		{
			name:         "proximity with 'to' and flags",
			input:        "rule proximity chair to table max=150 closest",
			expectedKind: "proximity",
			validate: func(t *testing.T, r *ast.RuleDecl) {
				if r.Target.Value != "chair" {
					t.Errorf("expected target chair, got %v", r.Target)
				}
				if r.Secondary == nil || r.Secondary.Value != "table" {
					t.Errorf("expected secondary table, got %v", r.Secondary)
				}
				if maxVal, ok := r.Params["max"].(int64); !ok || maxVal != 150 {
					t.Errorf("expected max=150, got %v", r.Params["max"])
				}
				if len(r.Flags) != 1 || r.Flags[0] != "closest" {
					t.Errorf("expected flags [closest], got %v", r.Flags)
				}
			},
		},
		{
			name:         "orientation with facing marker and direction vector",
			input:        "rule orientation chair.front -> [0, 1, 0]",
			expectedKind: "orientation",
			validate: func(t *testing.T, r *ast.RuleDecl) {
				if r.Target.Value != "chair" {
					t.Errorf("expected target chair, got %v", r.Target)
				}
				if r.Marker != "front" {
					t.Errorf("expected marker front, got %q", r.Marker)
				}
				if len(r.Vectors) != 1 || r.Vectors[0].Y != 1 {
					t.Errorf("expected vector [0, 1, 0], got %v", r.Vectors)
				}
			},
		},
		{
			name:         "containment with two bounding vectors",
			input:        "rule containment #furniture [0, 0, 0] [1000, 800, 300]",
			expectedKind: "containment",
			validate: func(t *testing.T, r *ast.RuleDecl) {
				if !r.Target.IsTag() || r.Target.Value != "furniture" {
					t.Errorf("expected target #furniture, got %v", r.Target)
				}
				if len(r.Vectors) != 2 {
					t.Fatalf("expected 2 vectors, got %d", len(r.Vectors))
				}
				if r.Vectors[0].X != 0 || r.Vectors[1].X != 1000 {
					t.Errorf("unexpected vectors: %v", r.Vectors)
				}
			},
		},
		{
			name:         "clearance with vector parameter",
			input:        "rule clearance desk wall padding=[10, 10, 0]",
			expectedKind: "clearance",
			validate: func(t *testing.T, r *ast.RuleDecl) {
				if r.Target.Value != "desk" {
					t.Errorf("expected target desk, got %v", r.Target)
				}
				if r.Secondary == nil || r.Secondary.Value != "wall" {
					t.Errorf("expected secondary wall, got %v", r.Secondary)
				}
				pad, ok := r.Params["padding"].(ast.Vec3)
				if !ok || pad.X != 10 || pad.Y != 10 || pad.Z != 0 {
					t.Errorf("expected padding=[10, 10, 0], got %v", r.Params["padding"])
				}
			},
		},
		{
			name:         "alignment with string and int params",
			input:        "rule alignment chair axis=x radius=50",
			expectedKind: "alignment",
			validate: func(t *testing.T, r *ast.RuleDecl) {
				if r.Target.Value != "chair" {
					t.Errorf("expected target chair, got %v", r.Target)
				}
				if axis, ok := r.Params["axis"].(string); !ok || axis != "x" {
					t.Errorf("expected axis=x, got %v", r.Params["axis"])
				}
				if rad, ok := r.Params["radius"].(int64); !ok || rad != 50 {
					t.Errorf("expected radius=50, got %v", r.Params["radius"])
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			l := lexer.New(tc.input)
			p := parser.New(l)
			prog, err := p.ParseProgram()
			if err != nil {
				t.Fatalf("ParseProgram failed: %v, errors=%v", err, p.Errors())
			}
			if len(prog.Rules) != 1 {
				t.Fatalf("expected 1 rule, got %d", len(prog.Rules))
			}

			rule := prog.Rules[0]
			if rule.Kind != tc.expectedKind {
				t.Errorf("rule kind: expected %q, got %q", tc.expectedKind, rule.Kind)
			}
			tc.validate(t, rule)
		})
	}
}

func TestParser_CompleteProgram(t *testing.T) {
	input := `
// Room zone
zone room [1000, 800, 300] @ [0, 0, 0] #indoor

// Items definition
item desk [120, 60, 75] can_rotate #furniture {
	front -> [0, 1, 0]
}

item chair [50, 50, 90] can_rotate #furniture {
	front -> [0, 1, 0]
}

item pillar [40, 40, 300] static #obstacle

// Spawns
spawn pillar 2 @ [200, 200, 0]
spawn desk 2
spawn chair 4

// Rules
rule containment #furniture [0, 0, 0] [1000, 800, 300]
rule no_collision * *
rule separation chair desk min=15
rule proximity chair to desk max=120 closest
rule orientation chair.front -> [0, 1, 0]
`

	l := lexer.New(input)
	p := parser.New(l)
	prog, err := p.ParseProgram()

	if err != nil {
		t.Fatalf("ParseProgram failed: %v, errors=%v", err, p.Errors())
	}

	if len(prog.Zones) != 1 {
		t.Errorf("expected 1 zone, got %d", len(prog.Zones))
	}
	if len(prog.Items) != 3 {
		t.Errorf("expected 3 items, got %d", len(prog.Items))
	}
	if len(prog.Spawns) != 3 {
		t.Errorf("expected 3 spawns, got %d", len(prog.Spawns))
	}
	if len(prog.Rules) != 5 {
		t.Errorf("expected 5 rules, got %d", len(prog.Rules))
	}
}

func TestParser_SyntaxErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name:        "missing closing bracket in vector",
			input:       "zone room [1000, 800",
			expectedErr: "expected integer coordinate",
		},
		{
			name:        "unexpected token at root",
			input:       "unknown_cmd foo",
			expectedErr: "unexpected token",
		},
		{
			name:        "item missing name",
			input:       "item [10, 10, 10]",
			expectedErr: "expected next token to be IDENT",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			l := lexer.New(tc.input)
			p := parser.New(l)
			_, err := p.ParseProgram()
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.expectedErr)
			}
			if len(p.Errors()) == 0 {
				t.Fatalf("expected recorded errors in p.Errors()")
			}
		})
	}
}
