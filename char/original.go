package char

import (
	"encoding/binary"
	"fmt"
	"math"
	"rules1e"
)

// OriginalSave is the fixed part of SAVGAMx.DAT (ovr18 LoadGame/SaveGame),
// read with encoding/binary; the party's purse follows it (ReadPurse).
type OriginalSave struct {
	Set                      byte
	Area                     rules1e.AreaVars
	Game                     rules1e.GameVars
	Aux                      rules1e.AuxVars
	ECL                      [0x1E00]byte
	X, Y, Facing, Pos3, Pos4 byte
	ViewMode, BeginFlag      byte
	Count                    byte          // party members
	Names                    [8]PascalName // their CHRDAT<key><n> files
}

// PascalName is a Turbo Pascal string[40]: length byte, then the text.
type PascalName [0x29]byte

func (p *PascalName) String() string { return string(p[1 : 1+min(int(p[0]), 40)]) }

func (p *PascalName) Set(s string) {
	*p = PascalName{byte(min(len(s), 40))}
	copy(p[1:], s)
}

// PurseSize is AppendPurse's length: gold, gems, jewelry, copper.
const PurseSize = 16

// ReadPurse reads AppendPurse's format; an older save's purse is gold, gems
// and jewelry only, before copper came; shorter than that, none.
func ReadPurse(b []byte) rules1e.Purse {
	v := func(i int) uint64 { return uint64(binary.LittleEndian.Uint32(b[4*i:])) }
	switch {
	case len(b) >= PurseSize:
		return rules1e.Purse{Gold: v(0), Gems: v(1), Jewelry: v(2), Copper: v(3)}
	case len(b) >= 12:
		return rules1e.Purse{Gold: v(0), Gems: v(1), Jewelry: v(2)}
	}
	return rules1e.Purse{}
}

// AppendPurse keeps the purse after the original save game data.
// ponytail: the format holds uint32, so a field above 4294967295 is saved as
// that (the byte format is unchanged); widen the format when it matters.
func AppendPurse(b []byte, p rules1e.Purse) []byte {
	for _, v := range []uint64{p.Gold, p.Gems, p.Jewelry, p.Copper} {
		b = binary.LittleEndian.AppendUint32(b, uint32(min(v, math.MaxUint32)))
	}
	return b
}

// originalSaveFile is the file ImportOriginal reads (ovr18 SaveGame).
const originalSaveFile = "SAVGAMA.DAT"

// ImportedSave is an original save read: the ruleset's canonical part (Rules,
// for core.Rules Load), what the engine keeps (positions, the party's file
// names as stored) and the record words by index, for the script fields that
// live in the records (the engine knows each record's address base).
type ImportedSave struct {
	Rules                    []byte
	Set                      byte
	X, Y, Facing, Pos3, Pos4 byte
	ViewMode, BeginFlag      byte
	AreaKind                 uint16
	Party                    []string
	Area, Game, Aux          []uint16
}

// recordWords are the words of a record, by index.
func recordWords(n int, word func(int) *uint16) []uint16 {
	w := make([]uint16, n)
	for i := range w {
		w[i] = *word(i)
	}
	return w
}

// ReadOriginalSave reads an original SAVGAMA.DAT image: the canonical blob,
// the engine's parts and the record words.
func ReadOriginalSave(b []byte) (ImportedSave, error) {
	var f OriginalSave
	n, err := binary.Decode(b, binary.LittleEndian, &f)
	if err != nil {
		return ImportedSave{}, fmt.Errorf("%s: %w", OriginalSaveFile, err)
	}
	o := ImportedSave{
		Set: f.Set, X: f.X, Y: f.Y, Facing: f.Facing, Pos3: f.Pos3, Pos4: f.Pos4,
		ViewMode: f.ViewMode, BeginFlag: f.BeginFlag, AreaKind: f.Area.AreaKind,
		Area: recordWords(rules1e.AreaWords, f.Area.Word),
		Game: recordWords(rules1e.GameWords, f.Game.Word),
		Aux:  f.Aux[:],
	}
	r := &rules1e.Rules{S: &rules1e.RuleState{Area: f.Area, Game: f.Game, Purse: ReadPurse(b[n:])}}
	r.S.Area.ClearScriptWords()
	r.S.Game.ClearScriptWords()
	if o.Rules, err = r.Save(); err != nil {
		return o, err
	}
	for i := range min(int(f.Count), len(f.Names)) {
		o.Party = append(o.Party, f.Names[i].String())
	}
	return o, nil
}

// OriginalSaveFile is the original's save of a slot (ovr18 SaveGame).
const OriginalSaveFile = "SAVGAMA.DAT"
