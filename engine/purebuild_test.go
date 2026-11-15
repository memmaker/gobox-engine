package engine

import (
	"fmt"
	"gobox/rules"
	"strings"
	"testing"
)

// printState is everything persistent the build could touch: the ruleset's
// save, the engine's save fields, and the extras no save covers.
func (e *Engine) printState(t *testing.T) string {
	t.Helper()
	blob, err := e.ruleset().Save()
	if err != nil {
		t.Fatal(err)
	}
	var list []string
	for _, n := range e.spellList {
		list = append(list, fmt.Sprintf("%+v", *n))
	}
	return fmt.Sprintf("%x|%v|%v|sel%d|loot%v|fx%d|%v|%v|%v|%v|%v|%v|%+v|%v|%v", blob, e.vars, e.locals, e.sel, e.lootSeen, len(e.fx),
		list, e.spellIDs, e.ext.labels, e.ext.battleLog, e.ext.panel, e.ext.note, e.snap.strike, e.ext.noteAt, e.events)
}

// TestBuildReadsOnly (REFACTOR G2b): build changes nothing persistent.
func TestBuildReadsOnly(t *testing.T) {
	scenes := []string{"hall", "shop", "temple", "camp", "castfight", "loot", "inventory", "sheet", "overland", "parlay"}
	open := map[string]func() *Engine{
		"combat":       func() *Engine { return devFirstTurn(t, "combat", false) },
		"castfight":    func() *Engine { return devFirstTurn(t, "castfight", false) },
		"reservefight": func() *Engine { return devFirstTurnOn(t, advEngine(t, advEnvPOD), "reservefight", false) },
	}
	for _, s := range scenes {
		if open[s] == nil {
			open[s] = func() *Engine { return devPlay(t, s) }
		}
	}
	for name, o := range open {
		e := o()
		e.emit(rules.Narrated{Text: "queued"}) // what an action queued: only the action applies it
		before := e.printState(t)
		for range 3 {
			e.build()
		}
		if after := e.printState(t); after != before {
			t.Errorf("%s: build changed state:\n%s", name, firstDiff(before, after))
		}
		if fx := e.ruleset().(rules.Pending).TakePending(); len(fx) > 0 {
			t.Errorf("%s: build left %d effects pending", name, len(fx))
		}
	}
}

// firstDiff names the field where a and b first differ.
func firstDiff(a, b string) string {
	pa, pb := strings.Split(a, "|"), strings.Split(b, "|")
	for i := range pa {
		if i >= len(pb) || pa[i] != pb[i] {
			return fmt.Sprintf("field %d", i)
		}
	}
	return "length"
}
