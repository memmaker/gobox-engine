package rules

// World is the engine's own world words: state the engine and the scripts
// read and write, which no ruleset reads. The engine keeps one (it was the 1e
// records' cells) and the canonical save holds it; it lives here, not in
// adventure or engine, because the original-save reader of a ruleset fills it.
// Each cell is 16 bits wide, as the records' words were.
type World struct {
	AreaKind   uint16 // the area's ECL block number; also the block NewECL left (POR keeps it here, as the original's DS:5D7D copy)
	PrevX      uint16 // the square before the step (the step sound)
	PrevY      uint16
	RiverOpen  uint16 // nonzero: the river can be crossed
	Search     uint16 // bit 0 sweeping, bit 1 search this square
	HoldMove   uint16 // >= 0xFF: the next move is skipped
	LeftMap    uint16 // the step left the map
	Geo        uint16 // the GEO block of the 3D map
	WallSplitA uint16 // with WallSplitB set, opLoadWalls loads into slots 1 and 3
	WallSplitB uint16
	NoCoords   uint16 // nonzero: no x,y on the status line, no area map
	MsgDelay   uint16 // text delay, the save file's copy
	ViewFlags  uint16 // the save file's copy of the view flags (bit 1 and bit 0)
	MaxDist    uint16 // opSetupMonster's distance
	Portrait   uint16 // HEAD/BODY pair shown instead of a picture, 0xFF none
	Temple     uint16 // 1: opCombat opens the temple
	Shop       uint16 // 1: opCombat opens the shop
	SymbolSet  uint16 // the 8X8D<n> symbol set
}

// Cells are the world's cells by their script field names (a layout names a
// field so when it has it).
func (w *World) Cells() map[string]*uint16 {
	return map[string]*uint16{
		"AreaKind": &w.AreaKind, "PrevX": &w.PrevX, "PrevY": &w.PrevY, "RiverOpen": &w.RiverOpen,
		"Search": &w.Search, "HoldMove": &w.HoldMove, "LeftMap": &w.LeftMap, "Geo": &w.Geo,
		"WallSplitA": &w.WallSplitA, "WallSplitB": &w.WallSplitB, "NoCoords": &w.NoCoords,
		"MsgDelay": &w.MsgDelay, "ViewFlags": &w.ViewFlags, "MaxDist": &w.MaxDist,
		"Portrait": &w.Portrait, "Temple": &w.Temple, "Shop": &w.Shop, "SymbolSet": &w.SymbolSet,
	}
}
