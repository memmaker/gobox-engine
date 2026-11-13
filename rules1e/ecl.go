package rules1e

import (
	"slices"
)

func (e *RuleState) Acting() *Character {
	if len(e.Party) == 0 {
		return &Character{}
	}
	return e.Party[e.Who]
}

func (e *RuleState) PartyIndex(c *Character) int { // ovr08:0756
	for i, p := range e.Party {
		if p == c {
			return i
		}
	}
	return len(e.Party)
}

// ItemSave = ovr08:1DF0 (c, pct): each item is lost with chance pct.
// extra: no reductions for heavy items (items weigh nothing), and the
// party inventory takes its chance with the first member.
// ponytail: the pool goes with party[0] only, so a hazard to another member alone spares it.
func (e *RuleState) ItemSave(c *Character, pct int) {
	list := c.Items
	if len(e.Party) > 0 && c == e.Party[0] {
		list = e.Carried(c)
	}
	for _, it := range slices.Clone(list) {
		if e.Dice(1, 100) <= pct {
			if it.Readied != 0 && it.Spells[2] > 0x7F { // a readied ring, cloak or gauntlets stops working when taken (the Pascal's fix)
				it.Readied = 0
				e.itemAffect(it, c, 1)
			}
			e.RemoveItem(c, it) // 00BA:1509
		}
	}
	e.ApplyItems(c) // AC and the rest without what was taken (the Pascal's fix: RecalcStats)
}

// MoveExtremes (was PartyHpRange) = ovr08:1FB1 (&hi, &lo): highest and lowest +0x11C in the
// party, doubled (special 0x27) or halved (0x2A) in byte arithmetic.
func (e *RuleState) MoveExtremes() (hi, lo int) {
	if len(e.Party) == 0 {
		return 0, 0
	}
	h8 := e.Party[0].Move
	l8 := h8
	for _, c := range e.Party {
		h := c.Move
		if e.FindAffect(c, AffHaste) != nil {
			h *= 2
		} else if e.FindAffect(c, AffSlow) != nil {
			h /= 2
		}
		if h > h8 {
			h8 = h
		}
		if h < l8 {
			l8 = h
		}
	}
	return int(h8), int(l8)
}
