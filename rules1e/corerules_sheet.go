package rules1e

import "gobox/core"

// The creation sheet's character (core/creation.go): the draft drawn as a
// record, seated after the party while the sheet is up.

// resolveSheet answers the Actions of the creation sheet.
func (r *Rules) resolveSheet(a core.Action) (core.Outcome, bool, error) {
	s := r.S
	switch a := a.(type) {
	case core.SheetDraft:
		return core.Outcome{Created: r.ID(r.Creation().CharOf(a.Draft))}, true, nil
	case core.SeatTemp:
		c := r.chars[a.Who]
		if c == nil {
			return core.Outcome{}, true, errNoChar
		}
		s.Party = append(s.Party, c)
	case core.TrimParty:
		s.Party = s.Party[:a.Size]
	case core.ReplaceMember:
		c := r.chars[a.Who]
		if c == nil {
			return core.Outcome{}, true, errNoChar
		}
		s.Party[a.At] = c
	case core.HoldCopy:
		c := r.chars[a.Who]
		if c == nil {
			return core.Outcome{}, true, errNoChar
		}
		if r.held == nil {
			r.held = map[core.CharID]Character{}
		}
		r.held[a.Who] = *c
	case core.RestoreCopy:
		c, was := r.chars[a.Who], r.held[a.Who]
		if c == nil {
			return core.Outcome{}, true, errNoChar
		}
		*c = was
		delete(r.held, a.Who)
	case core.SetMinimums:
		s.SetMinStats(a.Text)
	default:
		return core.Outcome{}, false, nil
	}
	return core.Outcome{}, true, nil
}
