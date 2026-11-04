package classic

import (
	"gobox/front"
	"gobox/game"
)

// Quests is the quest journal s holds (nil: none).
func Quests(s game.State) []front.Quest {
	var out []front.Quest
	for _, q := range s.Quests {
		out = append(out, front.Quest{Name: q.ID, Giver: q.Giver, Place: q.Place, Text: q.Brief, Done: q.Done})
	}
	return out
}

// Lore is the journal's notes s holds (nil: none).
func Lore(s game.State) []front.Lore {
	var out []front.Lore
	for _, l := range s.Lore {
		out = append(out, front.Lore{Title: l.Title, Place: l.Place, Text: l.Text, ID: l.ID})
	}
	return out
}
