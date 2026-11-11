package goldbox

import (
	"gobox/core"
	"gobox/core/dnd"
)

// The Gold Box character record's race and alignment bytes (POR's order,
// the scripts' SelRace/SelAlignment words).
var (
	RecordRace = [8]core.Race{dnd.Monster, dnd.Dwarf, dnd.Elf, dnd.Gnome, dnd.HalfElf, dnd.Halfling, dnd.HalfOrc, dnd.Human}
	// RecordAlignment: LG, LN, LE, NG, N, NE, CG, CN, CE.
	RecordAlignment = [9]core.Alignment{dnd.LawfulGood, dnd.LawfulNeutral, dnd.LawfulEvil, dnd.NeutralGood, dnd.TrueNeutral,
		dnd.NeutralEvil, dnd.ChaoticGood, dnd.ChaoticNeutral, dnd.ChaoticEvil}
)

// SaveOrder is the saving throws in the Gold Box games' order: a record's
// save bytes and a script's save number (DAMAGE) index it.
var SaveOrder = [5]core.Save{dnd.SaveParalysis, dnd.SavePetrify, dnd.SaveRod, dnd.SaveBreath, dnd.SaveSpell}
