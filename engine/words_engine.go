package engine

import (
	"fmt"
	"gobox/adventure"
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
	fieldPriceMult   = "PriceMultiplier" // the shop's price in sixteenths of a value (Buck Rogers), in the ruleset's unit: handed to it with StockShop
	fieldScriptBlock = "ScriptBlock"     // the block whose move entry last ran (Buck Rogers: its start script still reads the block left, a loaded game its own); the Pool games have none
	fieldWildTerrain = "WildTerrain"     // the adventure's terrain number for a fight outdoors (BR zone0 wildernessTerrain, Area[183]); the engine only carries it to the rules
	fieldTravelKind  = "TravelKind"      // the place's traits (rules.PlaceTraits, canonical; BR keeps its areaKind byte, enc=BrTravelKind); no traits while FacingOrRaw reads the eight-way facing (Buck Rogers); the Pool games have none
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

// placeTraits is the place's traits, the canonical read of the layout's TravelKind; ok false for a game with none
// (the Pool games): no traits.
func (e *Engine) placeTraits() (rules.PlaceTraits, bool) {
	r, ok := e.layoutField(fieldTravelKind)
	if !ok {
		return 0, false
	}
	return rules.PlaceTraits(e.canonValue(e.varField(r), e.host.Get(r))), true
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
	if r, ok := e.layoutField(fieldScriptBlock); ok { // Buck Rogers: EclNewEcl saves the block left in zone0's savedEclId (eclvm.pas), which the scripts read as ScriptBlock
		e.host.Set(r, int64(block))
	}
	e.world.AreaKind = uint16(block)
}

// markScriptBlock sets the layout's ScriptBlock to the block in play: once its start script has run (a
// begun game's, BR1 pO0CA 3209) and before every run of its move entry.
func (e *Engine) markScriptBlock() {
	if r, ok := e.layoutField(fieldScriptBlock); ok {
		e.host.Set(r, int64(e.gs.areaKind))
	}
}

// runMoveEntry runs the block's move entry (events[0]: before a step, after a block's start, a scripted
// place's every pass; BR1 pO0CA: Area[242] = the block at 3209's loop, 0FE4's step and input's default).
func (e *Engine) runMoveEntry() {
	e.markScriptBlock()
	e.RunECL(e.es.events[0])
}

// BoundFields are the canonical script field names the engine gives a meaning: a game's
// layout names a field so when, and only when, it has that field (each game's layout test).
func BoundFields() []string {
	names := []string{fieldLastBlock, fieldRemainder, fieldLoopCounter, fieldLoopLimit, fieldPriceScale, fieldPriceMult, fieldScriptBlock, fieldTravelKind, fieldVault, fieldWildTerrain, fieldFightState, fieldHours, tourDone}
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

// travelKindWritten is the screen a script's new TravelKind selects (BR eclmenu.pas WriteGbvm, GBVM_AREA_KIND: written a
// value other than the one it holds, the interface state becomes the wilderness, the orbit map or the 3D view): the
// place's own screen (a base's menus, the rocket) goes with the kind it was written under.
func (e *Engine) travelKindWritten(kind byte) {
	switch {
	case adventure.BRPlaceTraits(kind).Has(rules.TraitWilderness):
		e.world.In3D, e.world.Scripted, e.world.Flying = 0, 0, 0
	case kind == adventure.BROrbit:
		e.startFlying()
	default:
		e.world.In3D, e.world.Scripted, e.world.Flying = 1, 0, 0
	}
}
