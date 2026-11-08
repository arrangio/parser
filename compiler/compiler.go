package compiler

import (
	"fmt"

	"github.com/arrangio/core/entity"
	"github.com/arrangio/core/grid"
	"github.com/arrangio/core/rules"
	"github.com/arrangio/core/solver"
	"github.com/arrangio/core/zones"
	"github.com/arrangio/parser"
	"github.com/arrangio/parser/ast"
	"github.com/arrangio/parser/lexer"
)

// `CompiledScene` gets core structures ready for solver
type CompiledScene struct {
	Scene      *ast.SceneDecl
	State      *solver.State
	Entities   []*entity.Entity
	Zones      []*zones.Zone
	Rules      []rules.Rule
	EntityGrid *grid.Grid[*entity.Entity]
	ZoneGrid   *grid.Grid[*zones.Zone]
}

// `CompileSource` parses raw source and compiles it into a `CompiledScene`
func CompileSource(source string) (*CompiledScene, error) {
	p := parser.New(lexer.New(source))
	prog, err := p.ParseProgram()
	if err != nil {
		return nil, fmt.Errorf("parsing failed: %w", err)
	}
	return Compile(prog)
}

// `Compile` transforms an AST into a fully initialized `CompiledScene`
func Compile(prog *ast.Program) (*CompiledScene, error) {
	c := &compiler{
		reg:      newTagRegistry(),
		itemDefs: make(map[string]*ast.ItemDecl),
		itemTags: make(map[string]int),
		ruleDefs: make(map[string]*ast.RuleDefDecl),
		usedIDs:  make(map[uint64]bool),
		nextID:   1,
	}
	return c.compile(prog)
}

type compiler struct {
	reg      *tagRegistry
	itemDefs map[string]*ast.ItemDecl
	itemTags map[string]int
	ruleDefs map[string]*ast.RuleDefDecl
	entities []*entity.Entity
	zones    []*zones.Zone
	rules    []rules.Rule
	usedIDs  map[uint64]bool
	nextID   uint64
}

func (c *compiler) compile(prog *ast.Program) (*CompiledScene, error) {
	// register items
	for _, item := range prog.Items {
		c.itemDefs[item.Name] = item
		c.itemTags[item.Name] = c.reg.getOrRegister(item.Name)
	}

	// register rule_defs
	for _, rd := range prog.RuleDefs {
		c.ruleDefs[rd.Name] = rd
	}

	// instantiate entities
	for _, spawn := range prog.Spawns {
		if err := c.compileSpawn(spawn); err != nil {
			return nil, err
		}
	}

	// compile zones
	for i, z := range prog.Zones {
		c.zones = append(c.zones, c.compileZone(uint64(i+1), z))
	}

	// build spatial grids
	eg, zg, err := c.buildGrids(prog.Scene)
	if err != nil {
		return nil, err
	}
	for _, e := range c.entities {
		eg.Insert(e)
	}
	for _, z := range c.zones {
		zg.Insert(z)
	}

	// compile legacy rules
	for _, r := range prog.Rules {
		rule, err := c.compileRule(r)
		if err != nil {
			return nil, err
		}
		c.rules = append(c.rules, rule)
	}

	// compile mathematical constraints
	for _, cs := range prog.Constraints {
		rule, err := c.compileConstraint(cs, nil)
		if err != nil {
			return nil, err
		}
		c.rules = append(c.rules, rule)
	}

	// expand rule_def invocations
	for _, inv := range prog.Invocations {
		if err := c.expandInvocation(inv, 0); err != nil {
			return nil, err
		}
	}

	return &CompiledScene{
		Scene:      prog.Scene,
		State:      solver.NewState(eg, zg, c.entities, c.rules),
		Entities:   c.entities,
		Zones:      c.zones,
		Rules:      c.rules,
		EntityGrid: eg,
		ZoneGrid:   zg,
	}, nil
}
