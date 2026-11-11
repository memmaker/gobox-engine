package core

import "errors"

// ErrNotNow is a ruleset's refusal of a call the game does not allow now
// (por: the ruleset's rules1e.ErrNotNow is this value).
var ErrNotNow = errors.New("not now")
