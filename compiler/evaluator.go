package compiler

import (
	"fmt"

	"github.com/arrangio/core/geometry"
	"github.com/arrangio/parser/ast"
	"github.com/arrangio/parser/token"
)

// evalInt evaluates an expression to an int64 value, substituting variables from env
func evalInt(expr ast.Expr, env map[string]any) (int64, error) {
	if expr == nil {
		return 0, fmt.Errorf("nil expression")
	}

	switch e := expr.(type) {
	case *ast.NumberExpr:
		return e.Value, nil

	case *ast.IdentExpr:
		if env != nil {
			if v, ok := env[e.Name]; ok {
				switch val := v.(type) {
				case int64:
					return val, nil
				case int:
					return int64(val), nil
				case ast.Expr:
					return evalInt(val, env)
				default:
					return 0, fmt.Errorf("identifier %q is not an integer (got %T)", e.Name, v)
				}
			}
		}
		return 0, fmt.Errorf("unresolved variable %q in arithmetic expression", e.Name)

	case *ast.UnaryExpr:
		val, err := evalInt(e.Right, env)
		if err != nil {
			return 0, err
		}
		if e.Op == token.MINUS {
			return -val, nil
		}
		return val, nil

	case *ast.BinaryExpr:
		left, err := evalInt(e.Left, env)
		if err != nil {
			return 0, err
		}
		right, err := evalInt(e.Right, env)
		if err != nil {
			return 0, err
		}

		switch e.Op {
		case token.PLUS:
			return left + right, nil
		case token.MINUS:
			return left - right, nil
		case token.ASTERISK:
			return left * right, nil
		case token.SLASH:
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return left / right, nil
		default:
			return 0, fmt.Errorf("unsupported arithmetic operator %s", e.Op)
		}

	default:
		return 0, fmt.Errorf("cannot evaluate %T to integer", expr)
	}
}

// evalPoint64 evaluates a vector expression to geometry.Point64
func evalPoint64(expr ast.Expr, env map[string]any) (geometry.Point64, error) {
	switch e := expr.(type) {
	case *ast.VectorExpr:
		x, err := evalInt(e.X, env)
		if err != nil {
			return geometry.Point64{}, fmt.Errorf("vector X: %w", err)
		}
		y, err := evalInt(e.Y, env)
		if err != nil {
			return geometry.Point64{}, fmt.Errorf("vector Y: %w", err)
		}
		z, err := evalInt(e.Z, env)
		if err != nil {
			return geometry.Point64{}, fmt.Errorf("vector Z: %w", err)
		}
		return geometry.Point64{X: x, Y: y, Z: z}, nil

	case *ast.IdentExpr:
		if env != nil {
			if v, ok := env[e.Name]; ok {
				if ex, ok := v.(ast.Expr); ok {
					return evalPoint64(ex, env)
				}
			}
		}
		return geometry.Point64{}, fmt.Errorf("unresolved vector %q", e.Name)

	default:
		return geometry.Point64{}, fmt.Errorf("expected vector expression, got %T", expr)
	}
}
