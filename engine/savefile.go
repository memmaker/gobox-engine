//go:build fixture

package engine

import (
	"fmt"
	"os"
	"slices"

	"gobox/adventure"
	"gobox/ecl"
	"gobox/rules"
)

// SaveEditor edits a slot's canonical save outside the game: cmd/fixture
// builds the trace scenarios' save fixtures with it.
type SaveEditor struct {
	e *Engine // the game's fields, by name (SetVar)
	g adventure.Save
}

// OpenSave reads the canonical save of the slot folder dir; adv is the adventure folder, whose script
// layouts name the fields SetVar sets; rules builds the ruleset.
func OpenSave(ruleset Ruleset, dir, adv string) (*SaveEditor, error) {
	e := &Engine{newRules: ruleset, adv: adv}
	e.initVM()
	e.Cfg.SaveDir = dir
	g, err := e.readSlotSave()
	if err != nil {
		return nil, err
	}
	if err := e.ruleset().Load(g.Rules); err != nil {
		return nil, err
	}
	return &SaveEditor{e: e, g: g}, nil
}

// File is the save's file name in a slot.
func (s *SaveEditor) File() string { return adventure.GameFile }

// Bytes is the edited save.
func (s *SaveEditor) Bytes() []byte {
	blob, err := s.e.ruleset().Save()
	if err != nil {
		panic(err) // the ruleset's part always marshals
	}
	s.g.Rules = blob
	b, err := adventure.MarshalSave(s.g)
	if err != nil {
		panic(err) // a adventure.Save always marshals
	}
	return b
}

// SetVar sets the game's word name (layout/: a field's element 0 or a named element), as a script would.
func (s *SaveEditor) SetVar(name string, v uint16) error {
	fields, err := ecl.GameFields(os.DirFS(s.e.scriptDir()))
	if err != nil {
		return err
	}
	field, _, ok := wordIn(fields, name)
	i := slices.IndexFunc(fields, func(f ecl.Field) bool { return f.Name == field })
	if !ok {
		return fmt.Errorf("no game field %s", name)
	}
	if fields[i].Owner == ecl.OwnerMember {
		return fmt.Errorf("%s is a member's, not the save's", name)
	}
	e := s.e
	e.vars = s.g.Fields
	if e.vars == nil {
		e.vars = fieldCells{}
	}
	e.setSavedWord(name, int64(v))
	s.g.Fields = e.vars
	return nil
}

// EnterArea moves the save into ECL block of disk, loading 3D map geo and
// wall slots 1..3 from WALLDEF blocks walls (0xFF none), as a script's
// opNewECL would (ovr04:0BD1) and SaveGame records it. The load reloads the
// block's script, so ecl is not kept.
func (s *SaveEditor) EnterArea(disk, block, geo int, _ []byte, walls [3]int) {
	g := &s.g
	g.Set = rules.Set(disk)
	s.e.setWorldWord(rules.WordSymbolSet, disk)
	s.e.setWorldWord(rules.WordGeo, geo)
	v := rules.RecordView{AreaKind: uint16(block)}
	for i, id := range walls {
		blk, slot := id, i+1
		if id == 0xFF {
			blk, slot = -1, -1
		}
		v.WallsA[i], v.WallsB[i] = uint16(blk), uint16(slot)
	}
	s.e.do(v)
	g.Locals = fieldCells{} // the area's locals, cleared on entry (InitAreaScript)
}

// SetPos places the party; a negative value keeps the saved one.
func (s *SaveEditor) SetPos(x, y, facing int) {
	for _, p := range []struct {
		v   int
		dst *byte
	}{{x, &s.g.X}, {y, &s.g.Y}, {facing, &s.g.Facing}} {
		if p.v >= 0 {
			*p.dst = byte(p.v)
		}
	}
}

// Party is the base names of the members' .SAV/.SPC/.ITM files.
func (s *SaveEditor) Party() []string { return s.g.Party }
