package engine

import (
	"fmt"
	"gobox/adventure"
	"gobox/rules"
	"log"
	"slices"

	"gobox/ecl"
)

// newEngineFields is the engine's table of the fields it gives a meaning.
func (e *Engine) newEngineFields() map[string]engineField {
	coin := func(i int) (int64, bool) {
		if i%2 != 0 { // the odd ones are the original's high bytes, no field of the record
			return 0, false
		}
		f, ok := selCoin(i)
		if !ok {
			return 0, false
		}
		return e.selFact(f), true
	}
	m := map[string]engineField{
		"SymbolSet": { // ovr08:07BC/0A74 +0x312: the symbol set (0802:094A)
			get: func(int) (int64, bool) {
				if e.eclVersion.Behaviour.SymbolSetFrames {
					return int64(e.worldWord(rules.WordSymbolSet)), true
				}
				return int64(e.set), true
			},
			set: func(_ int, v int64) {
				e.setWorldWord(rules.WordSymbolSet, int(v))
				if !e.eclVersion.Behaviour.SymbolSetFrames {
					e.set = rules.Set(v)
				}
			},
		},
		"PartyCount": {
			get: func(int) (int64, bool) { return int64(e.worldWord(rules.WordPartyCount)), true },
			set: func(_ int, v int64) { e.setWorldWord(rules.WordPartyCount, int(v)) },
		},
		"WallSlot1": e.wallSlot(1), "WallSlot2": e.wallSlot(2), "WallSlot3": e.wallSlot(3),
		// the selected member's window (SpecialReadZ1/WriteZ1)
		"SelName": {
			stored:  true,
			set:     func(_ int, v int64) { e.memberGoneIf(uint16(v)) },
			text:    func() string { return e.ruleset().Sheet(e.selected()).Name },
			setText: func(s string) { e.setSelValue(rules.FactName, rules.Text(s)) },
		},
		"SelCon":  reads(func() uint16 { return uint16(e.selFact(rules.AbilityFact{Ability: rules.Constitution})) }),
		"SelRace": reads(func() uint16 { return uint16(selIndex(e, rules.FactRace, e.ruleset().Catalog().Races)) }), // a fix of the Pascal's (SpecialReadZ1): the race checks read 0
		"SelMemorized": {stored: true, set: func(i int, v int64) { // slot i+1 (slot 0 has no address)
			if k := i + 1; k < rules.MemoSlots {
				e.setSelValue(rules.MemoFact{Slot: k}, rules.Number(byte(v))) // the byte as SpecialWriteZ1 stores it
			}
		}},
		"SelSaveVsPetrify": reads(func() uint16 { return uint16(e.selFact(rules.SaveFact{Save: rules.SavePetrify})) }),
		"SelPickPockets":   reads(func() uint16 { return uint16(e.selFact(rules.SkillFact{Skill: rules.PickPockets})) }),
		fieldSelFindTraps:  reads(func() uint16 { return uint16(e.selFact(rules.SkillFact{Skill: rules.FindTraps})) }),
		"SelThiefSkill6":   reads(func() uint16 { return uint16(e.selFact(rules.SkillFact{Skill: rules.Climb})) }),
		"SelControl": {stored: true, get: func(int) (int64, bool) { return e.selFact(rules.FactControl), true },
			set: func(_ int, v int64) {
				if v > 0xB2 {
					v -= 0x32
				}
				e.setSelFact(rules.FactControl, v)
			}},
		"SelCoins": {stored: true, get: coin, set: func(i int, v int64) {
			if f, ok := selCoin(i); ok {
				e.setSelFact(f, int64(uint16(v)))
			}
		}},
		"SelAlignment": {stored: true, get: func(int) (int64, bool) {
			return int64(signedByte(selIndex(e, rules.FactAlignment, e.ruleset().Catalog().Alignments))), true
		}, set: func(_ int, v int64) { e.setSelFact(rules.FactAlignment, int64(uint8(v))) }}, // the record's byte (alignmentNameTable index)
		"SelExpValue": {stored: true, get: func(int) (int64, bool) { return e.selFact(rules.FactExpValue), true },
			set: func(_ int, v int64) { e.setSelFact(rules.FactExpValue, v) }},
		"SelExpPerHP": {stored: true, get: func(int) (int64, bool) { return e.selFact(rules.FactExpPerHP), true },
			set: func(_ int, v int64) { e.setSelFact(rules.FactExpPerHP, v) }},
		"SelStatus": {stored: true, get: func(int) (int64, bool) {
			if e.selValue(rules.FactActive).(rules.Yes) {
				return 1, true
			}
			return 0x80, true
		}, set: func(_ int, v int64) {
			if v >= 0x80 || v == adventure.CanonStatusOkay {
				st := rules.StatusInactive
				switch v {
				case adventure.CanonStatusStoned:
					st = rules.StatusStoned
				case adventure.CanonStatusOkay: // Okay revives the member (sotsb, pod; POR's scripts never write it)
					st = rules.StatusOkay
				}
				a := rules.SetStatusAction{Who: e.selected(), Status: st}
				if _, err := e.ruleset().Resolve(nil, a, nil); err != nil {
					panic(fmt.Sprintf("script: SelStatus %#x: %v", v, err))
				}
			}
			if v == 0 {
				e.es.memberDied = 1
			}
		}},
		"SelSide": {stored: true, get: func(int) (int64, bool) {
			return int64(adventure.ScriptSide[e.selValue(rules.FactSide).(rules.Side)]), true
		}, set: func(_ int, v int64) {
			for side, n := range adventure.ScriptSide {
				if int64(n) == v {
					e.setSelValue(rules.FactSide, side)
				}
			}
		}},
		fieldSelMove:     reads(func() uint16 { return uint16(e.selFact(rules.FactMove)) }),
		"SelPartyIndex":  reads(func() uint16 { return uint16(e.partyIndex(e.selected())) }),
		"SelPartyIndex2": reads(func() uint16 { return uint16(e.partyIndex(e.selected())) }),
		"SelReaction":    reads(func() uint16 { return charismaReaction(byte(e.selFact(rules.AbilityFact{Ability: rules.Charisma}))) }),
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
		// the way the party goes: during a step back or aside (gobox's: the view's
		// facing stays) the way of the step, as the originals' scripts take the
		// facing for it (they only ever stepped ahead): an exit, a door, a guard
		// by Facing is the one the step goes through
		"Facing": fact(func() uint16 { return uint16(e.moveDir() >> 1) }, func(v uint16) {
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
func (e *Engine) cells(f *ecl.Field) fieldCells {
	if f.Owner == ecl.OwnerLocal {
		return e.locals
	}
	return e.vars
}

// varField is the field r names in the running script's layout; an
// element past a number field's end is a layout too short for its script.
func (e *Engine) varField(r ecl.Ref) *ecl.Field {
	f := e.vm.Script.Layout.Field(r)
	if f.Owner == ecl.OwnerCall {
		panic(fmt.Sprintf("script: %s is not a variable", f.Name))
	}
	if f.Width == ecl.WidthBytes && !e.storesByOperand() {
		panic(fmt.Sprintf("script: %s is a byte buffer, which only a dialect that stores by operand has", f.Name))
	}
	return f
}

// storesByOperand: the running dialect's operand type says the width of a read or write.
func (e *Engine) storesByOperand() bool { return e.eclVersion.Behaviour.StoreByOperand }

// operandWidth is the width r's access reads and writes when the dialect stores by operand
// and the script addressed the cell through an operand; else (ok false) the field's own cell.
func (e *Engine) operandWidth(r ecl.Ref) (ecl.Width, bool) {
	return r.As, e.storesByOperand() && r.As != 0
}

// getVar = ReadGbvm (ovr08:0E0C) on a Ref.
func (e *Engine) getVar(r ecl.Ref) int64 {
	f := e.varField(r)
	if f.Width == ecl.WidthBytes {
		return e.getBytes(f, r)
	}
	v := e.getCell(f, r)
	if w, ok := e.operandWidth(r); ok {
		return readAsWidth(w, v)
	}
	return v
}

// getCell is the cell r names, whole.
func (e *Engine) getCell(f *ecl.Field, r ecl.Ref) int64 {
	if f.Owner == ecl.OwnerConst { // a table in the script's own code (the code image at OriginalSave.CodeBase)
		if r.Index < 0 || r.Index >= f.Len { // the layout does not cover the read (ecl.ConstLayout takes a field as far as its reads go)
			e.tableFault(f, r.Index)
		}
		return int64(e.vm.CodeByte(uint16(f.From + r.Index))) // as the original: the code byte at base plus index
	}
	ef, bi, bound := e.binding(f, r.Index)
	if bound && ef.get != nil {
		if v, ok := ef.get(bi); ok {
			return e.gameValue(f, v)
		}
	}
	if bound && !ef.stored {
		panic(fmt.Sprintf("script: %s has no value to read", f.Name))
	}
	if f.Width != ecl.WidthText {
		e.inField(f, r.Index)
	}
	if c := e.cells(f)[f.Name]; r.Index < len(c) {
		return c[r.Index]
	}
	return 0
}

// setVar = WriteGbvm (ovr08:0C77) on a Ref.
func (e *Engine) setVar(r ecl.Ref, v int64) {
	f := e.varField(r)
	if f.ReadOnly {
		panic(fmt.Sprintf("script: %s is read-only", f.Name))
	}
	if f.Width == ecl.WidthBytes {
		e.setBytes(f, r, v)
		return
	}
	if w, ok := e.operandWidth(r); ok {
		v = cutToOperand(w, v)
	} else {
		v = cutToWidth(f.Width, v)
	}
	ef, bi, bound := e.binding(f, r.Index)
	if !bound || ef.stored {
		if f.Width != ecl.WidthText {
			e.inField(f, r.Index)
		}
		e.storeCells(f, r.Index, []int64{v})
	}
	if bound && writesIgnored(f) {
		return
	}
	if bound && ef.set != nil {
		ef.set(bi, e.canonValue(f, v))
	} else if bound && !ef.stored {
		panic(fmt.Sprintf("script: %s cannot be written", f.Name))
	}
}

// binding is the engine's own field for element i of f and the element it
// reads there: a named element's binding (Boss, element 13 of POR's
// QuestTable) as its element 0, else f's at f's canonical element i.
func (e *Engine) binding(f *ecl.Field, i int) (engineField, int, bool) {
	if n := f.ElementName(i); n != "" {
		if ef, ok := e.engineFields[n]; ok {
			return ef, 0, true
		}
	}
	ef, ok := e.engineFields[f.Name]
	return ef, canonElement(f, i), ok
}

// storeCells writes vs into f's cells from element i on.
func (e *Engine) storeCells(f *ecl.Field, i int, vs []int64) {
	m := e.cells(f)
	c := m[f.Name]
	if n := max(f.Len, i+len(vs)); len(c) < n {
		c = append(c, make([]int64, n-len(c))...)
	}
	copy(c[i:], vs)
	m[f.Name] = c
}

// getText = ovr08:12FB: the text in r, up to its NUL.
func (e *Engine) getText(r ecl.Ref) string {
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
func (e *Engine) setText(r ecl.Ref, s string) {
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

func (e *Engine) textField(f *ecl.Field) {
	if f.Width != ecl.WidthText {
		panic(fmt.Sprintf("script: %s is no text field", f.Name))
	}
}

// savedWord is the game's word name (a field's element 0, or a named
// element) in any area (the quest log reads the quest table wherever the
// party is).
func (e *Engine) savedWord(name string) int64 {
	if ef, ok := e.engineFields[name]; ok && ef.get != nil {
		if v, ok := ef.get(0); ok {
			return v
		}
	}
	f, i := e.gameWord(name)
	if c := e.vars[f]; i < len(c) {
		return c[i]
	}
	return 0
}

// setSavedWord sets the game's word name, as a script would.
func (e *Engine) setSavedWord(name string, v int64) {
	if ef, ok := e.engineFields[name]; ok && !ef.stored {
		ef.set(0, v)
		return
	}
	f, i := e.gameWord(name)
	c := e.vars[f]
	if len(c) <= i {
		c = append(c, make([]int64, i+1-len(c))...)
	}
	c[i] = v
	e.vars[f] = c
}

// gameWord is the saved field and element the game's word name is: a field
// (element 0) or a field's named element. The engine's own words have none.
func (e *Engine) gameWord(name string) (field string, i int) {
	if f, i, ok := wordIn(e.gameFields(), name); ok {
		return f, i
	}
	if _, ok := e.engineFields[name]; ok {
		return name, 0
	}
	panic("script: no game field " + name)
}

// wordIn finds the word name in fields: a field's element 0 or a named element.
func wordIn(fields []ecl.Field, name string) (field string, i int, ok bool) {
	for _, f := range fields {
		if f.Name == name {
			return f.Name, 0, true
		}
		for k, n := range f.Elements {
			if n == name {
				return f.Name, k, true
			}
		}
	}
	return "", 0, false
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

// selCoin is the fact of SelCoins element i: copper (0) and gold (6), the
// money the import leaves the scripts (ecl.GoldCoins), then the unappraised
// gems (10) and jewelry (12), items. Any other element is no field.
func selCoin(i int) (rules.Fact, bool) {
	switch i {
	case 0:
		return rules.CoinFact{Coin: rules.CoinCopper}, true
	case 6:
		return rules.CoinFact{Coin: rules.CoinGold}, true
	case 10:
		return rules.PiecesFact{}, true
	case 12:
		return rules.PiecesFact{Jewel: true}, true
	}
	return nil, false
}

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
func cutToWidth(w ecl.Width, v int64) int64 {
	switch w {
	case ecl.WidthByte, ecl.WidthWord, ecl.WidthText:
		return int64(uint16(v))
	case ecl.WidthDword:
		return int64(uint32(v))
	}
	panic(fmt.Sprintf("script: width %d", w))
}

// cutToOperand is v as the operand of width w stores it: that many bytes (BR ecl.go PutN).
func cutToOperand(w ecl.Width, v int64) int64 {
	switch w {
	case ecl.WidthByte:
		return int64(uint8(v))
	case ecl.WidthWord:
		return int64(uint16(v))
	case ecl.WidthDword:
		return int64(uint32(v))
	}
	panic(fmt.Sprintf("script: operand width %d", w))
}

// readAsWidth is the cell value v as an operand of width w reads it: zero-extended for a byte
// or word, a signed 32-bit value for a long (BR ecl.go GetN).
func readAsWidth(w ecl.Width, v int64) int64 {
	if w == ecl.WidthDword {
		return int64(int32(v))
	}
	return cutToOperand(w, v)
}

// byteSpan is the bytes an access through r takes: its operand's width (a byte for the engine's
// own); an access past the buffer's end is a layout fault.
func (e *Engine) byteSpan(f *ecl.Field, r ecl.Ref) int {
	n := 1
	switch r.As {
	case ecl.WidthWord:
		n = 2
	case ecl.WidthDword:
		n = 4
	}
	if r.Index < 0 || r.Index+n > f.Len {
		panic(fmt.Sprintf("script: %s[%d..%d] is past the buffer's %d bytes (layout too short)", f.Name, r.Index, r.Index+n-1, f.Len))
	}
	return n
}

// getBytes reads r's width of bytes from byte buffer f, little-endian.
func (e *Engine) getBytes(f *ecl.Field, r ecl.Ref) int64 {
	n := e.byteSpan(f, r)
	c := e.cells(f)[f.Name]
	var v int64
	for k := n - 1; k >= 0; k-- {
		v <<= 8
		if r.Index+k < len(c) {
			v |= c[r.Index+k] & 0xFF
		}
	}
	return readAsWidth(widthOfBytes(n), v)
}

// setBytes writes v's r-width bytes into byte buffer f, little-endian.
func (e *Engine) setBytes(f *ecl.Field, r ecl.Ref, v int64) {
	n := e.byteSpan(f, r)
	vs := make([]int64, n)
	for k := range vs {
		vs[k] = (v >> (8 * k)) & 0xFF
	}
	e.storeCells(f, r.Index, vs)
}

// widthOfBytes is the operand width that takes n bytes.
func widthOfBytes(n int) ecl.Width {
	switch n {
	case 2:
		return ecl.WidthWord
	case 4:
		return ecl.WidthDword
	}
	return ecl.WidthByte
}

// TableFault is a read of a script's own table past the field the layout
// gives it: the index's range is not declared (layout: range=N on the
// variable) or the tables that run into one another are not one field.
type TableFault struct {
	Area       int
	Table      string
	Index, Len int
}

func (f TableFault) Error() string {
	return fmt.Sprintf("script: area %d: %s[%d] is past the field's %d entries", f.Area, f.Table, f.Index, f.Len)
}

// tableFault notes the fault, once each (TableFaults, and the log): the
// game goes on as the original does, which reads the code byte at base plus
// index whatever table it lies in; the fault says the layout is to be mended.
func (e *Engine) tableFault(f *ecl.Field, i int) {
	t := TableFault{Area: e.vm.Script.Area, Table: f.Name, Index: i, Len: f.Len}
	if slices.Contains(e.ext.tableFaults, t) {
		return
	}
	e.ext.tableFaults = append(e.ext.tableFaults, t)
	log.Print(t.Error())
}

// TableFaults are the table faults since the engine began, in their order.
func (e *Engine) TableFaults() []TableFault { return slices.Clone(e.ext.tableFaults) }

func (e *Engine) inField(f *ecl.Field, i int) {
	if i < 0 || i >= f.Len {
		panic(fmt.Sprintf("script: %s[%d] is past the field's %d (layout too short)", f.Name, i, f.Len))
	}
}

// wordField is the script field of the ruleset's word n.
func (e *Engine) wordField(n rules.WorldWord) engineField {
	return engineField{
		get: func(int) (int64, bool) { return int64(e.worldWord(n)), true },
		set: func(_ int, v int64) { e.setWorldWord(n, int(v)) },
	}
}

// effect is entry i of the party's affect table.
func (e *Engine) effect(i int) (uint16, bool) {
	a := e.ruleset().Check(rules.WordIs{Word: rules.WordEffect, Index: i}, nil)
	return uint16(a.Value), a.Yes
}

// partyIndex is c's place in the party, the party's size when it is not in it (ovr08:0756).
func (e *Engine) partyIndex(id rules.CharID) int {
	ids := e.partyIDs()
	if i := slices.Index(ids, id); i >= 0 {
		return i
	}
	return len(ids)
}

// scriptWords are the plain record words a script reads and writes, by the
// name its layout gives them (data -> the typed word, once, here).
var scriptWords = map[string]rules.WorldWord{
	"MapX":         rules.WordMapX,
	"MapY":         rules.WordMapY,
	"Geo":          rules.WordGeo,
	"TenMinutes":   rules.WordTenMinutes,
	"Hours":        rules.WordHours,
	"NoMagic":      rules.WordNoMagic,
	"In3D":         rules.WordIn3D,
	"WallSplitA":   rules.WordWallSplitA,
	"WallSplitB":   rules.WordWallSplitB,
	"PrevX":        rules.WordPrevX,
	"PrevY":        rules.WordPrevY,
	"AreaKind":     rules.WordAreaKind,
	"NoCoords":     rules.WordNoCoords,
	"MsgDelay":     rules.WordMsgDelay,
	"ViewFlags":    rules.WordViewFlags,
	"MarkFlags4":   rules.WordMarkFlags4,
	"Mark3":        rules.WordMark3,
	"Mark4Off":     rules.WordMark4Off,
	"MarkFlags57":  rules.WordMarkFlags57,
	"RiverOpen":    rules.WordRiverOpen,
	"Boss":         rules.WordBoss,
	"Mark4":        rules.WordMark4,
	"PartyExp":     rules.WordTrainable,
	"MaxDist":      rules.WordMaxDist,
	"MonsterDist":  rules.WordMonsterDist,
	"Morale":       rules.WordMorale,
	"CombatEnd":    rules.WordCombatEnd,
	"Beaten":       rules.WordBeaten,
	"HoldMove":     rules.WordHoldMove,
	"Search":       rules.WordSearch,
	"Surprise":     rules.WordSurprise,
	"RestEvery":    rules.WordRestEvery,
	"RestChance":   rules.WordRestChance,
	"LeftMap":      rules.WordLeftMap,
	"Portrait":     rules.WordPortrait,
	"Temple":       rules.WordTemple,
	"NoLoot":       rules.WordNoLoot,
	"Duel":         rules.WordDuel,
	"AttackedAlly": rules.WordAttackedAlly,
	"Shop":         rules.WordShop,
	"MonsterHit":   rules.WordMonsterHit,
	"PartyHit":     rules.WordPartyHit,
	"PartyMove":    rules.WordPartyMove,
}

// portraitWords are the sky words: a write redraws the portrait.
var portraitWords = map[string]rules.WorldWord{"Sky": rules.WordSky, "SkyInside": rules.WordSkyInside}
