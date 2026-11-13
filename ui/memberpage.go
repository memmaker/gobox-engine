package ui

import (
	"cmp"
	"strconv"
	"strings"

	"gobox/game"
)

// memberPage is m's sheet page from the Member alone, for rules that publish none.
func memberPage(m game.Member) *game.SheetPage {
	p := &game.SheetPage{Name: m.Name, Race: m.Race, Alignment: m.Alignment, Class: m.Class, Status: m.Status,
		AC: m.AC, HP: m.HP, MaxHP: m.MaxHP, THAC0: m.THAC0, Move: m.Move, Age: m.Age, Exp: int(m.Exp), Next: m.NextExp,
		Levels: m.LevelText}
	if p.Levels == "" { // "3/2": every class's level
		var l []string
		for _, n := range m.Levels {
			l = append(l, strconv.Itoa(n))
		}
		p.Levels = strings.Join(l, "/")
	}
	p.Class = cmp.Or(m.ClassText, m.Class)
	for i, k := range []string{"str", "int", "wis", "dex", "con", "cha"} {
		p.Stats[i] = m.Stats[k]
	}
	return p
}
