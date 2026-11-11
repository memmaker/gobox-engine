package core

// WordIs asks the value of world word Word (Answer.Value, 0 or 1 for the
// flag words; Yes: the ruleset keeps it, for WordEffect the entry Index).
// The one form for the rules' record words (fold E-R6 design W1).
type WordIs struct {
	Word  WorldWord
	Index int
}

// SetWord stores Value in world word Word; a word the ruleset does not keep
// is refused (ErrUnknownWord).
type SetWord struct {
	Word  WorldWord
	Value int32
}

func (WordIs) check()   {}
func (SetWord) action() {}
