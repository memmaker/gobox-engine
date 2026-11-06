package classic

import (
	"gobox/front"
	"gobox/game"
)

// Note is v with the status line's toast s.Note holds; v's own (a
// Refusal's, a shop's message) stays when the State has none.
func Note(v front.View, s game.State) front.View {
	if s.Note.Text == "" {
		return v
	}
	v.Note = front.Note{Text: s.Note.Text, At: s.Note.At}
	return v
}

// Results is v with the after-combat page s.Results holds. v where s has none.
func Results(v front.View, s game.State) front.View {
	r := s.Results
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
