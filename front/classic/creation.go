package classic

import (
	"gobox/front"
	"gobox/game"
)

// Creation is the creation sheet's parts from s (nil: none up) on a page
// of cols by rows cells: the sheet's text from the top left, the portrait
// against the right border and the combat icon under it, both at the
// whole scale the page leaves room for.
func Creation(s game.State, cols, rows int) *front.Creation {
	c := s.Creation
	if c == nil {
		return nil
	}
	d := c.Draft
	winRight, winBottom := max(cols, 40)-2, max(rows, 25)-3
	v := &front.Creation{Minimums: c.Minimums, Abandon: "Abandon this character?"}
	if a := s.Pending; a != nil && a.Kind == game.YesNo {
		v.Abandon = a.Prompt
	}
	v.Name = front.CellRect{X1: 1, Y1: 1, X2: max(1, len(d.Name)), Y2: 1}
	for f := range min(len(c.Fields), len(v.Fields)) {
		cf, fl := c.Fields[f], &v.Fields[f]
		fl.Name, fl.Cur = cf.Name, cf.Pick
		for i, o := range cf.Options {
			fl.Options = append(fl.Options, front.Option{Text: o, Fits: i < len(cf.Fits) && cf.Fits[i]})
		}
	}
	g := 0
	if sex := v.Fields[front.FieldSex]; sex.Cur >= 0 && sex.Cur < len(sex.Options) {
		g = len(sex.Options[sex.Cur].Text)
	}
	v.Fields[front.FieldSex].At = front.CellRect{X1: 1, Y1: 3, X2: g, Y2: 3}
	v.Fields[front.FieldRace].At = front.CellRect{X1: g + 2, Y1: 3, X2: g + 1 + len(d.Race), Y2: 3}
	v.Fields[front.FieldAlign].At = front.CellRect{X1: 1, Y1: 4, X2: len(d.Alignment), Y2: 4}
	v.Fields[front.FieldClass].At = front.CellRect{X1: 1, Y1: 5, X2: len(d.Class), Y2: 5}
	for i := range v.Stats {
		x1 := 6 // "STR 18", "(00)" after an exceptional strength
		if i == 0 && d.Stats["str"] == 18 && d.Str00 != 0 {
			x1 += len("(00)")
		}
		v.Stats[i] = front.CellRect{X1: 1, Y1: 7 + i, X2: x1, Y2: 7 + i}
	}
	sc := max(1, min((winRight-37)/11, (winBottom-3)/14)) // the combat stats reach column 37, the prompt row stays free
	left := winRight - 11*sc + 1                          // the portrait's 11x11 cells at 1x
	v.Head = front.CellRect{X1: left, Y1: 1, X2: winRight, Y2: 5 * sc}
	v.Body = front.CellRect{X1: left, Y1: 5*sc + 1, X2: winRight, Y2: 11 * sc}
	x, y := winRight-10*sc+1, 2+11*sc
	v.Icon, v.Scale = front.CellRect{X1: x, Y1: y, X2: x + 8*sc - 1, Y2: y + 3*sc - 1}, sc
	return v
}
