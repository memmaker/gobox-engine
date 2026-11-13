package rules1e

import (
	"encoding/json"
	"fmt"
	"gobox/rules"
)

// The session: init, new game, save records and character files (core/session.go).

// resolveSession answers the Actions of the session topic.
func (r *Rules) resolveSession(a rules.Action) (rules.Outcome, bool, error) {
	s := r.S
	switch a := a.(type) {
	case rules.InitRules:
		s.SpellTableInit() // ovr23:3413; ovr13:31A2's affect handler table is Tables.Affects (callAffect)
	case rules.SetSeed:
		s.RandSeed = a.Seed
	case rules.LoadItemTable:
		s.Items = r.ItemTable()
	case rules.NewAdventure:
		s.Game, s.Area = GameVars{}, AreaVars{} // DS:49D6, 49D2
		s.Purse, s.Loot, s.Pool = Purse{}, Purse{}, nil
	case rules.EndSession:
		s.Camp.On, s.Training, s.Learning = false, false, false
	case rules.RecordView:
		r.SetSaveView(SaveView{AreaKind: a.AreaKind, WallsA: a.WallsA, WallsB: a.WallsB})
	case rules.TakeSavedWalls:
		w := rules.SavedWalls{A: s.Area.WallsA, B: s.Area.WallsB}
		s.Area.WallsA, s.Area.WallsB = [3]uint16{}, [3]uint16{}
		return rules.Outcome{Effects: []rules.Effect{w}}, true, nil
	default:
		return rules.Outcome{}, false, nil
	}
	return rules.Outcome{}, true, nil
}

// CharacterJSON is a character between adventures: its record's fields and
// the extras saved beside it. core.CharacterFile holds it as JSON; gobox/char
// reads and writes the files (.cha/.sav, .itm, .spc, .mem, .dual, .pow,
// .png, .icon.png).
type CharacterJSON struct {
	Sheet          Sheet
	PoolIDs        PoolIDs
	Portrait, Icon rules.PNG
	Plan           *[MemorySlots]SpellID // nil: none saved (the spells held are the plan)
	OldLevels      [8]Level
	Powers         []int
	Surges         int
	Laid           int
	Tough          int
	Policies       map[int]Policy
	Items          []Item
	Affects        []AffectRecord
	// Pool is the party's shared pack (Export, of the party's first member):
	// a game save's items file carries it after the member's own.
	Pool []Item
	// The later layouts' fields (gobox/char's Pack422 and siblings).
	MulticlassLevel Level
	CanTrain        bool
	Kept            *KeptRecord
}

// KeptRecord is a COTAB/SotSB/POD record as read (core.RecordLayout bytes):
// the fields rules1e does not model ride along so the codec writes the file
// back byte-equal; the modelled ones are written over it from the Sheet.
type KeptRecord struct {
	Layout rules.RecordLayout
	Image  rules.RecordImage
}

// Import = por LoadCharacter's rules part (ovr18:0AEC): the record, the
// items and affects mended as ImportCharacter does for the member WordWho
// names, then the extras.
func (r *Rules) Import(b rules.CharacterFile) (rules.CharID, error) {
	var f CharacterJSON
	if err := json.Unmarshal(b, &f); err != nil {
		return rules.NoChar, fmt.Errorf("Import: %w", err)
	}
	if err := r.S.toIDs(f.Affects, f.Items); err != nil {
		return rules.NoChar, fmt.Errorf("Import: %w", err)
	}
	if err := sheetSpells(&f.Sheet, f.Plan, r.S.T().spellNumber); err != nil {
		return rules.NoChar, fmt.Errorf("Import: %w", err)
	}
	c := &Character{Sheet: f.Sheet, PoolIDs: f.PoolIDs}
	for _, it := range f.Items {
		c.Items = append(c.Items, &it)
	}
	for _, a := range f.Affects {
		c.Affects = append(c.Affects, &Affect{AffectRecord: a})
	}
	r.S.ImportCharacter(c)
	c.Portrait, c.Icon, c.OldLevels = f.Portrait, f.Icon, f.OldLevels
	c.MulticlassLevel, c.CanTrain, c.Kept = f.MulticlassLevel, f.CanTrain, f.Kept
	c.Powers, c.Surges, c.Laid, c.Tough = f.Powers, f.Surges, f.Laid, f.Tough
	for id, p := range f.Policies {
		c.SetPolicy(id, p)
	}
	if f.Plan != nil {
		c.Plan = *f.Plan
	} else {
		for i, v := range c.Memorized {
			c.Plan[i] = v.Spell()
		}
	}
	return r.ID(c), nil
}

// Export is the character id as Import reads it.
func (r *Rules) Export(id rules.CharID) (rules.CharacterFile, error) {
	c := r.chars[id]
	if c == nil {
		return nil, fmt.Errorf("Export %d: no character", id)
	}
	f := CharacterJSON{Sheet: c.Sheet, PoolIDs: c.PoolIDs, Portrait: c.Portrait, Icon: c.Icon, Plan: &c.Plan, OldLevels: c.OldLevels,
		Powers: c.Powers, Surges: c.Surges, Laid: c.Laid, Tough: c.Tough, Policies: c.Policies,
		MulticlassLevel: c.MulticlassLevel, CanTrain: c.CanTrain, Kept: c.Kept}
	for _, it := range c.Items {
		f.Items = append(f.Items, *it)
	}
	for _, a := range c.Affects {
		f.Affects = append(f.Affects, a.AffectRecord)
	}
	if len(r.S.Party) > 0 && c == r.S.Party[0] {
		for _, it := range r.S.Pool {
			f.Pool = append(f.Pool, *it)
		}
	}
	if err := r.S.toGame(f.Affects, f.Items); err != nil {
		return nil, fmt.Errorf("Export %d: %w", id, err)
	}
	if err := r.S.toGame(nil, f.Pool); err != nil {
		return nil, fmt.Errorf("Export %d: %w", id, err)
	}
	plan := c.Plan
	f.Plan = &plan
	if err := sheetSpells(&f.Sheet, f.Plan, r.S.T().gameSpellNumber); err != nil {
		return nil, fmt.Errorf("Export %d: %w", id, err)
	}
	return json.Marshal(f)
}

// TakePending takes the effects queued outside Resolve (core.Pending).
func (r *Rules) TakePending() []rules.Effect {
	fx := r.S.Fx
	r.S.Fx = nil
	return fx
}
