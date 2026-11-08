package compiler

import (
	"fmt"

	"github.com/arrangio/core/entity"
	"github.com/arrangio/core/geometry"
	"github.com/arrangio/core/zones"
	"github.com/arrangio/parser/ast"
)

func (c *compiler) compileSpawn(spawn *ast.SpawnDecl) error {
	def, ok := c.itemDefs[spawn.ItemName]
	if !ok {
		return fmt.Errorf("cannot spawn unknown item %q", spawn.ItemName)
	}

	count := max(1, spawn.Count)
	tagMask := c.reg.mask(append([]string{def.Name}, def.Tags...))
	shape := toBox(def.Size)
	facings := toFacings(def.Facings)
	anchor := toPoint64(spawn.Anchor)

	for i := 0; i < count; i++ {
		entID := c.nextID
		if spawn.ID != nil {
			entID = *spawn.ID + uint64(i)
		} else {
			for c.usedIDs[c.nextID] {
				c.nextID++
			}
			entID = c.nextID
			c.nextID++
		}

		if c.usedIDs[entID] {
			return fmt.Errorf("duplicate entity ID %d", entID)
		}
		c.usedIDs[entID] = true

		c.entities = append(c.entities, entity.NewEntity(
			entID, def.IsStatic, def.CanRotate, tagMask, nil, shape, facings, anchor,
		))
	}
	return nil
}

func (c *compiler) compileZone(id uint64, z *ast.ZoneDecl) *zones.Zone {
	return &zones.Zone{
		ID:        id,
		Name:      z.Name,
		Footprint: geometry.Footprint{Anchor: toPoint64(z.Anchor), Shape: toBox(z.Size)},
		Tags:      c.reg.mask(z.Tags),
	}
}

func toPoint64(v *ast.Vec3) geometry.Point64 {
	if v == nil {
		return geometry.Point64{}
	}
	return geometry.Point64{X: v.X, Y: v.Y, Z: v.Z}
}

func toBox(v ast.Vec3) geometry.Box {
	return geometry.Box{W: int16(v.X), H: int16(v.Y), D: int16(v.Z)}
}

func toFacings(facings []ast.Facing) *geometry.Facings {
	if len(facings) == 0 {
		return nil
	}
	fMap := make(geometry.Facings, len(facings))
	for _, f := range facings {
		fMap[f.Marker] = geometry.Point{X: int16(f.Vector.X), Y: int16(f.Vector.Y), Z: int16(f.Vector.Z)}
	}
	return &fMap
}
