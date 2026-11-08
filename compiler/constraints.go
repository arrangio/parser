package compiler

import (
	"fmt"
	"strings"

	"github.com/arrangio/core/geometry"
	"github.com/arrangio/core/rules"
	"github.com/arrangio/parser/ast"
	"github.com/arrangio/parser/token"
)

// compileConstraint translates a mathematical constraint expression into a rules.Rule
func (c *compiler) compileConstraint(stmt *ast.ConstraintStmt, env map[string]any) (rules.Rule, error) {
	if stmt == nil || stmt.Expr == nil {
		return nil, fmt.Errorf("empty constraint statement")
	}

	return c.compileExprRule(stmt.Expr, env)
}

func (c *compiler) compileExprRule(expr ast.Expr, env map[string]any) (rules.Rule, error) {
	switch e := expr.(type) {
	case *ast.BinaryExpr:
		return c.compileBinaryConstraint(e, env)

	case *ast.CallExpr:
		if e.Function == "no_collision" {
			target := rules.Selector{MatchAny: true}
			sec := rules.Selector{MatchAny: true}
			if len(e.Args) > 0 {
				var err error
				target, err = c.resolveExprSelector(e.Args[0], env)
				if err != nil {
					return nil, err
				}
			}
			if len(e.Args) > 1 {
				var err error
				sec, err = c.resolveExprSelector(e.Args[1], env)
				if err != nil {
					return nil, err
				}
			}
			return &rules.NoCollisionRule{Target: target, Obstacle: sec}, nil
		}
		return nil, fmt.Errorf("unsupported constraint function %q", e.Function)

	default:
		return nil, fmt.Errorf("unsupported constraint expression %T", expr)
	}
}

func (c *compiler) compileBinaryConstraint(bin *ast.BinaryExpr, env map[string]any) (rules.Rule, error) {
	// 1. Distance constraints: dist(A, B) >= expr | dist(A, B) <= expr
	if call, ok := bin.Left.(*ast.CallExpr); ok && call.Function == "dist" {
		if len(call.Args) < 2 {
			return nil, fmt.Errorf("dist() requires 2 arguments, got %d", len(call.Args))
		}
		target, err := c.resolveExprSelector(call.Args[0], env)
		if err != nil {
			return nil, fmt.Errorf("dist target: %w", err)
		}
		obstacle, err := c.resolveExprSelector(call.Args[1], env)
		if err != nil {
			return nil, fmt.Errorf("dist obstacle: %w", err)
		}

		distVal, err := evalInt(bin.Right, env)
		if err != nil {
			return nil, fmt.Errorf("dist limit: %w", err)
		}

		switch bin.Op {
		case token.GTE:
			return &rules.SeparationRule{Target: target, Obstacle: obstacle, MinDistance: distVal}, nil
		case token.GT:
			return &rules.SeparationRule{Target: target, Obstacle: obstacle, MinDistance: distVal + 1}, nil
		case token.LTE, token.LT:
			return &rules.ProximityRule{Target: target, To: obstacle, MaxDist: distVal}, nil
		default:
			return nil, fmt.Errorf("unsupported comparison operator for dist(): %s", bin.Op)
		}
	}

	// 2. Member constraints: chair.z == 0, chair.x == desk.x, chair.front -> [0, 1, 0]
	if mem, ok := bin.Left.(*ast.MemberExpr); ok {
		target, err := c.resolveExprSelector(mem.Object, env)
		if err != nil {
			return nil, err
		}

		// Orientation constraint: chair.front -> [0, 1, 0]
		if bin.Op == token.ARROW {
			vec := geometry.Point64{X: 0, Y: 1, Z: 0}
			if pt, err := evalPoint64(bin.Right, env); err == nil {
				vec = pt
			}
			return &rules.OrientationRule{Target: target, Marker: mem.Property, TargetVector: vec}, nil
		}

		// Axis constraints: chair.z == 0 or chair.x == desk.x
		if axis, ok := parseAxis(mem.Property); ok {
			// Alignment: chair.x == desk.x
			if rightMem, ok := bin.Right.(*ast.MemberExpr); ok {
				if rAxis, ok := parseAxis(rightMem.Property); ok && rAxis == axis && bin.Op == token.EQ {
					return &rules.AlignmentRule{Target: target, Axis: axis, Radius: 50}, nil
				}
			}

			// AxisRestriction: chair.z == 0, chair.x <= 500
			op, ok := parseOp(bin.Op)
			if !ok {
				return nil, fmt.Errorf("unsupported axis operator %s", bin.Op)
			}

			refVal, err := evalInt(bin.Right, env)
			if err != nil {
				return nil, fmt.Errorf("axis ref value: %w", err)
			}
			return &rules.AxisRestrictionRule{Target: target, Axis: axis, Op: op, Ref: refVal}, nil
		}
	}

	// 3. Containment constraint: chair in [0, 0, 0]..[1000, 800, 300]
	if bin.Op == token.IN {
		target, err := c.resolveExprSelector(bin.Left, env)
		if err != nil {
			return nil, err
		}

		if rng, ok := bin.Right.(*ast.RangeExpr); ok {
			minPt, err := evalPoint64(rng.From, env)
			if err != nil {
				return nil, fmt.Errorf("containment min: %w", err)
			}
			maxPt, err := evalPoint64(rng.To, env)
			if err != nil {
				return nil, fmt.Errorf("containment max: %w", err)
			}
			return &rules.ContainmentRule{Target: target, Min: minPt, Max: maxPt}, nil
		}
	}

	// 4. Zone exclusion: chair not in forbidden_zone
	if bin.Op == token.NOT {
		target, err := c.resolveExprSelector(bin.Left, env)
		if err != nil {
			return nil, err
		}
		zone, err := c.resolveExprSelector(bin.Right, env)
		if err != nil {
			return nil, err
		}
		return &rules.ZoneExclusionRule{Target: target, Zone: zone}, nil
	}

	return nil, fmt.Errorf("unsupported binary constraint: %s", bin.Op)
}

func (c *compiler) resolveExprSelector(expr ast.Expr, env map[string]any) (rules.Selector, error) {
	if expr == nil {
		return rules.Selector{MatchAny: true}, nil
	}

	switch e := expr.(type) {
	case *ast.IdentExpr:
		name := e.Name
		if env != nil {
			if v, ok := env[name]; ok {
				if s, ok := v.(string); ok {
					name = s
				} else if id, ok := v.(*ast.IdentExpr); ok {
					name = id.Name
				}
			}
		}
		return c.resolveSelectorString(name)

	default:
		return rules.Selector{}, fmt.Errorf("cannot resolve selector from %T", expr)
	}
}

func parseAxis(prop string) (uint8, bool) {
	switch strings.ToLower(prop) {
	case "x":
		return rules.AxisX, true
	case "y":
		return rules.AxisY, true
	case "z":
		return rules.AxisZ, true
	default:
		return 0, false
	}
}

func parseOp(t token.TokenType) (rules.Operator, bool) {
	switch t {
	case token.EQ:
		return rules.OpEq, true
	case token.NOT_EQ:
		return rules.OpNot, true
	case token.LT:
		return rules.OpLt, true
	case token.LTE:
		return rules.OpLe, true
	case token.GT:
		return rules.OpGt, true
	case token.GTE:
		return rules.OpGe, true
	default:
		return 0, false
	}
}
