package classic

import (
	"fmt"

	"gobox/front"
	"gobox/game"
)

// Training is the training hall's offer t to member n of party
// (nil: the hall is not open): what it becomes, or why it may not train;
// while a spell is being learned, the spells to choose from.
func Training(t *game.Training, party []game.Member, n int, purse map[string]int) *front.Training {
	if t == nil {
		return nil
	}
	if n < 0 || n >= len(party) {
		return &front.Training{Why: "No one to train"}
	}
	v := &front.Training{Name: party[n].Name, Price: t.Cost, Gold: purse[Coin]}
	if t.Learning {
		for _, sp := range t.Learn {
			v.Learn = append(v.Learn, front.BookSpell{ID: sp.ID, Name: sp.Name, Level: levelName(sp.Level)})
		}
		return v
	}
	v.Why, v.Becomes = t.Why[n], t.Becomes[n]
	return v
}

// levelName is a spell level's heading: "1st Level" .. "9th Level".
func levelName(l int) string {
	if l < 1 {
		return ""
	}
	suf := "th"
	if l < 4 {
		suf = [...]string{"", "st", "nd", "rd"}[l]
	}
	return fmt.Sprintf("%d%s Level", l, suf)
}
