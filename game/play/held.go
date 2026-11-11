package play

// Held keeps what a game (or gobox) read from its save folders (the roster, the slot
// list) while it waits for the player. State, View and every check they
// call run each frame and never touch the disk: they read through Held.
//
// Change starts a step that may write (Do, Load, LoadSlot, Save): it drops
// everything held, and until its done is called nothing is held, so the
// step's own reads see its own writes. Steps nest.
//
// Detect the bug class: a game's State with its save folder gone (renamed
// away after the first State) must be the same as before; grep for
// os.ReadDir / os.ReadFile / os.Stat reachable from State, allowed, common
// or MusicCue without Hold around it.
type Held struct {
	busy int
	vals map[any]any
}

// Change drops what is held and holds nothing until done is called.
func (h *Held) Change() (done func()) {
	h.busy++
	h.vals = nil
	return func() { h.busy--; h.vals = nil }
}

// Hold is load's value under key: read once while the game waits, again
// on every call inside a Change. key is a comparable value of a type of
// the caller's own (one per thing held), so two keys never collide.
// The value is shared: callers do not change it.
func Hold[T any](h *Held, key any, load func() T) T {
	if h.busy > 0 {
		return load()
	}
	if v, ok := h.vals[key]; ok {
		return v.(T)
	}
	v := load()
	if h.vals == nil {
		h.vals = map[any]any{}
	}
	h.vals[key] = v
	return v
}
