package pool

// The script's variables are typed fields (reference queue 1b-e): a Ref
// names a field of the running script's layout (data/assets/scripts/layout)
// and an index into it, never an address. A field the engine gives a meaning
// (engineFields) is read and written through it; every other field keeps
// its own cells, by name: vars for the game's saved state, locals for the
// scratch a fresh area clears. The save file's records are only the file
// format (savecodec.go).

// fieldCells are the values of the fields no engine state stands for: one
// slice per field, its Len cells (a text field grows to its string and the
// NUL after it).
type fieldCells map[string][]int64

// engineField is a layout field the engine gives a meaning, as ovr08's
// ReadGbvm/WriteGbvm (0E0C, 0C77) and their special cases (07BC, 0A74):
//   - get: the value read; ok false falls back to the field's own cells
//   - set: what a write does besides storing (stored) or instead of it
//   - stored: the field keeps its cells in vars as well (a word of the
//     original's game record that the engine also watches)
//   - text/setText: the field read or written as text (SelName)
type engineField struct {
	get     func(i int) (int64, bool)
	set     func(i int, v int64)
	stored  bool
	text    func() string
	setText func(s string)
}
