package classic

import (
	"gobox/front"
	"gobox/game"
)

// Quests is the quest journal from qs (nil: none).
func Quests(qs []game.Quest) []front.Quest {
	var out []front.Quest
	for _, q := range qs {
		out = append(out, front.Quest{Name: q.ID, Giver: q.Giver, Place: q.Place, Text: q.Brief, Done: q.Done})
	}
	return out
}

// Lore is the journal's notes from ls (nil: none).
func Lore(ls []game.Lore) []front.Lore {
	var out []front.Lore
	for _, l := range ls {
		out = append(out, front.Lore{Title: l.Title, Place: l.Place, Text: l.Text, ID: l.ID})
	}
	return out
}
