package compiler_test

import (
	"testing"

	"github.com/arrangio/core/solver"
	"github.com/arrangio/parser/compiler"
)

func TestCompiler_BasicSceneCompilation(t *testing.T) {
	input := `
scene [1000, 800, 300]

zone room [1000, 800, 300] @ [0, 0, 0] #indoor

item desk [120, 60, 75] can_rotate #furniture {
	front -> [0, 1, 0]
}

item chair [50, 50, 90] can_rotate #furniture {
	front -> [0, 1, 0]
}

spawn desk 1 @ [500, 500, 0] id=10
spawn chair 2 @ [500, 400, 0]

rule no_collision * *
rule separation chair desk min=15
`

	scene, err := compiler.CompileSource(input)
	if err != nil {
		t.Fatalf("CompileSource failed: %v", err)
	}

	if scene.Scene == nil {
		t.Fatalf("expected scene declaration in CompiledScene, got nil")
	}
	if scene.Scene.Size.X != 1000 || scene.Scene.Size.Y != 800 || scene.Scene.Size.Z != 300 {
		t.Errorf("expected scene size [1000, 800, 300], got %v", scene.Scene.Size)
	}

	if len(scene.Zones) != 1 {
		t.Fatalf("expected 1 zone, got %d", len(scene.Zones))
	}
	if len(scene.Entities) != 3 {
		t.Fatalf("expected 3 entities (1 desk + 2 chairs), got %d", len(scene.Entities))
	}
	if len(scene.Rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(scene.Rules))
	}

	// Verify desk has ID 10
	desk := scene.Entities[0]
	if desk.Def.ID != 10 {
		t.Errorf("expected desk ID 10, got %d", desk.Def.ID)
	}

	// Verify chairs have sequential IDs
	if scene.Entities[1].Def.ID == scene.Entities[2].Def.ID {
		t.Errorf("chair IDs must be unique, both got %d", scene.Entities[1].Def.ID)
	}
}

func TestCompiler_EndToEndSolving(t *testing.T) {
	input := `
zone room [500, 500, 200] @ [0, 0, 0] #indoor

item box1 [20, 20, 20] can_rotate #box
item box2 [30, 30, 30] can_rotate #box

// Spawn two overlapping boxes at identical positions
spawn box1 1 @ [250, 250, 50] id=1
spawn box2 1 @ [250, 250, 50] id=2

rule no_collision * *
rule separation box1 box2 min=5
rule containment #box [0, 0, 0] [500, 500, 200]
`

	scene, err := compiler.CompileSource(input)
	if err != nil {
		t.Fatalf("CompileSource failed: %v", err)
	}

	cfg := solver.DefaultForceSolverConfig()
	cfg.MaxIterations = 50

	s := solver.NewForceSolver(scene.State, cfg)
	res := s.Solve()

	if !res.Converged {
		t.Fatalf("expected solver to converge, got reason=%v (stress: %d, violators: %d)",
			res.Reason, res.FinalStress, res.ViolatorsCount)
	}

	b1 := scene.Entities[0]
	b2 := scene.Entities[1]

	if b1.State.Anchor == b2.State.Anchor {
		t.Fatalf("entities failed to separate: both still at %v", b1.State.Anchor)
	}
}

func TestCompiler_SceneBoundsForm(t *testing.T) {
	input := `
scene bounds [-500, -400, 0] to [500, 400, 300]
item box [50, 50, 50]
spawn box 1 @ [0, 0, 0]
`
	scene, err := compiler.CompileSource(input)
	if err != nil {
		t.Fatalf("CompileSource failed: %v", err)
	}

	if scene.Scene == nil {
		t.Fatalf("expected scene, got nil")
	}
	if scene.Scene.Min.X != -500 || scene.Scene.Max.X != 500 {
		t.Errorf("expected bounds -500 to 500 on X, got min=%v max=%v", scene.Scene.Min, scene.Scene.Max)
	}
	if scene.Scene.Size.X != 1000 || scene.Scene.Size.Y != 800 || scene.Scene.Size.Z != 300 {
		t.Errorf("expected size [1000, 800, 300], got %v", scene.Scene.Size)
	}
}

func TestCompiler_AxisRestrictionRule(t *testing.T) {
	input := `
item chair [50, 50, 90]
spawn chair 1 @ [100, 100, 0]
rule axis_restriction chair axis=z op=eq ref=0
`
	scene, err := compiler.CompileSource(input)
	if err != nil {
		t.Fatalf("CompileSource failed: %v", err)
	}
	if len(scene.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(scene.Rules))
	}
}

func TestCompiler_MathConstraintsCompilation(t *testing.T) {
	input := `
scene [1000, 800, 300]

item desk [120, 60, 75] can_rotate #furniture { front -> [0, 1, 0] }
item chair [50, 50, 90] can_rotate #furniture { front -> [0, 1, 0] }

spawn desk 1 @ [500, 500, 0] id=10
spawn chair 1 @ [500, 400, 0] id=20

no_collision *
dist(chair, desk) >= 15 + 5 * 2
dist(chair, desk) <= 120
chair.front -> [0, 1, 0]
chair.z == 0
desk.z == 0
chair.x == desk.x
chair in [0, 0, 0]..[1000, 800, 300]
`

	scene, err := compiler.CompileSource(input)
	if err != nil {
		t.Fatalf("CompileSource failed: %v", err)
	}

	if len(scene.Entities) != 2 {
		t.Fatalf("expected 2 entities, got %d", len(scene.Entities))
	}
	// 8 math constraints compiled into core rules
	if len(scene.Rules) != 8 {
		t.Fatalf("expected 8 rules from math constraints, got %d", len(scene.Rules))
	}
}

func TestCompiler_RuleDefEndToEndSolving(t *testing.T) {
	input := `
scene [500, 500, 200]

rule_def workstation(seat, table, gap=15) {
    dist(seat, table) >= gap
    seat.front -> [0, 1, 0]
}

item desk [100, 60, 75]
item chair [40, 40, 80] can_rotate { front -> [0, 1, 0] }

// Two overlapping objects
spawn desk 1 @ [250, 250, 0] id=1
spawn chair 1 @ [250, 250, 0] id=2

no_collision *
rule workstation chair desk gap=20
`

	scene, err := compiler.CompileSource(input)
	if err != nil {
		t.Fatalf("CompileSource failed: %v", err)
	}

	// 1 no_collision + 2 rules inside workstation
	if len(scene.Rules) != 3 {
		t.Fatalf("expected 3 rules (1 global + 2 from rule_def), got %d", len(scene.Rules))
	}

	cfg := solver.DefaultForceSolverConfig()
	cfg.MaxIterations = 100

	s := solver.NewForceSolver(scene.State, cfg)
	res := s.Solve()

	if !res.Converged {
		t.Fatalf("expected solver to converge with rule_def, got reason=%v (stress: %d)", res.Reason, res.FinalStress)
	}

	d := scene.Entities[0]
	c := scene.Entities[1]

	if d.State.Anchor == c.State.Anchor {
		t.Fatalf("entities failed to separate: both at %v", d.State.Anchor)
	}
}

func TestCompiler_Errors(t *testing.T) {
	t.Run("unknown item spawn", func(t *testing.T) {
		input := `spawn ghost 1`
		_, err := compiler.CompileSource(input)
		if err == nil {
			t.Fatalf("expected error for unknown item spawn, got nil")
		}
	})

	t.Run("duplicate entity ID", func(t *testing.T) {
		input := `
item box [10, 10, 10]
spawn box 1 id=5
spawn box 1 id=5
`
		_, err := compiler.CompileSource(input)
		if err == nil {
			t.Fatalf("expected error for duplicate ID, got nil")
		}
	})
}
