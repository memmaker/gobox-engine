package present

import (
	"gobox/front"
	"gobox/game"
	"gobox/game/play"
	"slices"
	"strconv"
	"strings"
)

// Creation is the creation sheet's parts from s (nil: none up) on a page
// of cols by rows cells: the sheet's text from the top left, the portrait
// against the right border and the combat icon under it, both at the
// whole scale the page leaves room for.
func Creation(mode game.Mode, c *game.Creation, pending *play.Ask, cols, rows int) *front.Creation {
	if !Drafting(mode, c) {
		return nil
	}
	d := c.Draft
	winRight, winBottom := max(cols, 40)-2, max(rows, 25)-3
	v := &front.Creation{Minimums: minimums(c.Minimums), Abandon: "Abandon this character?"}
	if a := pending; a != nil && a.Kind == game.YesNo {
		v.Abandon = a.Prompt
	}
	v.Name = front.CellRect{X1: 1, Y1: 1, X2: max(1, len(d.Name)), Y2: 1}
	main := min(len(c.Fields), 4) // the four sheet fields, then the power picks
	if rowFields(c) {
		main = len(c.Fields)
	}
	for i, cf := range c.Fields {
		fl := front.Field{Name: cf.Name, Cur: cf.Pick}
		for i, o := range cf.Options {
			fl.Options = append(fl.Options, front.Option{Text: o, Fits: i < len(cf.Fits) && cf.Fits[i]})
		}
		if i < main {
			v.Fields = append(v.Fields, fl)
		} else {
			v.Powers = append(v.Powers, fl)
		}
	}
	if rowFields(c) { // one field per row, as draftRows draws them
		for i := range v.Fields {
			v.Fields[i].At = front.CellRect{X1: 1, Y1: fieldRow + i, X2: max(1, len(picked(c.Fields[i]))), Y2: fieldRow + i}
		}
	} else {
		g := 0
		if len(c.Fields) > 0 {
			g = len(picked(c.Fields[front.FieldSex]))
		}
		for i := range v.Fields {
			v.Fields[i].At = []front.CellRect{
				{X1: 1, Y1: 3, X2: g, Y2: 3},
				{X1: g + 2, Y1: 3, X2: g + 1 + len(d.Race), Y2: 3},
				{X1: 1, Y1: 4, X2: len(d.Alignment), Y2: 4},
				{X1: 1, Y1: 5, X2: len(d.Class), Y2: 5},
			}[i]
		}
	}
	if c.Stats != nil {
		for i, st := range c.Stats {
			v.StatNames = append(v.StatNames, st.Label)
			x2 := front.StatCol(st.Label) + 1 + len(st.Extra)
			v.Stats = append(v.Stats, front.CellRect{X1: 1, Y1: 7 + i, X2: x2, Y2: 7 + i})
		}
	} else {
		v.StatNames = []string{"STR", "INT", "WIS", "DEX", "CON", "CHA"}
		for i := range v.StatNames {
			x1 := 6 // "STR 18", "(00)" after an exceptional strength
			if i == 0 && d.Stats["str"] == 18 && d.Str00 != 0 {
				x1 += len("(00)")
			}
			v.Stats = append(v.Stats, front.CellRect{X1: 1, Y1: 7 + i, X2: x1, Y2: 7 + i})
		}
	}
	sc := max(1, min((winRight-37)/11, (winBottom-3)/14)) // the combat stats reach column 37, the prompt row stays free
	left := winRight - 11*sc + 1                          // the portrait's 11x11 cells at 1x
	v.Head = front.CellRect{X1: left, Y1: 1, X2: winRight, Y2: 5 * sc}
	v.Body = front.CellRect{X1: left, Y1: 5*sc + 1, X2: winRight, Y2: 11 * sc}
	x, y := winRight-10*sc+1, 2+11*sc
	v.Icon, v.Scale = front.CellRect{X1: x, Y1: y, X2: x + 8*sc - 1, Y2: y + 3*sc - 1}, sc
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
	return !slices.ContainsFunc(c.Fields, func(f game.Field) bool { return strings.EqualFold(f.Name, "alignment") })
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

// minimums is the six abilities' minimums as Minimal shows them, STR
// first ("-": none).
func minimums(m map[string]int) string {
	var f []string
	for _, k := range []string{"str", "int", "wis", "dex", "con", "cha"} {
		if v, ok := m[k]; ok {
			f = append(f, strconv.Itoa(v))
		} else {
			f = append(f, "-")
		}
	}
	return strings.Join(f, " ")
}

// Coin is game.Data.Coin: the money Purse, Sale and Pooled count in.
var Coin = game.Gold
