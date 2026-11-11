package core

// WordFact is the ruleset's word for a character's fact Of, as its sheet
// shows it (por: RaceNames, GenderNames, AlignmentNames, ClassName); the
// value is a Text.
type WordFact struct{ Of PlainFact }

func (WordFact) fact() {}
