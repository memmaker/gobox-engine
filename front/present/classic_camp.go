package present

import (
	"gobox/front"
	"gobox/game"
)

// Camp is c (the game's) with what is given in its place: whether the
// selected member may cast, its spell book with each spell's plan and its
// level's slots, and what resting does, and the bestiary. The scrolls
// and settings stay the game's. nil where c is.
func Camp(c *front.Camp, party []game.Member, sel int, heals, studies, noRest bool, beasts []game.Beast) *front.Camp {
	if c == nil || sel < 0 || sel >= len(party) {
		return c
	}
	v := *c
	m := party[sel]
	v.CanCast, v.Healing, v.Studying, v.Book = m.CanCast, heals, studies, nil
	v.NoRest = noRest // camp with no AdvanceTimeBy
	planned := map[string]int{}
	for _, sp := range m.Spells {
		if !sp.Memorized {
			planned[sp.Group] += sp.Planned
		}
	}
	for _, sp := range m.Spells {
		if !sp.Memorized {
			v.Book = append(v.Book, front.BookSpell{ID: sp.ID, Name: sp.Name, Level: levelName(sp.Level), Count: sp.Planned,
				Planned: planned[sp.Group], Slots: m.SlotsMax[sp.Group]})
		}
	}
	v.Bestiary = nil
	for _, b := range beasts {
		v.Bestiary = append(v.Bestiary, front.Beast{Name: b.Name, Danger: b.Danger, Text: b.Text, IconPix: pix(b.Icon), PicPix: pix(b.Art)})
	}
	return &v
}
