package classic

import (
	"gobox/front"
	"gobox/game"
)

// Camp is c (the game's) with what s gives in its place: whether the
// selected member may cast, its spell book with each spell's plan and its
// level's slots, and what resting does, and the bestiary. The scrolls
// and settings stay the game's. nil where c is.
func Camp(c *front.Camp, s game.State) *front.Camp {
	if c == nil || s.Selected < 0 || s.Selected >= len(s.Party) {
		return c
	}
	v := *c
	m := s.Party[s.Selected]
	v.CanCast, v.Healing, v.Studying, v.Icon, v.Book = m.CanCast, s.RestHeals, s.RestStudies, s.Allowed["SetIcon"], nil
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
	for _, b := range s.Bestiary {
		v.Bestiary = append(v.Bestiary, front.Beast{Name: b.Name, Danger: b.Danger, Text: b.Text, IconPix: pix(b.Icon), PicPix: pix(b.Art)})
	}
	return &v
}
