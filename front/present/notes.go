package present

import (
	"gobox/front"
	"gobox/game"
)

// Note is v with the status line's toast n; v's own (a Refusal's, a shop's
// message) stays when n is empty.
func NoteView(v front.View, n game.Note) front.View {
	if n.Text == "" {
		return v
	}
	v.Note = front.Note{Text: n.Text, At: n.At}
	return v
}

// Results is v with the after-combat page r. v where r is nil.
func Results(v front.View, r *game.Results) front.View {
	if r == nil {
		return v
	}
	f := &front.Results{Shares: r.Shares, Outcome: r.Outcome, Exp: r.Exp, Taken: r.Taken, End: r.End,
		EndAt: front.CellRect{X1: r.EndAt[0], Y1: r.EndAt[1], X2: r.EndAt[2], Y2: r.EndAt[3]}}
	for _, h := range r.Harmed {
		f.Harmed = append(f.Harmed, front.Harm{Name: h.Name, What: h.What})
	}
	v.Results = f
	return v
}
