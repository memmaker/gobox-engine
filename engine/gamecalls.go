package engine

import "gobox/game/play"

// refuseIf is a refusal with text when no, else nil.
func refuseIf(no bool, text string) error {
	if no {
		return &play.Refusal{Text: text}
	}
	return nil
}
