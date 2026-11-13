package engine

import "gobox/game"

// refuseIf is a refusal with text when no, else nil.
func refuseIf(no bool, text string) error {
	if no {
		return &game.Refusal{Text: text}
	}
	return nil
}
