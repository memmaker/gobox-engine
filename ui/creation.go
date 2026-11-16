package ui

import (
	"gobox/game"
	"slices"
	"strconv"
	"strings"
)

// statLongCol is the column the creation sheet's ability values start at, after the names written out.
const statLongCol = 14

// defaultStat lists the six abilities a state that lists none shows, in the 1e order.
var defaultStat = [...]game.StatDef{{Key: "str", Label: "STR", Name: "Strength"}, {Key: "int", Label: "INT", Name: "Intelligence"},
	{Key: "wis", Label: "WIS", Name: "Wisdom"}, {Key: "dex", Label: "DEX", Name: "Dexterity"},
	{Key: "con", Label: "CON", Name: "Constitution"}, {Key: "cha", Label: "CHA", Name: "Charisma"}}

// defaultStats is the default six with the values by key.
func defaultStats(values map[string]int) []game.StatDef {
	out := slices.Clone(defaultStat[:])
	for i := range out {
		out[i].Value = values[out[i].Key]
	}
	return out
}

// pageStats is the abilities page sp lists: its own, else the six of the old engines (Stats, Str00, Mods), else the
// default six.
func pageStats(sp *game.SheetPage) []game.StatDef {
	if len(sp.Abilities) > 0 {
		return sp.Abilities
	}
	out := defaultStats(nil)
	for i := range out {
		out[i].Value, out[i].Mod = sp.Stats[i], sp.Mods[i]
	}
	out[0].Extra = sp.Str00
	return out
}

// statsOf is the abilities c's draft lists: its own, else the default six.
func statsOf(c *game.Creation) []game.StatDef {
	if len(c.Stats) > 0 {
		return c.Stats
	}
	d := defaultStats(c.Draft.Stats)
	if d[0].Value == 18 && c.Draft.Str00 != 0 {
		d[0].Extra = "(00)"
	}
	return d
}

// Creation is the creation sheet's parts from s (nil: none up) on a page
// of cols by rows cells: the sheet's text from the top left, the portrait
// against the right border and the combat icon under it, both at the
// whole scale the page leaves room for.
func Creation(mode game.Mode, c *game.Creation, pending *game.Ask, cols, rows int) *FrontCreation {
	if !Drafting(mode, c) {
		return nil
	}
	d := c.Draft
	winRight, winBottom := max(cols, 40)-2, max(rows, 25)-3
	v := &FrontCreation{Minimums: minimums(c.Minimums, statsOf(c)), Abandon: "Abandon this character?"}
	if a := pending; a != nil && a.Kind == game.YesNo {
		v.Abandon = a.Prompt
	}
	v.Name = CellRect{X1: 1, Y1: 1, X2: max(1, len(d.Name)), Y2: 1}
	main := min(len(c.Fields), 4) // the four sheet fields, then the power picks
	if rowFields(c) {
		main = len(c.Fields)
	}
	for i, cf := range c.Fields {
		fl := Field{Name: cf.Name, Cur: cf.Pick}
		for i, o := range cf.Options {
			fl.Options = append(fl.Options, Option{Text: o, Fits: i < len(cf.Fits) && cf.Fits[i]})
		}
		if i < main {
			v.Fields = append(v.Fields, fl)
		} else {
			v.Powers = append(v.Powers, fl)
		}
	}
	if rowFields(c) { // one field per row, as draftRows draws them
		for i := range v.Fields {
			v.Fields[i].At = CellRect{X1: 1, Y1: fieldRow + i, X2: max(1, len(picked(c.Fields[i]))), Y2: fieldRow + i}
		}
	} else {
		g := 0
		if len(c.Fields) > 0 {
			g = len(picked(c.Fields[FieldSex]))
		}
		for i := range v.Fields {
			v.Fields[i].At = []CellRect{
				{X1: 1, Y1: 3, X2: g, Y2: 3},
				{X1: g + 2, Y1: 3, X2: g + 1 + len(d.Race), Y2: 3},
				{X1: 1, Y1: 4, X2: len(d.Alignment), Y2: 4},
				{X1: 1, Y1: 5, X2: len(d.Class), Y2: 5},
			}[i]
		}
	}
	for i, st := range statsOf(c) { // the abilities written out, their values from statLongCol, an extra (an exceptional strength) after
		v.StatNames = append(v.StatNames, st.Name)
		v.StatLabels = append(v.StatLabels, st.Label)
		v.Stats = append(v.Stats, CellRect{X1: 1, Y1: 7 + i, X2: statLongCol + 1 + len(st.Extra), Y2: 7 + i})
	}
	sc := max(1, min((winRight-statLongCol-7)/11, (winBottom-3)/14)) // the abilities reach past their value column, the bottom rows stay free
	left := winRight - 11*sc + 1                                     // the portrait's 11x11 cells at 1x
	v.Head = CellRect{X1: left, Y1: 1, X2: winRight, Y2: 5 * sc}
	v.Body = CellRect{X1: left, Y1: 5*sc + 1, X2: winRight, Y2: 11 * sc}
	x, y := winRight-10*sc+1, 2+11*sc
	v.Icon, v.Scale = CellRect{X1: x, Y1: y, X2: x + 8*sc - 1, Y2: y + 3*sc - 1}, sc
	return v
}

// Drafting says whether the creation (or Modify) sheet is up in s: a
// Creation in ModeCreate. A game may state its Creation (the new
// character's first step, the roster) on the party menu too; that is no sheet.
func Drafting(mode game.Mode, c *game.Creation) bool { return c != nil && mode == game.ModeCreate }

// rowFields says whether c's sheet puts its fields one per row in the game's
// order, with no power picks: a sheet without an alignment (the Buck Rogers
// sheets, BR1 O0D2:00E3, BR2 O07E:00EA). A sheet with one has the four fields
// at their places (sex and race, alignment, class), the power picks after.
func rowFields(c *game.Creation) bool {
	return !slices.ContainsFunc(c.Fields, func(f game.Field) bool { return f.Alignment })
}

// fieldRow is the first row of the adnd2-br sheet's fields.
const fieldRow = 3

// picked is f's picked option ("": none).
func picked(f game.Field) string {
	if f.Pick >= 0 && f.Pick < len(f.Options) {
		return f.Options[f.Pick]
	}
	return ""
}

// minimums is the abilities' minimums as Minimal shows them, in the sheet's order ("-": none).
func minimums(m map[string]int, stats []game.StatDef) string {
	var f []string
	for _, st := range stats {
		if v, ok := m[st.Key]; ok {
			f = append(f, strconv.Itoa(v))
		} else {
			f = append(f, "-")
		}
	}
	return strings.Join(f, " ")
}

// Coin is game.Data.Coin: the money Purse, Sale and Pooled count in.
var Coin = game.Gold
