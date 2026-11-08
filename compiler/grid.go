package compiler

import (
	"math"

	"github.com/arrangio/core/entity"
	"github.com/arrangio/core/geometry"
	"github.com/arrangio/core/grid"
	"github.com/arrangio/core/zones"
	"github.com/arrangio/parser/ast"
)

func (c *compiler) buildGrids(scene *ast.SceneDecl) (*grid.Grid[*entity.Entity], *grid.Grid[*zones.Zone], error) {
	var minPt, maxPt geometry.Point64

	if scene != nil {
		minPt = toPoint64(&scene.Min)
		maxPt = toPoint64(&scene.Max)
	} else {
		minPt = geometry.Point64{X: math.MaxInt64, Y: math.MaxInt64, Z: math.MaxInt64}
		maxPt = geometry.Point64{X: math.MinInt64, Y: math.MinInt64, Z: math.MinInt64}

		expand := func(bMin, bMax geometry.Point64) {
			minPt.X, minPt.Y, minPt.Z = min(minPt.X, bMin.X), min(minPt.Y, bMin.Y), min(minPt.Z, bMin.Z)
			maxPt.X, maxPt.Y, maxPt.Z = max(maxPt.X, bMax.X), max(maxPt.Y, bMax.Y), max(maxPt.Z, bMax.Z)
		}

		for _, z := range c.zones {
			expand(z.WorldBounds())
		}
		for _, e := range c.entities {
			expand(e.WorldBounds())
		}

		if minPt.X == math.MaxInt64 {
			minPt, maxPt = geometry.Point64{X: -1000, Y: -1000, Z: -1000}, geometry.Point64{X: 1000, Y: 1000, Z: 1000}
		} else {
			pad := int64(200)
			minPt.X -= pad
			minPt.Y -= pad
			minPt.Z -= pad
			maxPt.X += pad
			maxPt.Y += pad
			maxPt.Z += pad
		}
	}

	eg, err := grid.NewGrid[*entity.Entity](4, minPt.X, minPt.Y, minPt.Z, maxPt.X, maxPt.Y, maxPt.Z, 128)
	if err != nil {
		return nil, nil, err
	}
	zg, err := grid.NewGrid[*zones.Zone](4, minPt.X, minPt.Y, minPt.Z, maxPt.X, maxPt.Y, maxPt.Z, 128)
	return eg, zg, err
}
