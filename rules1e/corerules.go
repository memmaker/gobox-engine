package rules1e

import (
	"encoding/json"
	"errors"
	"fmt"
	"gobox/rules"
	"slices"
)

// Rules is the 1st edition ruleset as core.Rules, over the RuleState the por
// Engine holds (fold step E-R6). Each method names the por call it stands for;
// a method whose call still lives in the Engine (gobox-engine/pool *_engine.go)
// answers ErrNotPorted until that concern moves here.
type Rules struct {
	S        *RuleState
	campaign rules.Campaign
	options  rules.Options
	ids      map[*Character]rules.CharID
	chars    map[rules.CharID]*Character
	next     rules.CharID
	handles  map[*Item]rules.ItemHandle
	items    map[rules.ItemHandle]*Item
	nextItem rules.ItemHandle
	valued   *Item // the appraised gem or jewel, kept or sold next (core.Valuate)
	// Monsters is the current data set's monster records (por LoadMonster's
	// MON<n>CHA/SPC/ITM), filled by whoever holds the adventure's files.
	Monsters func(rules.MonsterID) (Monster, bool)
	// FirstMonster is the lowest monster id of the current data set (false: none).
	FirstMonster func() (int, bool)
	// ItemTable is the base item records (por's ITEMS file), read by core.LoadItemTable.
	ItemTable func() []ItemBase
	// TreasureBlock is the data set's item block n (empty: none), read by core.AddTreasureBlock.
	TreasureBlock func(n int) []Item
	// ItemRecords is items as the game's file records (core.TreasureRecordsFact); the rules hold no bytes.
	ItemRecords func([]Item) rules.ItemRecords
	// Host is the engine's side of the creation (Creation), set by whoever holds the roster and picture files.
	Host     CreationHost
	held     map[rules.CharID]Character // HoldCopy's copies
	tmpl     *Character                 // the group's first monster, which Spawn copies
	empty    *Character                 // the placeholder EmptyMember empties (core.EmptyMember)
	tmplID   rules.MonsterID
	view     *SaveView // SetSaveView
	straight bool      // BeginFight's Board.Straight: the board is deskewed
	boss     uint16    // buildGround's BuildGround.Boss, while it runs
}

// ID is c's CharID, given at its first use and kept while c lives. Until the
// Engine holds CharIDs itself, it names its characters through ID.
func (r *Rules) ID(c *Character) rules.CharID {
	if c == nil {
		return rules.NoChar
	}
	if id, ok := r.ids[c]; ok {
		return id
	}
	if r.ids == nil {
		r.ids, r.chars = map[*Character]rules.CharID{}, map[rules.CharID]*Character{}
	}
	r.next++
	id := r.next
	r.ids[c], r.chars[id] = id, c
	return id
}

var errNoChar = errors.New("rules1e: no such character")
var errWoundNotNeeded = errors.New("rules1e: a wound cure for one who does not need it")
var errNoBlock = errors.New("rules1e: no such item block")
var errNotFight = errors.New("rules1e: not in a fight")
var errNoFact = errors.New("rules1e: fact not held")
var errBadValue = errors.New("rules1e: record value outside its table")

var _ rules.Rules = (*Rules)(nil)

// ErrNotPorted is a core.Rules method whose por code is still in the Engine.
var ErrNotPorted = errors.New("rules1e: not behind core.Rules yet")

// Catalog: what the records number (the engine reads the tables by place): the races and alignments in record order, the saving throws in sheet order.
func (r *Rules) Catalog() rules.Catalog {
	return rules.Catalog{Races: RecordRace[:], Alignments: RecordAlignment[:], Saves: SaveOrder[:]}
}

func (r *Rules) Begin(c rules.Campaign, o rules.Options) { r.campaign, r.options = c, o }

// Ships: 1st edition has no ship fights.
func (r *Rules) Ships() rules.ShipRules { return nil }

func (r *Rules) Allowed(sit rules.Situation) rules.Allowed {
	if f, ok := sit.(rules.Fight); ok {
		return r.fightAllowed(f)
	}
	if t, ok := sit.(rules.Training); ok {
		return r.classOffers(t)
	}
	if u, ok := sit.(rules.UsingItem); ok {
		return r.usingItem(u)
	}
	return rules.Allowed{}
}
func (r *Rules) Resolve(sit rules.Situation, a rules.Action, _ *rules.RNG) (rules.Outcome, error) {
	if u, ok := a.(rules.UseItem); ok && !inFight(sit) {
		return r.useItemOutside(u)
	}
	if o, ok, err := r.resolveItems(sit, a); ok {
		return o, err
	}
	if o, ok, err := r.resolveParty(a); ok {
		return o, err
	}
	if o, ok, err := r.resolveSheet(a); ok {
		return o, err
	}
	if o, ok, err := r.resolveFight(sit, a); ok {
		return o, err
	}
	if o, ok, err := r.resolveSpells(sit, a); ok {
		return o, err
	}
	if o, ok, err := r.resolveCamp(a); ok {
		return o, err
	}
	if o, ok, err := r.resolveTown(a); ok {
		return o, err
	}
	if o, ok, err := r.resolveSession(a); ok {
		return o, err
	}
	if o, ok, err := r.resolveWorld(sit, a); ok {
		return o, err
	}
	return rules.Outcome{}, ErrNotPorted
}

// setStatus is SpecialWriteZ1 0x100: every status but okay takes the member
// out of the active ones (Okay 0); stoned (0x87) also stones; StatusInactive
// is the script's other numbers >= 0x80 (the record's status kept).
func (r *Rules) setStatus(a rules.SetStatusAction) error {
	c := r.chars[a.Who]
	if c == nil {
		return errNoChar
	}
	switch a.Status {
	case rules.StatusOkay: // revived: okay and in play (sotsb, pod SpecialWriteZ1 $100; no POR script writes it)
		c.Okay, c.Status = 1, StatusOkay
	case rules.StatusStoned:
		c.Okay, c.Status = 0, StatusStoned
	case rules.StatusInactive:
		c.Okay = 0
	default:
		return errBadValue
	}
	return nil
}

// Elapse: por stepRest (camp.go), memTick.
func (r *Rules) Elapse(_ rules.Party, minutes int32, _ *rules.RNG) (o rules.Outcome) {
	return r.elapse(minutes)
}

// SaveBlob is the rules' part of a canonical save (por GAME.JSON's Rules).
type SaveBlob struct {
	Area  AreaVars // the named words only (vars.go)
	Game  GameVars
	Purse Purse
}

// SaveView is the engine's view state a save records on Area (por saveSlot):
// the wall slots, which the live Area keeps cleared.
type SaveView struct {
	WallsA, WallsB [3]uint16
}

// SetSaveView is the view the next Saves record.
func (r *Rules) SetSaveView(v SaveView) { r.view = &v }

// Save, Load: por saveSlot/loadSlot (files_engine.go).
func (r *Rules) Save() ([]byte, error) {
	b := SaveBlob{Area: r.S.Area, Game: r.S.Game, Purse: r.S.Purse}
	if v := r.view; v != nil {
		b.Area.WallsA, b.Area.WallsB = v.WallsA, v.WallsB
	}
	return json.Marshal(b)
}

func (r *Rules) Load(b []byte) error {
	var s SaveBlob
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	r.S.Area, r.S.Game, r.S.Purse = s.Area, s.Game, s.Purse
	return nil
}

// Spawn is por addMonsters (ovr04:055B): the group's first monster is loaded
// and rolled (MakeMonster, RuleState's dice: the *core.RNG is unused so POR's
// dice stay); each later one of the same id is a byte copy of it, its item and
// affect lists reversed (ovr04:063E..07E7). It joins S.Party.
func (r *Rules) Spawn(id rules.MonsterID, _ *rules.RNG) (rules.CharID, error) {
	c := &Character{}
	if t := r.tmpl; t != nil && r.tmplID == id {
		c.Sheet, c.Slots, c.Combat = t.Sheet, t.Slots, t.Combat
		for k := len(t.Items) - 1; k >= 0; k-- {
			it := *t.Items[k]
			c.Items = append(c.Items, &it)
		}
		for k := len(t.Affects) - 1; k >= 0; k-- {
			a := *t.Affects[k]
			c.Affects = append(c.Affects, &a)
		}
	} else {
		m, err := r.monster(id)
		if err != nil {
			return 0, fmt.Errorf("Spawn: %w", err)
		}
		r.S.MakeMonster(c, m)
		r.tmpl, r.tmplID = c, id
	}
	r.S.Party = append(r.S.Party, c)
	return r.ID(c), nil
}

var errNoMonster = errors.New("rules1e: no such monster")

// monster is monster id's records in the rules' ids: the game's race, spell and affect
// numbers mapped as Import maps a character's (the one way a monster is read).
func (r *Rules) monster(id rules.MonsterID) (Monster, error) {
	if r.Monsters == nil {
		return Monster{}, errNoMonster
	}
	m, ok := r.Monsters(id)
	if !ok {
		return Monster{}, fmt.Errorf("monster %d: %w", id, errNoMonster)
	}
	m.Affects, m.Items = slices.Clone(m.Affects), slices.Clone(m.Items) // the records are the game's: ids on a copy
	if err := r.S.toIDs(m.Affects, m.Items); err != nil {
		return Monster{}, fmt.Errorf("monster %d: %w", id, err)
	}
	if err := sheetSpells(&m.Sheet, nil, r.S.T().spellNumber); err != nil {
		return Monster{}, fmt.Errorf("monster %d: %w", id, err)
	}
	return m, nil
}

// Monster is a monster's records: MON<n>CHA as a Sheet, SPC, ITM.
type Monster struct {
	Sheet   Sheet
	Affects []AffectRecord
	Items   []Item
}

// MakeMonster = ovr18:0EC4's rules: c becomes m, icon size by race, icon
// colours rolled (RuleState's dice), the items mended (Tables.MendItem).
func (e *RuleState) MakeMonster(c *Character, m Monster) {
	c.Sheet = m.Sheet
	c.Items, c.Affects = nil, nil
	if n := iconSizeOf(c.Race); n != 0 {
		c.IconSize = n
	}
	for i := 1; i <= 3; i++ {
		r := byte(e.Dice(1, 6))
		c.IconColour[i-1] = Colours((r+8)<<4 | r)
	}
	c.IconColour[3] = 0xC4
	for i := 5; i <= 6; i++ {
		r := byte(e.Dice(1, 6))
		c.IconColour[i-1] = Colours((r+8)<<4 | r)
	}
	for _, a := range m.Affects {
		c.Affects = append(c.Affects, &Affect{AffectRecord: a})
	}
	for _, it := range m.Items {
		e.T().MendItem(&it)
		c.Items = append(c.Items, &it)
	}
}

// Item: the base table DS:5582 (RuleState.Items).
func (r *Rules) Item(rules.ItemRef) (rules.Item, error) { return rules.Item{}, ErrNotPorted }

// Dismiss is RemoveSelected's rules: id leaves S.Party and is forgotten.
func (r *Rules) Dismiss(id rules.CharID) {
	c := r.chars[id]
	if i := slices.Index(r.S.Party, c); c != nil && i >= 0 {
		r.S.Party = append(r.S.Party[:i], r.S.Party[i+1:]...) // as RemoveSelected: the tail stays in the backing array
	}
	delete(r.ids, c)
	delete(r.chars, id)
}

// abilities are core's names of Stats' indexes (StatStr..StatCha).
var abilities = [6]rules.Ability{rules.Strength, rules.Intelligence, rules.Wisdom, rules.Dexterity, rules.Constitution, rules.Charisma}

// Sheet is a character's facts for the screens: por members
// (combatflow_engine.go), playfacts.go. AC is as displayed (60 - the record's).
func (r *Rules) Sheet(id rules.CharID) rules.Sheet {
	c := r.chars[id]
	if c == nil {
		return rules.Sheet{}
	}
	s := rules.Sheet{Name: c.Name(), HP: int32(c.HP), MaxHP: int32(c.MaxHP), AC: 60 - int32(c.AC[0]), RearAC: 60 - int32(c.AC[1]),
		Exp: int64(int32(c.Exp)), Scores: map[rules.Ability]int32{rules.StrengthPercent: int32(c.Str00)}}
	for i, a := range abilities {
		s.Scores[a] = int32(c.Stats[i])
	}
	s.LevelExp, s.NextExp = levelSpan(c)
	r.sheetView(c, &s)
	if int(c.Status) < len(PorStatus) {
		s.Status = PorStatus[c.Status]
	}
	return s
}

// Draft: por porcreation.go (rules1e.Creation is the creation rules).
func (r *Rules) Draft() *rules.Draft { return nil }

// Purse: por purseCoin.
func (r *Rules) Purse(rules.CharID) rules.Copper { return 0 }

// Shared: the party inventory (RuleState.Pool).
func (r *Rules) Shared() []rules.Carried { return nil }

func (r *Rules) Spell(rules.SpellID) (rules.Spell, error) { return rules.Spell{}, ErrNotPorted }

// Fact, SetFact: the selected member's words of por's script (SpecialReadZ1,
// SpecialWriteZ1; por-go scriptvars_engine.go). Values are the record's own
// numbers; the script's encodings stay with the engine.
func (r *Rules) Fact(id rules.CharID, f rules.Fact) (rules.Value, error) {
	if v, ok, err := r.factItems(id, f); ok {
		return v, err
	}
	if v, ok, err := r.factTown(id, f); ok {
		return v, err
	}
	if v, ok, err := r.factSpells(id, f); ok {
		return v, err
	}
	if v, ok, err := r.factParty(id, f); ok {
		return v, err
	}
	if v, ok, err := r.factFight(id, f); ok {
		return v, err
	}
	if r.chars[id] == nil {
		return nil, errNoChar
	}
	return nil, errNoFact
}

func (r *Rules) SetFact(id rules.CharID, f rules.Fact, v rules.Value) error {
	c := r.chars[id]
	if c == nil {
		return errNoChar
	}
	if f == rules.FactSide {
		switch v {
		case rules.SideParty:
			c.Side, c.Auto = SideParty, 0
		case rules.SideAuto:
			c.Side, c.Auto = SideParty, 1
		case rules.SideMonsters:
			c.Side, c.Auto = SideMonsters, 1
		default:
			return errBadValue
		}
		return nil
	}
	if f == rules.FactSideOnly {
		switch v {
		case rules.SideParty:
			c.Side = SideParty
		case rules.SideMonsters:
			c.Side = SideMonsters
		default:
			return errBadValue
		}
		return nil
	}
	if f, ok := f.(rules.MemoFact); ok {
		return setMemo(r.S.T(), c, f.Slot, v)
	}
	if f == rules.FactName {
		t, ok := v.(rules.Text)
		if !ok {
			return errBadValue
		}
		c.SetName(string(t))
		return nil
	}
	n, ok := v.(rules.Number)
	if !ok {
		return errNoFact
	}
	if f, ok := f.(rules.PiecesFact); ok {
		*pursePieces(&r.S.Purse, f.Jewel) = uint64(uint16(n))
		return nil
	}
	if f, ok := f.(rules.CoinFact); ok {
		v, ok := purseCoin(&r.S.Purse, f.Coin)
		if !ok {
			return errNoFact
		}
		*v = uint64(uint16(n))
		return nil
	}
	switch f {
	case rules.FactExpValue:
		c.ExpValue = uint16(n)
	case rules.FactExpPerHP:
		c.ExpPerHP = byte(n)
	case rules.FactControl:
		c.Control = byte(n)
	case rules.FactAlignment: // the record's byte, as a read gives it
		c.Alignment = byte(n)
	default:
		return errNoFact
	}
	return nil
}

// Check: por opCheckParty, opPartyStrength, opPartySurprise.
func (r *Rules) Check(k rules.Check, _ *rules.RNG) rules.Answer {
	if a, ok := r.checkItems(k); ok {
		return a
	}
	if a, ok := r.checkParty(k); ok {
		return a
	}
	if a, ok := r.checkMonsters(k); ok {
		return a
	}
	if a, ok := r.checkFight(k); ok {
		return a
	}
	if a, ok := r.checkSpells(k); ok {
		return a
	}
	if a, ok := r.checkCamp(k); ok {
		return a
	}
	if a, ok := r.checkTown(k); ok {
		return a
	}
	return r.checkWorld(k)
}

// The fight: por combatSetup, nextCombatant, takeTurn, computerTurn, endTurn,
// sidesCheck, pushAfterCombat (combat*_engine.go).

// BeginRound = por stepFight's round start (ovr09): the sides counted,
// every combatant's TurnSetup (initiative, RuleState dice), surprise over.
// The actor (Who) is whoever the engine selected.
func (r *Rules) BeginRound(rules.Fight, *rules.RNG) rules.Outcome {
	r.S.CountSides()
	for _, c := range r.S.Party {
		r.S.TurnSetup(c)
	}
	r.S.Terms.Surprised = nil
	return rules.Outcome{}
}

// NextActor = por NextCombatant (ovr09:0131): the highest initiative left,
// RuleState's dice for the ties (the *core.RNG is not used: coordinator ruling).
func (r *Rules) NextActor(rules.Fight, *rules.RNG) (rules.CharID, bool) {
	c := r.S.NextCombatant()
	if c == nil {
		return rules.NoChar, false
	}
	return r.ID(c), true
}

// Char is the character id names, nil if none: the Engine's way back from
// a CharID until it holds ids itself.
func (r *Rules) Char(id rules.CharID) *Character { return r.chars[id] }

// Decide = por computerTurn: one AutoTurn, the whole turn played by Resolve
// (the engine fills in its Key).
func (r *Rules) Decide(rules.Fight, rules.CharID, *rules.RNG) rules.Action { return rules.AutoTurn{} }

// EndTurn = por endTurn's rules part: the rule set's TurnEnd (extra).
func (r *Rules) EndTurn(_ rules.Fight, id rules.CharID, _ *rules.RNG) rules.Outcome {
	if c := r.chars[id]; c != nil {
		r.S.Rules().TurnEnd(r.S, c)
	}
	return rules.Outcome{}
}

// EndRound = por endRound's rules part (ovr09): a round of the clock,
// affect group 0x13 and stinking cloud on everyone, the dying a round
// nearer death (dead after 10), a dying teammate called out (Bandage).
// The actor (Who) is whoever the engine selected.
func (r *Rules) EndRound(rules.Fight, *rules.RNG) rules.Outcome {
	s := r.S
	s.AdvanceClock(1, 1)
	s.CombatRound++
	s.FoesHealthPercent()
	for _, c := range s.Party {
		s.AffectGroup(c, 0x13)
		s.inClouds(c)
		if c.Status == StatusDying {
			c.Combat.Dying++
			if c.Combat.Dying > 9 {
				c.Status = StatusDead
			}
		}
	}
	var o rules.Outcome
	if s.Bandage(false) {
		o.Effects = append(o.Effects, rules.Narrated{Text: "Your Teammate is Dying"})
	}
	return o
}

// Verdict = por sidesCheck (CountSides, 00BA:23DB): the fight is over when a
// side has nobody standing. A party side down is PartyLost whether its
// members fell or fled; the engine tells them apart afterwards (AfterCombat).
func (r *Rules) Verdict(rules.Fight) rules.Verdict {
	r.S.CountSides()
	switch {
	case r.S.Rv.Side[0] == 0:
		return rules.PartyLost
	case r.S.Rv.Side[1] == 0:
		return rules.PartyWon
	}
	return rules.FightOn
}

// setMemo writes memory slot `slot` (SpecialWriteZ1 0x1F..0x70). A core.Number
// is a byte no typed Memo names (a spell number outside SpellIDs), stored as
// the original stores it.
func setMemo(t *Tables, c *Character, slot int, v rules.Value) error {
	if slot < 0 || slot >= MemorySlots {
		return errBadValue
	}
	switch v := v.(type) {
	case rules.Memo:
		k := slices.Index(SpellIDs[:], v.Spell)
		if k < 0 {
			return errBadValue
		}
		c.Memorized[slot] = SpellID(k)
		if v.Pending {
			c.Memorized[slot] |= SpellPending
		}
	case rules.Number: // the game's byte: its number mapped, the flag kept
		s := SpellOfByte(byte(v))
		id, err := t.spellNumber(s.Spell())
		if err != nil {
			return err
		}
		c.Memorized[slot] = id | s&SpellPending
	default:
		return errBadValue
	}
	return nil
}

// Icon sizes (Character.IconSize).
const (
	iconSmall byte = 1
	iconLarge byte = 2
)

// iconSizeOf is race's icon size (ovr18:0EC4); 0: the race has none.
func iconSizeOf(r rules.Race) byte {
	switch r {
	case RaceDwarf, RaceGnome, RaceHalfling:
		return iconSmall
	case RaceElf, RaceHalfElf, RaceHalfOrc, RaceHuman:
		return iconLarge
	}
	return 0
}

// defaultIconSet is a fresh character's icon colours (defaultIconColours).
func defaultIconSet() (out [len(defaultIconColours) - 1]Colours) {
	for i := range out {
		t := defaultIconColours[i+1]
		out[i] = Colours((t+8)<<4 + t)
	}
	return out
}
