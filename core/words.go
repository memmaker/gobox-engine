package core

// WordValue asks the value of a world word (Value); Yes: the ruleset keeps it.
// The one form for the game and area record words (fold E-R6 design W1).
type WordValue struct{ Word WorldWord }

// SetWord stores Value in a world word; an unknown word is refused
// (ErrUnknownWord).
type SetWord struct {
	Word  WorldWord
	Value int32
}

func (WordValue) check() {}
func (SetWord) action()  {}
