package pool

import (
	"fmt"
	"gobox/core"
	"gobox/ecl/script"
	"io/fs"
	"slices"
)

// The canonical fields of the words the VM itself keeps (every game's layout names them so,
// when it has them).
const (
	fieldLastBlock  = "LastBlock"       // the block NewECL left (sotsb, cotab, pod); POR has none: core.WordLastBlock
	fieldRemainder  = "DivideRemainder" // where DIVIDE leaves its remainder (sotsb, cotab, pod); POR has none
	fieldPriceScale = "PriceScale"      // the shops' price factor (cotab, sotsb, pod); POR has none
)

// layoutField is the loaded script's field called name; ok false when the game's layout has none.
func (e *Engine) layoutField(name string) (script.Ref, bool) {
	l := e.vm.Script.Layout
	if l == nil {
		return script.Ref{}, false
	}
	r, err := l.Named(name)
	return r, err == nil
}

// bindWords points DIVIDE's remainder at the layout's DivideRemainder, for the script just
// loaded; a game without one keeps no remainder.
func (e *Engine) bindWords() {
	e.vm.Remainder, _ = e.layoutField(fieldRemainder)
}

// setLastBlock keeps block, the one NewECL leaves, to come back to: in the layout's LastBlock
// field, else (POR) in the ruleset's core.WordLastBlock.
func (e *Engine) setLastBlock(block int) {
	if r, ok := e.layoutField(fieldLastBlock); ok {
		e.host.Set(r, int64(block))
		return
	}
	e.setWorldWord(core.WordLastBlock, block)
}

// BoundFields are the canonical script field names the engine gives a meaning: a game's
// layout names a field so when, and only when, it has that field (each game's layout test).
func BoundFields() []string {
	names := []string{fieldLastBlock, fieldRemainder, fieldPriceScale, fieldScratch6, fieldHours}
	for n := range (&Engine{}).newEngineFields() {
		names = append(names, n)
	}
	slices.Sort(names)
	return slices.Compact(names)
}

// UnboundFields are the BoundFields that no layout of data (a game's script data: layout/global.txt
// and every layout/area<N>.txt) names: the bindings the game does not have.
func UnboundFields(data fs.FS) ([]string, error) {
	names, err := fs.Glob(data, "layout/area*.txt")
	if err != nil {
		return nil, err
	}
	have := map[string]bool{}
	for _, n := range append(names, "layout/global.txt") {
		a := 0
		if n != "layout/global.txt" {
			if _, err := fmt.Sscanf(n, "layout/area%d.txt", &a); err != nil {
				return nil, err
			}
		}
		l, err := script.ReadLayout(data, a)
		if err != nil {
			return nil, err
		}
		for _, f := range l.Fields {
			have[f.Name] = true
		}
	}
	var out []string
	for _, n := range BoundFields() {
		if !have[n] {
			out = append(out, n)
		}
	}
	return out, nil
}

// shopPrice is price as the shop asks it for an item of count pieces: by the layout's price
// factor (a game whose layout has none, POR, asks the price), then, in a game whose shops sell
// bundles at their count times the price (pod, core.BundlePrices), times the count.
func (e *Engine) shopPrice(price, count int32) int32 {
	if r, ok := e.layoutField(fieldPriceScale); ok {
		price = scalePrice(price, int32(e.host.Get(r)))
	}
	if count > 1 && e.ruleset().Check(core.BundlePrices{}, nil).Yes {
		price *= count
	}
	return price
}

// scalePrice is price times factor over 16, for the factors the shops set (1 2 4 8 32 64 128;
// any other, 16 the normal one, leaves it).
func scalePrice(price, factor int32) int32 {
	switch factor {
	case 1, 2, 4, 8, 32, 64, 128:
		return price * factor / 16
	}
	return price
}
