package core

import "errors"

// ErrNotNow is a ruleset's refusal of a call the game does not allow now
// (por: the ruleset's rules1e.ErrNotNow is this value).
var ErrNotNow = errors.New("not now")

// ErrUnknownWord refuses a SetWord of a word the ruleset does not keep.
var ErrUnknownWord = errors.New("unknown world word")
