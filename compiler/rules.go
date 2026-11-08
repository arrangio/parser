package compiler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/arrangio/core/geometry"
	"github.com/arrangio/core/rules"
	"github.com/arrangio/core/tags"
	"github.com/arrangio/parser/ast"
)

func (c *compiler) compileRule(r *ast.RuleDecl) (rules.Rule, error) {
	target, err := c.resolveSelector(&r.Target)
	if err != nil {
		return nil, fmt.Errorf("rule %q target: %w", r.Kind, err)
	}
	sec, _ := c.resolveSelector(r.Secondary)

	switch r.Kind {
	case "no_collision":
		return &rules.NoCollisionRule{Target: target, Obstacle: sec}, nil
	case "separation":
		return &rules.SeparationRule{Target: target, Obstacle: sec, MinDistance: intParam(r.Params, "min", 10)}, nil
	case "proximity":
		return &rules.ProximityRule{Target: target, To: sec, MaxDist: intParam(r.Params, "max", 100), RequireClosest: hasFlag(r.Flags, "closest")}, nil
	case "orientation":
		marker := r.Marker
		if marker == "" {
			marker = "front"
		}
		vec := geometry.Point64{X: 0, Y: 1, Z: 0}
		if len(r.Vectors) > 0 {
			vec = toPoint64(&r.Vectors[0])
		}
		return &rules.OrientationRule{Target: target, Marker: marker, TargetVector: vec}, nil
	case "containment":
		var bMin, bMax geometry.Point64
		if len(r.Vectors) >= 2 {
			bMin, bMax = toPoint64(&r.Vectors[0]), toPoint64(&r.Vectors[1])
		}
		return &rules.ContainmentRule{Target: target, Min: bMin, Max: bMax}, nil
	case "clearance":
		var pad geometry.Point64
		if v, ok := r.Params["padding"].(ast.Vec3); ok {
			pad = toPoint64(&v)
		}
		return &rules.ClearanceRule{Target: target, Obstacle: sec, Padding: pad}, nil
	case "alignment":
		axis := rules.AxisX
		if a, _ := r.Params["axis"].(string); a == "y" || a == "Y" {
			axis = rules.AxisY
		} else if a == "z" || a == "Z" {
			axis = rules.AxisZ
		}
		return &rules.AlignmentRule{Target: target, Axis: axis, Radius: intParam(r.Params, "radius", 50)}, nil
	case "zone_exclusion":
		return &rules.ZoneExclusionRule{Target: target, Zone: sec}, nil
	case "axis_restriction":
		axis := rules.AxisX
		if a, _ := r.Params["axis"].(string); a == "y" || a == "Y" {
			axis = rules.AxisY
		} else if a == "z" || a == "Z" {
			axis = rules.AxisZ
		}

		op := rules.OpEq
		if o, ok := r.Params["op"].(string); ok {
			switch o {
			case "==", "eq":
				op = rules.OpEq
			case "!=", "neq", "not":
				op = rules.OpNot
			case "<", "lt":
				op = rules.OpLt
			case "<=", "lte", "le":
				op = rules.OpLe
			case ">", "gt":
				op = rules.OpGt
			case ">=", "gte", "ge":
				op = rules.OpGe
			}
		}

		ref := intParam(r.Params, "ref", 0)
		return &rules.AxisRestrictionRule{Target: target, Axis: axis, Op: op, Ref: ref}, nil
	default:
		return nil, fmt.Errorf("unsupported rule %q", r.Kind)
	}
}

func (c *compiler) resolveSelector(s *ast.Selector) (rules.Selector, error) {
	if s == nil || s.IsWildcard() {
		return rules.Selector{MatchAny: true}, nil
	}
	if s.IsTag() {
		return rules.Selector{Mask: tags.NewMask().With(c.reg.getOrRegister(s.Value))}, nil
	}
	return c.resolveSelectorString(s.Value)
}

func (c *compiler) resolveSelectorString(name string) (rules.Selector, error) {
	if name == "" || name == "*" {
		return rules.Selector{MatchAny: true}, nil
	}
	if strings.HasPrefix(name, "#") {
		return rules.Selector{Mask: tags.NewMask().With(c.reg.getOrRegister(name[1:]))}, nil
	}
	if id, err := strconv.ParseUint(name, 10, 64); err == nil {
		return rules.Selector{TargetID: id}, nil
	}
	if tagID, ok := c.itemTags[name]; ok {
		return rules.Selector{Mask: tags.NewMask().With(tagID)}, nil
	}
	// Check if it's a zone name
	for _, z := range c.zones {
		if z.Name == name {
			return rules.Selector{TargetID: z.ID}, nil
		}
	}
	return rules.Selector{}, fmt.Errorf("unknown target %q", name)
}

func intParam(params map[string]any, key string, fallback int64) int64 {
	if v, ok := params[key].(int64); ok {
		return v
	}
	return fallback
}

func hasFlag(flags []string, target string) bool {
	for _, f := range flags {
		if f == target {
			return true
		}
	}
	return false
}
