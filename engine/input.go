package engine

// DisplayMessageWait = 0709:0D68 (str, colour): message on row 24 for
// msgDelay*225 ms, then cleared.
func (e *Engine) DisplayMessageWait(colour int, s string) {
	e.say(s) // extra: a flash at the bottom (the original's row 24 is never drawn on the remastered 3D view); nothing to confirm
}
