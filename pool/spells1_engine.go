package pool

// ovr23 (spells), first part (0x0000..0x100F) and the tail (0x2540..0x37B5).
// The spell effects between are in spells2.go. The spell record is
// DS:31FA + 16*spell (see spellRec): [0] class (casterLevel), [1] list
// group, [2]/[3] duration base/per level, [4]/[5] amount base/per level,
// [6] minimum 1 flag, [7] targeting kind, [8] save mode, [9] save type,
// [0xA] affect type.

import (
	"fmt"
	"gobox/core"
	"gobox/core/goldbox"
)

// buildSpellList = ovr23:059C (kind): fills the menu list with the
// rules' SpellMenuFact of the selected member; kinds as in SpellList.
// Each scroll line notes its scroll (DS:477F; kinds 3 and 6 start the
// list afresh, kind 2 goes on from the last). True when there is at
// least one line.
func (e *Engine) buildSpellList(kind int) bool {
	e.spellList = nil
	e.spellIDs = nil
	if kind == 3 || kind == 6 {
		e.spellItems = [0x31]core.ItemHandle{}
		e.spellItemCount = 0
	}
	for _, l := range e.fact(e.selected(), core.SpellMenuFact{Kind: kind}).(core.SpellMenu) {
		if l.Header {
			text := goldbox.Cut(l.Text, textCols)
			if l.Slots >= 0 { // extra: the level headers carry the slots planned
				text = goldbox.Cut(fmt.Sprintf("%s  %d/%d", text, l.Planned, l.Slots), textCols)
			}
			e.spellList = append(e.spellList, &ListNode{Text: text, Flag: 1})
			continue
		}
		prefix := "  "
		if l.Memo { // bit 7: to be memorised
			prefix = " *"
		}
		text := goldbox.Cut(prefix+l.Text, textCols)
		if l.Count > 0 { // extra: how often it is memorized on a rest (memorizeMenu)
			text = goldbox.Cut(fmt.Sprintf("%-30s x%d", text, l.Count), 0x25)
		}
		e.spellList = append(e.spellList, &ListNode{Text: text, Spell: l.Number})
		e.spellIDs = append(e.spellIDs, byte(l.Number))
		if l.Item != core.NoItem {
			if e.spellItemCount < len(e.spellItems) {
				e.spellItems[e.spellItemCount] = l.Item
			}
			e.spellItemCount++
		}
	}
	return len(e.spellList) > 0
}
