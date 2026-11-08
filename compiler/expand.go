package compiler

import (
	"fmt"

	"github.com/arrangio/parser/ast"
)

func (c *compiler) expandInvocation(inv *ast.RuleInvocationStmt, depth int) error {
	if depth > 50 {
		return fmt.Errorf("exceeded maximum recursion depth while expanding rule_def %q", inv.Name)
	}

	def, ok := c.ruleDefs[inv.Name]
	if !ok {
		return fmt.Errorf("unknown rule_def %q", inv.Name)
	}

	env := make(map[string]any)

	// Bind parameters
	for i, param := range def.Params {
		if i < len(inv.Args) {
			env[param.Name] = inv.Args[i]
		} else if val, ok := inv.Params[param.Name]; ok {
			env[param.Name] = val
		} else if param.DefaultValue != nil {
			env[param.Name] = param.DefaultValue
		} else {
			return fmt.Errorf("missing required parameter %q for rule_def %q", param.Name, def.Name)
		}
	}

	// Expand body statements
	for _, stmt := range def.Body {
		switch s := stmt.(type) {
		case *ast.ConstraintStmt:
			rule, err := c.compileConstraint(s, env)
			if err != nil {
				return fmt.Errorf("in rule_def %q: %w", def.Name, err)
			}
			c.rules = append(c.rules, rule)

		case *ast.RuleInvocationStmt:
			// Resolve arguments in the current environment before recursive expansion
			resolvedInv := &ast.RuleInvocationStmt{
				Token:  s.Token,
				Name:   s.Name,
				Args:   make([]ast.Expr, len(s.Args)),
				Params: make(map[string]ast.Expr),
			}
			for idx, arg := range s.Args {
				if id, ok := arg.(*ast.IdentExpr); ok {
					if val, found := env[id.Name]; found {
						if exprVal, ok := val.(ast.Expr); ok {
							resolvedInv.Args[idx] = exprVal
							continue
						}
					}
				}
				resolvedInv.Args[idx] = arg
			}
			for k, v := range s.Params {
				if id, ok := v.(*ast.IdentExpr); ok {
					if val, found := env[id.Name]; found {
						if exprVal, ok := val.(ast.Expr); ok {
							resolvedInv.Params[k] = exprVal
							continue
						}
					}
				}
				resolvedInv.Params[k] = v
			}
			if err := c.expandInvocation(resolvedInv, depth+1); err != nil {
				return err
			}

		case *ast.RuleDecl:
			rule, err := c.compileRule(s)
			if err != nil {
				return fmt.Errorf("in rule_def %q: %w", def.Name, err)
			}
			c.rules = append(c.rules, rule)

		default:
			return fmt.Errorf("unsupported statement in rule_def %q: %T", def.Name, stmt)
		}
	}

	return nil
}
