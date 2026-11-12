package char

import (
	"encoding/binary"
	"fmt"
	"io"
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

// SaveLayout is a game's SAVGAMx.DAT format (its LoadGame/SaveGame), the
// import's data: a lead byte, the records, the code image, the position and
// states, the wall slots, then the party's count and file names.
type SaveLayout struct {
	Lead         SaveLead     // the byte before the records
	Records      []SaveRecord // in file order
	CodeBase     int          // the script address of the ECL code image (POR 0x9900, the later games 0x8000)
	CodeSize     int          // its bytes (0x1E00 in all four)
	CodeSaved    bool         // the code image follows the records in the save (POR, cotab)
	WallBytes    int          // the wall slots after the position and states (POR none)
	AreaKindAt   *int         // the script address of the record word naming the area's locals layout (POR 0x49F2); nil: the lead's, or none
	RulesRecords bool         // the records are rules1e's AreaVars, GameVars, AuxVars (POR): the ruleset's state is read from them
}

// SaveLead is what a save's first byte is.
type SaveLead int

const (
	LeadNone     SaveLead = iota // no lead byte (pod)
	LeadSet                      // the symbol set (POR)
	LeadAreaKind                 // the area (sotsb, cotab)
)

// SaveRecord is one record of a save: Cells cells of CellBytes each, the
// first at script address Base.
type SaveRecord struct{ Base, Cells, CellBytes int }

// RecordCells are a save record's cells, the first at script address Base.
type RecordCells struct {
	Base  int
	Cells []uint16
}

// ImportedSave is an original save read: the ruleset's canonical part (Rules,
// for core.Rules Load), what the engine keeps (positions, the party's file
// names as stored) and the record cells by script address, for the script
// fields that live in the records.
type ImportedSave struct {
	Rules                    []byte
	Set                      byte
	X, Y, Facing, Pos3, Pos4 byte
	ViewMode, BeginFlag      byte
	AreaKind                 uint16
	Party                    []string
	Records                  []RecordCells
}

// Cell is the record cell at script address a, nil outside the records.
func (o *ImportedSave) Cell(a int) *uint16 {
	for i := range o.Records {
		r := &o.Records[i]
		if a >= r.Base && a < r.Base+len(r.Cells) {
			return &r.Cells[a-r.Base]
		}
	}
	return nil
}

// ReadOriginalSave reads an original SAVGAMx.DAT image laid out as l: the
// canonical blob, the engine's parts and the record cells.
func ReadOriginalSave(b []byte, l SaveLayout) (ImportedSave, error) {
	if len(l.Records) == 0 {
		return ImportedSave{}, fmt.Errorf("%s: the game has no original save layout (gametables.json OriginalSave)", OriginalSaveFile)
	}
	p := 0
	short := fmt.Errorf("%s: %w", OriginalSaveFile, io.ErrUnexpectedEOF)
	read := func(n int) []byte {
		if p+n > len(b) {
			return nil
		}
		r := b[p : p+n]
		p += n
		return r
	}
	var o ImportedSave
	if l.Lead != LeadNone {
		v := read(1)
		if v == nil {
			return o, short
		}
		if l.Lead == LeadSet {
			o.Set = v[0]
		} else {
			o.AreaKind = uint16(v[0])
		}
	}
	for _, r := range l.Records {
		raw := read(r.Cells * r.CellBytes)
		if raw == nil {
			return o, short
		}
		c := RecordCells{Base: r.Base, Cells: make([]uint16, r.Cells)}
		for i := range c.Cells {
			switch r.CellBytes {
			case 1:
				c.Cells[i] = uint16(raw[i])
			case 2:
				c.Cells[i] = binary.LittleEndian.Uint16(raw[2*i:])
			default:
				return o, fmt.Errorf("%s: a record cell of %d bytes", OriginalSaveFile, r.CellBytes)
			}
		}
		o.Records = append(o.Records, c)
	}
	code := 0
	if l.CodeSaved {
		code = l.CodeSize
	}
	tail := read(code + 5 + 2 + l.WallBytes + 1 + len(PascalName{})*8)
	if tail == nil {
		return o, short
	}
	t := tail[code:]
	o.X, o.Y, o.Facing, o.Pos3, o.Pos4, o.ViewMode, o.BeginFlag = t[0], t[1], t[2], t[3], t[4], t[5], t[6]
	t = t[7+l.WallBytes:]
	count := int(t[0])
	var names [8]PascalName
	for i := range names {
		copy(names[i][:], t[1+i*len(PascalName{}):])
	}
	if l.AreaKindAt != nil {
		k := o.Cell(*l.AreaKindAt)
		if k == nil {
			return o, fmt.Errorf("%s: AreaKindAt %#x is in no record", OriginalSaveFile, *l.AreaKindAt)
		}
		o.AreaKind = *k
	}
	s := &rules1e.RuleState{Purse: ReadPurse(b[p:])}
	if l.RulesRecords {
		if len(o.Records) != 3 || len(o.Records[0].Cells) != rules1e.AreaWords || len(o.Records[1].Cells) != rules1e.GameWords || len(o.Records[2].Cells) != len(rules1e.AuxVars{}) {
			return o, fmt.Errorf("%s: RulesRecords needs the area, game and aux records", OriginalSaveFile)
		}
		for i, v := range o.Records[0].Cells {
			*s.Area.Word(i) = v
		}
		for i, v := range o.Records[1].Cells {
			*s.Game.Word(i) = v
		}
		s.Area.ClearScriptWords()
		s.Game.ClearScriptWords()
	}
	r := &rules1e.Rules{S: s}
	var err error
	if o.Rules, err = r.Save(); err != nil {
		return o, err
	}
	for i := range min(count, len(names)) {
		o.Party = append(o.Party, names[i].String())
	}
	return o, nil
}

// OriginalSaveFile is the original's save of a slot (ovr18 SaveGame).
const OriginalSaveFile = "SAVGAMA.DAT"
