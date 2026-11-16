package ui

import (
	"image"
	"image/color"
	"slices"
	"strings"
	"testing"

	"gobox/game"
	"gobox/rules"
)

// br1Stats are the seven abilities a constructed state lists in the original's order (gdconsts.pas abilityAbbreviations).
var br1Stats = []game.StatDef{
	{Key: "str", Label: "STR", Name: "Strength", Value: 14}, {Key: "dex", Label: "DEX", Name: "Dexterity", Value: 13},
	{Key: "con", Label: "CON", Name: "Constitution", Value: 12}, {Key: "int", Label: "INT", Name: "Intelligence", Value: 11},
	{Key: "wis", Label: "WIS", Name: "Wisdom", Value: 10}, {Key: "cha", Label: "CHA", Name: "Charisma", Value: 9},
	{Key: "tch", Label: "TCH", Name: "Tech", Value: 15},
}

// The creation sheet lists the abilities the state lists, in its order: one row and one name each.
func TestCreationListsStateAbilities(t *testing.T) {
	c := &game.Creation{Draft: game.Member{Name: "Zed"}, Stats: br1Stats}
	v := Creation(game.ModeCreate, c, nil, 40, 25)
	if v == nil || !slices.Equal(v.StatNames, []string{"Strength", "Dexterity", "Constitution", "Intelligence", "Wisdom", "Charisma", "Tech"}) ||
		!slices.Equal(v.StatLabels, []string{"STR", "DEX", "CON", "INT", "WIS", "CHA", "TCH"}) || len(v.Stats) != 7 {
		t.Fatalf("%+v", v)
	}
	if got := minimums(map[string]int{"tch": 13}, br1Stats); got != "- - - - - - 13" {
		t.Errorf("minimums %q", got)
	}
}

// The sheet page draws every ability the page lists, the seventh on the row after the sixth.
func TestSheetListsSevenAbilities(t *testing.T) {
	sp := &game.SheetPage{Name: "Zed", Abilities: br1Stats, Creating: true}
	c := newCellPage(image.Pt(40, 25), [16]color.RGBA{}, nil)
	c.sheet(sp, nil, 0, nil, nil)
	var rows []string
	for y := 7; y < 14; y++ {
		var s string
		for _, r := range c.lines()[y] {
			s += r.Text
		}
		rows = append(rows, s)
	}
	for i, st := range br1Stats {
		if !strings.HasPrefix(rows[i], st.Name) {
			t.Errorf("row %d: %q, want the name %s", i, rows[i], st.Name)
		}
	}
}

// A state that offers one roll gets Roll alone: no Minimal on the bar, no Roll or Point Buy question for a new character.
func TestOnlyOfferedRolls(t *testing.T) {
	p, _ := draftOn(&game.Draft{Stats: br1Stats, Rolls: []rules.RollKind{rules.RollFull}})
	var words []string
	for _, w := range p.barWords() {
		words = append(words, w.label)
	}
	if !slices.Equal(words, []string{"Roll", "Keep", "Exit"}) {
		t.Errorf("bar %v", words)
	}
	p, _ = draftOn(&game.Draft{Stats: br1Stats})
	words = nil
	for _, w := range p.barWords() {
		words = append(words, w.label)
	}
	if !slices.Equal(words, []string{"Roll", "Minimal", "Keep", "Exit"}) {
		t.Errorf("a state listing no rolls offers every way: %v", words)
	}
}

// A state with points to spend gets the Points screen: the title says what is left, a row that wants more says what,
// and Enter and Subtract send Spend for the row.
func TestAllotmentScreen(t *testing.T) {
	d := &game.Draft{Allotments: []game.Allotment{{Name: "Career skill points", Left: 80, Max: 30, Rows: []game.AllotRow{
		{Name: "Treat Light", Value: 12, Bonus: 15, Total: 27, Open: true},
		{Name: "Treat Serious", Wants: []game.Want{{Name: "Treat Light", Level: 30}}},
	}}}}
	p, _ := draftOn(d)
	if ws := p.barWords(); !slices.ContainsFunc(ws, func(w footWord) bool { return w.label == "Points" }) {
		t.Fatalf("bar %v", ws)
	}
	p.barDo('P')()
	if p.scr.sub != scrAllot {
		t.Fatalf("sub %d", p.scr.sub)
	}
	title, items := p.screenItems()
	if !strings.Contains(title, "Career skill points: 80") || !strings.Contains(title, "at most 30") ||
		!strings.Contains(items[1].Text, "needs 30 Treat Light") || strings.Contains(items[0].Text, "needs") {
		t.Fatalf("%q %+v", title, items)
	}
	items[1].Do()
	p.scr.sel = 0
	_, items = p.screenItems() // the row last highlighted is the one Subtract takes from
	items[len(items)-2].Do()
	want := []Action{{Kind: ActSpend, N: 0, X: 1, Y: 1}, {Kind: ActSpend, N: 0, X: 0, Y: -1}}
	if !slices.Equal(p.inv.acts, want) {
		t.Errorf("sent %v, want %v", p.inv.acts, want)
	}
}
