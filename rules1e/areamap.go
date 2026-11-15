package rules1e

import "gobox/rules"

// The area map (wilderness / city overview): ovr28 and ovr27:0000.

// MapTerrain is the area map's terrain at map square (x, y) (DS:3648).
func (e *RuleState) MapTerrain(x, y int) Terrain { return TerrainMap[y][x] }

// AreaTile is the area map's tile at map square (tx,ty): the terrain, or a
// landmark of the DS:0C30 table (x, y, icon) while it shows.
func (e *RuleState) AreaTile(tx, ty int, m rules.AreaMarks) byte {
	icon := TerrainMap[int8(byte(ty))][int8(byte(tx))] - 1
	for k := 1; k <= 9; k++ {
		ent := areaMapMarks[k]
		if int(int8(byte(tx))) != int(ent[0]) || int(int8(byte(ty))) != int(ent[1]) {
			continue
		}
		set := false
		switch k {
		case 3:
			set = m.Mark3 != 0xFF
		case 4:
			set = m.Mark4 != 0 || m.Flags4&8 != 0 || m.Mark4Off == 0
		case 5:
			set = m.Flags57&2 == 0
		case 6:
			set = true
		case 7:
			set = m.Flags57&1 == 0
		case 8, 9:
			set = m.Boss > 0xFD
		}
		if set {
			icon = Terrain(ent[2] - 1)
		}
	}
	return byte(icon)
}
