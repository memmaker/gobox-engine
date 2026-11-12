package pool

import (
	"fmt"
	"gobox/core"
	"gobox/core/dnd"
	"gobox/core/goldbox"
	"slices"

	"gobox/ecl/script"
)

// newEngineFields is the engine's table of the fields it gives a meaning.
func (e *Engine) newEngineFields() map[string]engineField {
	coin := func(i int) (int64, bool) {
		if i%2 != 0 { // the odd ones are the original's high bytes, no field of the record
			return 0, false
		}
		return e.selFact(core.CoinFact{Coin: scriptCoin[i/2]}), true
	}
	m := map[string]engineField{
		"SymbolSet": { // ovr08:07BC/0A74 +0x312: the symbol set (0802:094A)
			get: func(int) (int64, bool) { return int64(e.set), true },
			set: func(_ int, v int64) { e.setWorldWord(core.WordSymbolSet, int(v)); e.set = goldbox.Set(v) },
		},
		"PartyCount": {
			get: func(int) (int64, bool) { return int64(e.worldWord(core.WordPartyCount)), true },
			set: func(_ int, v int64) { e.setWorldWord(core.WordPartyCount, int(v)) },
		},
		"WallSlot1": e.wallSlot(1), "WallSlot2": e.wallSlot(2), "WallSlot3": e.wallSlot(3),
		// the selected member's window (SpecialReadZ1/WriteZ1)
		"SelName": {
			stored:  true,
			set:     func(_ int, v int64) { e.memberGoneIf(uint16(v)) },
			text:    func() string { return e.ruleset().Sheet(e.selected()).Name },
			setText: func(s string) { e.setSelValue(core.FactName, core.Text(s)) },
		},
		"SelCon":  reads(func() uint16 { return uint16(e.selFact(core.AbilityFact{Ability: core.Constitution})) }),
		"SelRace": reads(func() uint16 { return uint16(selIndex(e, core.FactRace, goldbox.RecordRace[:])) }), // a fix of the Pascal's (SpecialReadZ1): the race checks read 0
		"SelMemorized": {stored: true, set: func(i int, v int64) { // slot i+1 (slot 0 has no address)
			if k := i + 1; k < core.MemoSlots {
				e.setSelValue(core.MemoFact{Slot: k}, core.Number(byte(v))) // the byte as SpecialWriteZ1 stores it
			}
		}},
		"SelSaveVsPetrify": reads(func() uint16 { return uint16(e.selFact(core.SaveFact{Save: dnd.SavePetrify})) }),
		"SelPickPockets":   reads(func() uint16 { return uint16(e.selFact(core.SkillFact{Skill: dnd.PickPockets})) }),
		fieldSelFindTraps:  reads(func() uint16 { return uint16(e.selFact(core.SkillFact{Skill: dnd.FindTraps})) }),
		"SelThiefSkill6":   reads(func() uint16 { return uint16(e.selFact(core.SkillFact{Skill: dnd.Climb})) }),
		"SelControl": {stored: true, get: func(int) (int64, bool) { return e.selFact(core.FactControl), true },
			set: func(_ int, v int64) {
				if v > 0xB2 {
					v -= 0x32
				}
				e.setSelFact(core.FactControl, v)
			}},
		"SelCoins": {stored: true, get: coin, set: func(i int, v int64) {
			if i%2 == 0 {
				e.setSelFact(core.CoinFact{Coin: scriptCoin[i/2]}, int64(uint16(v)))
			}
		}},
		"SelAlignment": reads(func() uint16 { return signedByte(selIndex(e, core.FactAlignment, goldbox.RecordAlignment[:])) }),
		"SelExpValue": {stored: true, get: func(int) (int64, bool) { return e.selFact(core.FactExpValue), true },
			set: func(_ int, v int64) { e.setSelFact(core.FactExpValue, v) }},
		"SelExpPerHP": {stored: true, get: func(int) (int64, bool) { return e.selFact(core.FactExpPerHP), true },
			set: func(_ int, v int64) { e.setSelFact(core.FactExpPerHP, v) }},
		"SelStatus": {stored: true, get: func(int) (int64, bool) {
			if e.selValue(core.FactActive).(core.Yes) {
				return 1, true
			}
			return 0x80, true
		}, set: func(_ int, v int64) {
			if v >= 0x80 {
				st := core.StatusInactive
				if v == 0x87 {
					st = core.StatusStoned
				}
				a := core.SetStatusAction{Who: e.selected(), Status: st}
				if _, err := e.ruleset().Resolve(nil, a, nil); err != nil {
					panic(fmt.Sprintf("script: SelStatus %#x: %v", v, err))
				}
			}
			if v == 0 {
				e.es.memberDied = 1
			}
		}},
		"SelSide": {stored: true, get: func(int) (int64, bool) {
			return int64(scriptSide[e.selValue(core.FactSide).(core.Side)]), true
		}, set: func(_ int, v int64) {
			for side, n := range scriptSide {
				if int64(n) == v {
					e.setSelValue(core.FactSide, side)
				}
			}
		}},
		fieldSelMove:     reads(func() uint16 { return uint16(e.selFact(core.FactMove)) }),
		"SelPartyIndex":  reads(func() uint16 { return uint16(e.partyIndex(e.selected())) }),
		"SelPartyIndex2": reads(func() uint16 { return uint16(e.partyIndex(e.selected())) }),
		"SelReaction":    reads(func() uint16 { return charismaReaction(byte(e.selFact(core.AbilityFact{Ability: core.Charisma}))) }),
		// the engine facts outside the records (ReadGbvm/WriteGbvm's specials)
		"ProbeCell":   fact(func() uint16 { return e.es.specialB1 }, nil),
		"UnusedB8":    fact(nil, func(v uint16) { e.es.scriptVarB8 = v }),
		"UnusedB9":    fact(nil, func(v uint16) { e.es.scriptVarB9 = v }),
		"ProbeX":      fact(func() uint16 { return e.es.specialFB }, nil),
		"ProbeY":      fact(func() uint16 { return e.es.specialFC }, nil),
		"FacingRaw":   fact(func() uint16 { return uint16(e.facing) }, nil),
		"ProbeWall":   fact(nil, nil),
		"Sound":       fact(nil, func(v uint16) { e.es.soundVar = v }),
		"MapPatchA":   fact(nil, nil),
		"MapPatchB":   fact(nil, nil),
		"UnknownD015": fact(nil, nil),
		"PartyX": fact(func() uint16 { return signedByte(e.posX) }, func(v uint16) {
			e.es.posChanged = 1
			e.posX = int(v)
		}),
		"PartyY": fact(func() uint16 { return signedByte(e.posY) }, func(v uint16) {
			e.posY = int(v)
			e.es.posChanged = 1
		}),
		"Facing": fact(func() uint16 { return uint16(e.facing >> 1) }, func(v uint16) {
			e.facing = [4]int{0, 2, 4, 6}[v%4]
			e.ext.back = 0 // the script chose the facing
			e.es.posChanged = 1
		}),
		"WallAhead": fact(func() uint16 { return uint16(e.es.wallAhead) }, nil),
		"Cell":      fact(func() uint16 { return uint16(e.es.cell) }, nil),
		"SetGroundShade": fact(nil, func(v uint16) {
			e.es.groundShade = byte(v)
			e.es.skyChanged = 1
		}),
		"SetSkyShade": fact(nil, func(v uint16) {
			e.es.skyShade = byte(v)
			e.es.skyChanged = 1
		}),
		// area 2's well: the party's affects (the Pascal's fix, effectTable)
		fieldEffectTable: {stored: true, get: func(i int) (int64, bool) { v, ok := e.effect(i); return int64(v), ok }},
	}
	for name, w := range scriptWords {
		m[name] = e.wordField(w)
	}
	for name, w := range portraitWords {
		f := e.wordField(w) // ovr08:0C77: writing the HEAD/BODY words redraws the portrait
		set := f.set
		f.set = func(i int, v int64) { set(i, v); e.es.portraitChanged = 1 }
		m[name] = f
	}
	return m
}

// wallSlot is game word WallSlot<slot>: a value above 0x80 loads that
// walldef into the slot (ovr08:0A74 +0x322/0x324/0x326).
func (e *Engine) wallSlot(slot int) engineField {
	return engineField{stored: true, set: func(_ int, v int64) {
		if v > 0x80 {
			e.LoadWalldef(int(v&0x7F), slot)
		}
	}}
}

// memberGoneIf: writing 0 to the selected member's name word removes it.
func (e *Engine) memberGoneIf(v uint16) {
	if v == 0 {
		e.es.memberGone = 1
	}
}

// cells is the store of field f: the scratch for a local, the game's
// saved state otherwise.
func (e *Engine) cells(f *script.Field) fieldCells {
	if f.Owner == script.OwnerLocal {
		return e.locals
	}
	return e.vars
}

// varField is the field r names in the running script's layout; an
// element past a number field's end is a layout too short for its script.
func (e *Engine) varField(r script.Ref) *script.Field {
	f := e.vm.Script.Layout.Field(r)
	if f.Owner == script.OwnerCall {
		panic(fmt.Sprintf("script: %s is not a variable", f.Name))
	}
	return f
}

// getVar = ReadGbvm (ovr08:0E0C) on a Ref.
func (e *Engine) getVar(r script.Ref) int64 {
	f := e.varField(r)
	if f.Owner == script.OwnerConst { // a table in the script's own code (the code image at OriginalSave.CodeBase)
		e.inField(f, r.Index)
		return int64(e.vm.CodeByte(uint16(f.From + r.Index)))
	}
	ef, bound := e.engineFields[f.Name]
	if bound && ef.get != nil {
		if v, ok := ef.get(r.Index); ok {
			return v
		}
	}
	if bound && !ef.stored {
		panic(fmt.Sprintf("script: %s has no value to read", f.Name))
	}
	if f.Width != script.WidthText {
		e.inField(f, r.Index)
	}
	if c := e.cells(f)[f.Name]; r.Index < len(c) {
		return c[r.Index]
	}
	return 0
}

// setVar = WriteGbvm (ovr08:0C77) on a Ref.
func (e *Engine) setVar(r script.Ref, v int64) {
	f := e.varField(r)
	if f.ReadOnly {
		panic(fmt.Sprintf("script: %s is read-only", f.Name))
	}
	v = cutToWidth(f.Width, v)
	ef, bound := e.engineFields[f.Name]
	if !bound || ef.stored {
		if f.Width != script.WidthText {
			e.inField(f, r.Index)
		}
		e.storeCells(f, r.Index, []int64{v})
	}
	if bound && ef.set != nil {
		ef.set(r.Index, v)
	} else if bound && !ef.stored {
		panic(fmt.Sprintf("script: %s cannot be written", f.Name))
	}
}

// storeCells writes vs into f's cells from element i on.
func (e *Engine) storeCells(f *script.Field, i int, vs []int64) {
	m := e.cells(f)
	c := m[f.Name]
	if n := max(f.Len, i+len(vs)); len(c) < n {
		c = append(c, make([]int64, n-len(c))...)
	}
	copy(c[i:], vs)
	m[f.Name] = c
}

// getText = ovr08:12FB: the text in r, up to its NUL.
func (e *Engine) getText(r script.Ref) string {
	f := e.varField(r)
	if ef, ok := e.engineFields[f.Name]; ok && ef.text != nil {
		return ef.text()
	}
	e.textField(f)
	var b []byte
	for _, v := range e.cells(f)[f.Name][min(r.Index, len(e.cells(f)[f.Name])):] {
		if v == 0 || len(b) == 255 {
			break
		}
		b = append(b, byte(v))
	}
	return string(b)
}

// setText = ovr08:0F53: s and its NUL into r.
func (e *Engine) setText(r script.Ref, s string) {
	f := e.varField(r)
	if ef, ok := e.engineFields[f.Name]; ok && ef.setText != nil {
		ef.setText(s)
		return
	}
	e.textField(f)
	vs := make([]int64, len(s)+1)
	for i := range len(s) {
		vs[i] = int64(s[i])
	}
	e.storeCells(f, r.Index, vs)
}

func (e *Engine) textField(f *script.Field) {
	if f.Width != script.WidthText {
		panic(fmt.Sprintf("script: %s is no text field", f.Name))
	}
}

// savedWord is element 0 of the game's field name in any area (the quest
// log reads the quest table wherever the party is).
func (e *Engine) savedWord(name string) int64 {
	e.mustGameField(name)
	if ef, ok := e.engineFields[name]; ok && ef.get != nil {
		if v, ok := ef.get(0); ok {
			return v
		}
	}
	if c := e.vars[name]; len(c) > 0 {
		return c[0]
	}
	return 0
}

// setSavedWord sets element 0 of the game's field name, as a script would.
func (e *Engine) setSavedWord(name string, v int64) {
	e.mustGameField(name)
	if ef, ok := e.engineFields[name]; ok && !ef.stored {
		ef.set(0, v)
		return
	}
	c := e.vars[name]
	if len(c) == 0 {
		c = []int64{0}
	}
	c[0] = v
	e.vars[name] = c
}

// mustGameField: name is a field of the game's layouts.
func (e *Engine) mustGameField(name string) {
	if _, ok := e.engineFields[name]; ok {
		return
	}
	if !slices.ContainsFunc(e.gameFields(), func(f script.Field) bool { return f.Name == name }) {
		panic("script: no game field " + name)
	}
}

// The fields the engine itself names (layout/global.txt; EffectTable is
// area 2's).
const (
	fieldEffectTable  = "EffectTable"
	fieldSelFindTraps = "SelFindTraps"
	fieldSelMove      = "SelMove"
	fieldSurprise     = "Surprise"
)

// reads is a field whose read is f, whose write only stores.
func reads(f func() uint16) engineField {
	return engineField{get: func(int) (int64, bool) { return int64(f()), true }, stored: true}
}

// fact is a field outside the records: read is r (nil: 0), write is w (nil: ignored).
func fact(r func() uint16, w func(v uint16)) engineField {
	f := engineField{get: func(int) (int64, bool) { return 0, true }, set: func(int, int64) {}}
	if r != nil {
		f.get = func(int) (int64, bool) { return int64(r()), true }
	}
	if w != nil {
		f.set = func(_ int, v int64) { w(uint16(v)) }
	}
	return f
}

// scriptCoin is the coin of each SelCoins word pair (copper .. platinum).
var scriptCoin = [5]core.Coin{dnd.Copper, dnd.Silver, dnd.Electrum, dnd.Gold, dnd.Platinum}

func signedByte(v int) uint16 { return uint16(int16(int8(v))) }

// charismaReaction = ovr08:07BC +0x2CF: the reaction adjustment by charisma.
func charismaReaction(v byte) uint16 {
	switch {
	case v == 3:
		return 0
	case v == 4:
		return 5
	case v == 5:
		return 0xA
	case v == 6:
		return 0xF
	case v == 7:
		return 0x14
	case v >= 8 && v <= 0xC:
		return 0x19
	case v == 0xD:
		return 0x1E
	case v == 0xE:
		return 0x23
	case v == 0xF:
		return 0x28
	case v == 0x10:
		return 0x32
	case v == 0x11:
		return 0x37
	case v >= 0x12 && v <= 0x19:
		return 0x3C
	}
	return 0
}

// cutToWidth is v as a field of width w holds it. The original's
// WriteGbvm (eclmenu.pas) took a Word and kept it whole in every writable
// zone (0-2 are word cells; zone 3, the code, is the const fields'), so a
// byte, word or text field holds a word: the layout's byte is the operand's
// type, not the cell's. A dword holds 32 bits.
func cutToWidth(w script.Width, v int64) int64 {
	switch w {
	case script.WidthByte, script.WidthWord, script.WidthText:
		return int64(uint16(v))
	case script.WidthDword:
		return int64(uint32(v))
	}
	panic(fmt.Sprintf("script: width %d", w))
}

func (e *Engine) inField(f *script.Field, i int) {
	if i < 0 || i >= f.Len {
		panic(fmt.Sprintf("script: %s[%d] is past the field's %d (layout too short)", f.Name, i, f.Len))
	}
}

// wordField is the script field of the ruleset's word n.
func (e *Engine) wordField(n core.WorldWord) engineField {
	return engineField{
		get: func(int) (int64, bool) { return int64(e.worldWord(n)), true },
		set: func(_ int, v int64) { e.setWorldWord(n, int(v)) },
	}
}

// effect is entry i of the party's affect table.
func (e *Engine) effect(i int) (uint16, bool) {
	a := e.ruleset().Check(core.WordIs{Word: core.WordEffect, Index: i}, nil)
	return uint16(a.Value), a.Yes
}

// partyIndex is c's place in the party, the party's size when it is not in it (ovr08:0756).
func (e *Engine) partyIndex(id core.CharID) int {
	ids := e.partyIDs()
	if i := slices.Index(ids, id); i >= 0 {
		return i
	}
	return len(ids)
}

// scriptWords are the plain record words a script reads and writes, by the
// name its layout gives them (data -> the typed word, once, here).
var scriptWords = map[string]core.WorldWord{
	"MapX":         core.WordMapX,
	"MapY":         core.WordMapY,
	"Geo":          core.WordGeo,
	"TenMinutes":   core.WordTenMinutes,
	"Hours":        core.WordHours,
	"NoMagic":      core.WordNoMagic,
	"In3D":         core.WordIn3D,
	"WallSplitA":   core.WordWallSplitA,
	"WallSplitB":   core.WordWallSplitB,
	"PrevX":        core.WordPrevX,
	"PrevY":        core.WordPrevY,
	"AreaKind":     core.WordAreaKind,
	"NoCoords":     core.WordNoCoords,
	"MsgDelay":     core.WordMsgDelay,
	"ViewFlags":    core.WordViewFlags,
	"MarkFlags4":   core.WordMarkFlags4,
	"Mark3":        core.WordMark3,
	"Mark4Off":     core.WordMark4Off,
	"MarkFlags57":  core.WordMarkFlags57,
	"RiverOpen":    core.WordRiverOpen,
	"Boss":         core.WordBoss,
	"Mark4":        core.WordMark4,
	"PartyExp":     core.WordTrainable,
	"MaxDist":      core.WordMaxDist,
	"MonsterDist":  core.WordMonsterDist,
	"Morale":       core.WordMorale,
	"CombatEnd":    core.WordCombatEnd,
	"Beaten":       core.WordBeaten,
	"HoldMove":     core.WordHoldMove,
	"Search":       core.WordSearch,
	"Surprise":     core.WordSurprise,
	"RestEvery":    core.WordRestEvery,
	"RestChance":   core.WordRestChance,
	"LeftMap":      core.WordLeftMap,
	"Portrait":     core.WordPortrait,
	"Temple":       core.WordTemple,
	"NoLoot":       core.WordNoLoot,
	"Duel":         core.WordDuel,
	"AttackedAlly": core.WordAttackedAlly,
	"Shop":         core.WordShop,
	"MonsterHit":   core.WordMonsterHit,
	"PartyHit":     core.WordPartyHit,
	"PartyMove":    core.WordPartyMove,
}

// portraitWords are the sky words: a write redraws the portrait.
var portraitWords = map[string]core.WorldWord{"Sky": core.WordSky, "SkyInside": core.WordSkyInside}
