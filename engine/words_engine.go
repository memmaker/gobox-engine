package engine

import (
	"fmt"
	"gobox/ecl"
	"gobox/rules"
	"io/fs"
	"slices"
)

// The canonical fields of the words the VM itself keeps (every game's layout names them so,
// when it has them).
const (
	fieldLastBlock   = "LastBlock"       // the block NewECL left (sotsb, cotab, pod); POR has none: the world's AreaKind
	fieldRemainder   = "DivideRemainder" // where DIVIDE leaves its remainder (sotsb, cotab, pod); POR has none
	fieldLoopCounter = "LoopCounter"     // FOR/NEXT's counter (Buck Rogers); the Pool games have none
	fieldLoopLimit   = "LoopLimit"       // FOR/NEXT's limit (Buck Rogers); the Pool games have none
	fieldPriceScale  = "PriceScale"      // the shops' price factor (cotab, sotsb, pod); POR has none
	fieldVault       = "VaultRequest"    // 1: Combat with no monsters opens the vault (pod); the others have none
)

// layoutField is the loaded script's field called name; ok false when the game's layout has none.
func (e *Engine) layoutField(name string) (ecl.Ref, bool) {
	l := e.vm.Script.Layout
	if l == nil {
		return ecl.Ref{}, false
	}
	r, err := l.Named(name)
	return r, err == nil
}

// bindWords points DIVIDE's remainder at the layout's DivideRemainder, for the script just
// loaded; a game without one keeps no remainder.
func (e *Engine) bindWords() {
	e.vm.Remainder, _ = e.layoutField(fieldRemainder)
	e.vm.LoopCounter, _ = e.layoutField(fieldLoopCounter)
	e.vm.LoopLimit, _ = e.layoutField(fieldLoopLimit)
}

// setLastBlock keeps block, the one NewECL leaves, to come back to: in the layout's LastBlock
// field, else (POR) in the world's AreaKind (POR keeps the block in the area's kind word).
func (e *Engine) setLastBlock(block int) {
	if r, ok := e.layoutField(fieldLastBlock); ok {
		e.host.Set(r, int64(block))
		return
	}
	e.world.AreaKind = uint16(block)
}

// BoundFields are the canonical script field names the engine gives a meaning: a game's
// layout names a field so when, and only when, it has that field (each game's layout test).
func BoundFields() []string {
	names := []string{fieldLastBlock, fieldRemainder, fieldLoopCounter, fieldLoopLimit, fieldPriceScale, fieldVault, fieldHours, tourDone}
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
		l, err := ecl.ReadLayout(data, a)
		if err != nil {
			return nil, err
		}
		for _, f := range l.Fields {
			have[f.Name] = true
			for _, e := range f.Elements {
				have[e] = true
			}
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
	if count > 1 && e.ruleset().Check(rules.BundlePrices{}, nil).Yes {
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
