package ui

// A hard UI rule: the input that changes the screen never acts on the
// screen it opens. When the screen changes, every key, mouse button, pad
// button and stick direction still held is stale: ignored until it is let
// go and pressed again (a held click does not drag on the new screen, a
// held arrow does not repeat into a new menu).

// Latch tracks the held inputs that are stale; an input is any comparable
// key (a key, a mouse button, a pad button, a stick direction).
type Latch struct {
	screen string
	stale  map[any]bool
}

// See starts a frame on screen (the game's name for the screens up, built
// with Screens) with the inputs held now: a stale input no longer held is
// fresh again; when the screen differs from the last one, every input held
// turns stale.
func (l *Latch) See(screen string, held []any) {
	now := map[any]bool{}
	for _, h := range held {
		now[h] = true
	}
	if screen != l.screen {
		l.screen, l.stale = screen, now
		return
	}
	for k := range l.stale {
		if !now[k] {
			delete(l.stale, k)
		}
	}
}

// Down is whether input in, pressed or not, counts as held: a stale one
// never does.
func (l *Latch) Down(in any, pressed bool) bool { return pressed && !l.stale[in] }
