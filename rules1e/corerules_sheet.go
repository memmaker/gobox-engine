package rules1e

import "gobox/rules"

// The creation sheet's character (core/creation.go): the draft drawn as a
// record, seated after the party while the sheet is up.

// resolveSheet answers the Actions of the creation sheet.
func (r *Rules) resolveSheet(a rules.Action) (rules.Outcome, bool, error) {
	s := r.S
	switch a := a.(type) {
	case rules.SheetDraft:
		return rules.Outcome{Created: r.ID(r.Creation().CharOf(a.Draft))}, true, nil
	case rules.SeatTemp:
		c := r.chars[a.Who]
		if c == nil {
			return rules.Outcome{}, true, errNoChar
		}
		s.Party = append(s.Party, c)
	case rules.TrimParty:
		s.Party = s.Party[:a.Size]
	case rules.ReplaceMember:
		c := r.chars[a.Who]
		if c == nil {
			return rules.Outcome{}, true, errNoChar
		}
		s.Party[a.At] = c
	case rules.HoldCopy:
		c := r.chars[a.Who]
		if c == nil {
			return rules.Outcome{}, true, errNoChar
		}
		if r.held == nil {
			r.held = map[rules.CharID]Character{}
		}
		r.held[a.Who] = *c
	case rules.RestoreCopy:
		c, was := r.chars[a.Who], r.held[a.Who]
		if c == nil {
			return rules.Outcome{}, true, errNoChar
		}
		*c = was
		delete(r.held, a.Who)
	case rules.SetMinimums:
		s.SetMinStats(a.Text)
	default:
		return rules.Outcome{}, false, nil
	}
	return rules.Outcome{}, true, nil
}
