package rules1e

// PartyChars are the party's characters, as View.Party lists them.
func (e *RuleState) PartyChars() []*Character {
	var out []*Character
	for _, c := range e.Party {
		if IsParty(c) {
			out = append(out, c)
		}
	}
	return out
}
