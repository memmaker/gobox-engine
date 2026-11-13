package engine

import (
	"fmt"
	"maps"

	"gobox/ecl/script"
)

// codec is a game's encoding of a canonical field (script.Field.Encoding): the engine's
// bindings always see the canonical value and element; the script sees the game's.
type codec struct {
	toGame       func(e *Engine, canonical int64) int64 // a read: the canonical value as the game keeps it
	toCanon      func(e *Engine, game int64) int64      // a write or an imported save: the game's value as the canonical one
	element      func(i int) int                        // the game's element i as the canonical element; nil: the same
	ignoreWrites bool                                   // the game's write only stores the cell; the binding never sees it
}

// The canonical SelStatus values (POR's, scriptvars_engine.go): a read is Okay or Out.
const (
	canonStatusDied   = 0    // a write: the member died
	canonStatusOkay   = 1    // active; a write revives the member
	canonStatusStore  = 2    // a write that only stores
	canonStatusOut    = 0x80 // inactive (a write >= 0x80 but Stoned)
	canonStatusStoned = 0x87 // a write: stoned
)

// codecs are every game's encodings by name; each game's file adds its own.
var codecs = joinCodecs(sotsbCodecs, cotabCodecs, podCodecs)

func joinCodecs(sets ...map[script.Encoding]codec) map[script.Encoding]codec {
	all := map[script.Encoding]codec{}
	for _, s := range sets {
		for n := range s {
			if _, ok := all[n]; ok {
				panic("pool: encoding " + string(n) + " twice")
			}
		}
		maps.Copy(all, s)
	}
	return all
}

// codecOf is f's codec; ok false for the canonical encoding.
func codecOf(f *script.Field) (codec, bool) {
	if f.Encoding == "" {
		return codec{}, false
	}
	c, ok := codecs[f.Encoding]
	if !ok {
		panic(fmt.Sprintf("script: %s: no encoding %s", f.Name, f.Encoding)) // the game's layout names an encoding pool lacks
	}
	return c, true
}

// writesIgnored reports whether f's game ignores a script's write of it.
func writesIgnored(f *script.Field) bool {
	c, ok := codecOf(f)
	return ok && c.ignoreWrites
}

// canonElement is f's element i as its binding numbers it.
func canonElement(f *script.Field, i int) int {
	if c, ok := codecOf(f); ok && c.element != nil {
		return c.element(i)
	}
	return i
}

// gameValue is the canonical value v as f's game keeps it.
func (e *Engine) gameValue(f *script.Field, v int64) int64 {
	if c, ok := codecOf(f); ok && c.toGame != nil {
		return c.toGame(e, v)
	}
	return v
}

// canonValue is the game's value v of f as the canonical one.
func (e *Engine) canonValue(f *script.Field, v int64) int64 {
	if c, ok := codecOf(f); ok && c.toCanon != nil {
		return c.toCanon(e, v)
	}
	return v
}
