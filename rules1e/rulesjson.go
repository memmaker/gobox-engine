package rules1e

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"

	"gobox/rules"
)

// rulesJSON is the part of an imported game's data/rules.json (tablespec,
// the cotab family's names) that maps onto Tables. Byte tables decode as
// numbers (tablespec writes arrays, never base64). Where the games name a
// table differently both names are listed; a game carries one of them.
type rulesJSON struct {
	Thac0    [][]int16 `json:"thac0Table"`       // by class, by level
	ConHP    []int8    `json:"conHitPointBonus"` // by constitution
	Saves    []int16   `json:"saveTableData"`    // flat: class*SaveRows*5 + level*5 + save
	OldSaves []int16   `json:"oldClassSaveData"` // flat: level*5 + save
	Paladin  []int16   `json:"paladinSlotData"`  // flat: level*stride + spell level (1..)
	Ranger   []int16   `json:"rangerSlotData"`
	Cleric   []int16   `json:"clericSlotData"`
	Mage     []int16   `json:"mageSlotData"`

	XPRaw   [][]int16 `json:"expTableData"` // COTAB: 99-byte rows of little-endian longints
	XP      [][]int32 `json:"expTable"`     // SotSB, POD: by class, by level
	XPSteps []int32   `json:"expSteps"`     // POD: what each level past the table adds

	HitDiceFirst  []int16   `json:"classHitDiceFirst"` // COTAB; SotSB/POD: classFirstHitDice
	FirstHitDice  []int16   `json:"classFirstHitDice"`
	HitDiceSides  []int16   `json:"classHitDiceSides"`
	ThiefBase     []int16   `json:"thiefBaseData"`
	ThiefRaceDex  []int8    `json:"thiefRaceDexData"` // COTAB: race rows, the dexterity rows at thiefDexOffset
	ThiefRace     []int8    `json:"thiefRaceData"`    // SotSB, POD
	ThiefDex      []int8    `json:"thiefDexData"`
	RaceClasses   [][]int16 `json:"raceClasses"`
	RaceAges      [][]int16 `json:"raceAges"` // per race: 7 x (base word, count, sides)
	AgeSteps      [][]int32 `json:"ageSteps"`
	StatLimits    [][]int16 `json:"raceStatLimits"`    // COTAB
	AbilityLimits [][]int16 `json:"raceAbilityLimits"` // SotSB, POD (SotSB carries both: this one is read)
	ClassMin      [][]int16 `json:"classStatMinimum"`
	ClassAlign    [][]int16 `json:"classAlignments"`

	Ground       [][]int16                  `json:"groundInfo"`       // by ground number: move cost, height, sight, tile
	WildFlags    []int16                    `json:"wildTerrainFlags"` // the terrain flags by terrain number (CotAB; the others carry their own in combatGround)
	CombatGround *GroundGen                 `json:"combatGround"`     // the importer's: how the game lays its combat ground
	Kinds        ItemKinds                  `json:"itemKinds"`        // the importer's: the dnd id of each base item type
	Morale       *bool                      `json:"moraleUncapped"`   // the importer's: a morale roll above 102 stands (PoD)
	GroundIDs    *GroundIDs                 `json:"groundIds"`        // the importer's: the grounds the rules lay
	ImportFlag   *bool                      `json:"importFlag"`       // the importer's: the game flags a character added from another game
	Town         *TownData                  `json:"town"`             // the importer's: the hall fee and the temple services
	Start        *StartData                 `json:"creationStart"`    // the importer's, from the game's code (StartData)
	Library      []*Item                    `json:"itemLibrary"`      // the importer's: the data set's item library (POD ITEM0)
	Scroll       *RandomScroll              `json:"randomScroll"`     // the importer's, from the game's code
	Treasure     *TreasureRolls             `json:"randomTreasure"`   // the importer's, from the game's code
	Appraisal    *Appraisal                 `json:"appraisal"`        // the importer's, from the game's code
	Affects      AffectMap                  `json:"affectIds"`        // the importer's: each affect number of the game as its id (AffectMap)
	Spells       [][]SpellProp              `json:"spellTableData"`   // the spell table, by the game's spell number
	CasterLevel  *CasterLevelRule           `json:"casterLevel"`      // the importer's: the game's caster-level rule
	OldAge       *OldAgeRule                `json:"oldAge"`           // the importer's: death of old age after a fight
	Hammer       *Item                      `json:"spiritualHammer"`  // the importer's: the game's Spiritual Hammer
	Resist       *MagicResistRule           `json:"magicResistance"`  // the importer's: the game's own magic resistance
	Giants       *GiantRule                 `json:"giantRule"`        // the importer's: how the game tells a giant
	AffectRule   *AffectRule                `json:"affectRule"`       // the importer's: the game's affect renewal and down affects
	ItemSpell    *ItemSpellCode             `json:"itemSpellCode"`    // the importer's: how an item's +0x3D holds its spell
	SpellIDs     SpellMap                   `json:"spellIds"`         // the importer's: each spell number of the game as its id (SpellMap)
	SpellsGone   SpellMap                   `json:"spellsGone"`       // the importer's: removed spells' numbers as the id they read as (Tables.SpellsGone)
	Classes      SpellClassMap              `json:"spellClasses"`     // the importer's: each spell class byte of the game as its SpellClass
	Powers       PowerMap                   `json:"itemPowers"`       // the importer's: each item power code of the game as a unified power (PowerMap)
	Mends        []ItemMend                 `json:"itemMends"`        // the importer's: the game's data fixes (none: no game runs another's)
	Groups       [][]AffectType             `json:"affectGroups"`     // the importer's: the game's affect groups by meaning (grp*), over the ids
	Variants     map[AffectType]HandlerName `json:"affectVariants"`   // the importer's: the game's handler where it plays an id differently from POR (or the id is the game's own)
}

// Layout is what a game's rules.json does not say about its tables' shape.
type Layout struct {
	SlotStride       int          // spell levels per slot row (COTAB 5, SotSB 7, POD 9)
	SlotTotals       bool         // the slot rows hold totals at a level (POD), not gains
	RangerMageColumn int          // the ranger row's column of magic-user level 1 (POD 5); 0: 4, after druid 1..3
	SlotLevels       int          // the last class level the slot tables hold before other data (COTAB 12); 0: the tables' end
	Races            []rules.Race // the race of each race row; nil = POR's order; "" = a row of no race (SotSB's placeholder 0)
}

// Table geometry of the cotab family's flat tables.
const (
	saveCount      = 5  // saves per row
	thiefSkills    = 8  // thiefBaseData / thiefRaceData row: a leading byte, then 8 skills
	thiefDexSkills = 5  // thiefDexData row: a leading byte, then 5 skills
	thiefDexOffset = 19 // COTAB thiefRaceDexData: where the dexterity rows start
	ageClasses     = 7  // raceAges: entries per race
	ageEntry       = 4  // raceAges entry: base word, count, sides
)

// A raceStatLimits / raceAbilityLimits row: strength min, max and
// exceptional max by gender (male first), then min and max of int wis dex
// con cha. All games keep the male values (Felix 20:05; Tables.RaceLimits).
const (
	limStrMin, limStrMax, limStr00, limRest = 0, 2, 4, 6
)

// WithRules is base with the tables rules.json (src) carries; a table it
// does not carry stays base's. The class tables are 8 classes in classRow's
// order (SotSB and POD: 7, no monk); race tables follow l.Races.
func (base Tables) WithRules(src io.Reader, l Layout) (Tables, error) {
	var r rulesJSON
	if err := json.NewDecoder(src).Decode(&r); err != nil {
		return Tables{}, err
	}
	classes := len(r.Thac0)
	if classes != len(classRow) && classes != len(classRow)-1 || l.SlotStride < saveCount {
		return Tables{}, fmt.Errorf("rules1e: rules.json has %d thac0 rows (want classRow's order, 8 or 7 without the monk), slot stride %d", classes, l.SlotStride)
	}
	t := base
	t.Thac0 = r.Thac0
	t.ItemMends = r.Mends
	t.Races = l.Races
	set(&t.ConHPByScore, r.ConHP)
	if r.Saves != nil {
		t.Saves, t.SaveRows, t.PaladinSaves = rows(r.Saves, 0, saveCount, saveCount), len(r.Saves)/saveCount/classes, nil
	}
	if r.OldSaves != nil {
		t.MonkOldSaves = rows(r.OldSaves, 0, saveCount, saveCount)
	}
	slotRows := func(flat []int16, totals bool, width int) [][]int16 {
		g := rows(flat, 1, l.SlotStride, width)
		if l.SlotLevels > 0 && len(g) > l.SlotLevels+1 {
			g = g[:l.SlotLevels+1]
		}
		switch {
		case totals && !l.SlotTotals:
			return toTotals(g)
		case !totals && l.SlotTotals:
			return toGains(g)
		case !totals:
			for lvl := range min(2, len(g)) { // COTAB, SotSB: rows 0 and 1 are the table's header bytes, no gain
				g[lvl] = make([]int16, len(g[lvl]))
			}
		}
		return g
	}
	if r.Paladin != nil {
		t.PaladinSpells = slotRows(r.Paladin, false, saveCount)
	}
	if r.Ranger != nil {
		t.RangerSpells = slotRows(r.Ranger, false, saveCount)
		if c := l.RangerMageColumn; c > 0 { // POD: magic-user levels 1..2 at columns c, c+1
			wide := slotRows(r.Ranger, false, l.SlotStride)
			for lvl, row := range wide {
				t.RangerSpells[lvl] = []int16{row[0], row[1], row[2], row[c-1], row[c]}
			}
		}
	}
	if r.Cleric != nil {
		t.ClericSpells = slotRows(r.Cleric, true, l.SlotStride)
	}
	if r.Mage != nil {
		t.MageSpells = slotRows(r.Mage, true, l.SlotStride)
	}
	if r.Cleric != nil || r.Mage != nil { // the imported rows end at the table's end
		t.SlotLevels, t.SlotsClamp = min(len(t.ClericSpells), len(t.MageSpells))-1, true
	}
	if r.XPRaw != nil {
		t.XP = xpRows(xpFromBytes(r.XPRaw))
	}
	if r.XP != nil {
		t.XP = xpRows(r.XP)
	}
	set(&t.XPSteps, r.XPSteps)
	set(&t.HitDieCount, r.HitDiceFirst)
	set(&t.HitDieCount, r.FirstHitDice)
	set(&t.HitDieSides, r.HitDiceSides)
	if r.ThiefBase != nil {
		t.ThiefByLevel = rows(r.ThiefBase, 1, thiefSkills, thiefSkills)
	}
	race, dex := r.ThiefRace, r.ThiefDex
	if r.ThiefRaceDex != nil {
		race, dex = r.ThiefRaceDex, r.ThiefRaceDex[thiefDexOffset:]
	}
	if race != nil {
		t.ThiefByRace = byRace(t.ThiefByRace, rows(race, 1, thiefSkills, thiefSkills), l.Races)
	}
	if dex != nil {
		t.ThiefByDex = rows(dex, 1, thiefDexSkills, thiefDexSkills)
	}
	if r.RaceClasses != nil {
		t.RaceClasses = byRace(t.RaceClasses, r.RaceClasses, l.Races)
	}
	if r.RaceAges != nil {
		t.StartAge = byRace(t.StartAge, ages(r.RaceAges), l.Races)
	}
	if r.AgeSteps != nil {
		t.AgeSteps = byRace(t.AgeSteps, r.AgeSteps, l.Races)
	}
	limits := r.StatLimits
	set(&limits, r.AbilityLimits)
	if limits != nil {
		t.RaceLimits = byRace(t.RaceLimits, raceLimits(limits), l.Races)
	}
	set(&t.ClassMinStats, r.ClassMin)
	set(&t.ClassAlignments, r.ClassAlign)
	if r.Ground != nil {
		t.Ground = make([][]GroundProp, len(r.Ground))
		for i, g := range r.Ground {
			for _, v := range g {
				t.Ground[i] = append(t.Ground[i], GroundProp(v))
			}
		}
	}
	if r.WildFlags != nil {
		t.Terrain = make([]TerrainFlag, len(r.WildFlags))
		for i, v := range r.WildFlags {
			t.Terrain[i] = TerrainFlag(v)
		}
	}
	if r.CombatGround != nil {
		t.Gen = *r.CombatGround
	}
	if r.Kinds != nil {
		t.Kinds = r.Kinds
	}
	if r.Morale != nil {
		t.MoraleUncapped = *r.Morale
	}
	if r.GroundIDs != nil {
		t.Grounds = *r.GroundIDs
	}
	if r.ImportFlag != nil {
		t.ImportFlag = *r.ImportFlag
	}
	if r.Town != nil {
		t.Town = *r.Town
	}
	if r.Start != nil {
		t.Start = *r.Start
	}
	if r.Affects != nil {
		t.AffectIDs = r.Affects
	}
	if r.SpellIDs != nil {
		t.SpellIDs = r.SpellIDs
	}
	if r.SpellIDs != nil { // a game of its own numbers: its own gone list (POR's numbering keeps PorSpellsGone)
		t.SpellsGone = r.SpellsGone
	}
	if r.CasterLevel != nil {
		t.CasterLevel = *r.CasterLevel
	}
	if r.OldAge != nil {
		t.OldAge = r.OldAge
	}
	t.Hammer, t.Resist = r.Hammer, r.Resist
	if r.Giants != nil {
		t.Giants = *r.Giants
	}
	if r.AffectRule != nil {
		t.AffectRule = *r.AffectRule
	}
	if r.ItemSpell != nil {
		t.ItemSpell = r.ItemSpell
	}
	if r.Spells != nil {
		if t.SpellIDs == nil {
			return Tables{}, fmt.Errorf("rules1e: rules.json has a spell table without spellIds")
		}
		if r.Classes == nil {
			return Tables{}, fmt.Errorf("rules1e: rules.json has a spell table without spellClasses")
		}
		named, err := classesOf(r.Spells, r.Classes)
		if err != nil {
			return Tables{}, err
		}
		sp, err := spellsByID(named, t.SpellIDs, t.SpellsGone)
		if err != nil {
			return Tables{}, err
		}
		t.Spells = sp
	}
	if r.Powers != nil {
		t.PowerIDs = r.Powers
	}
	if r.Variants != nil {
		a, err := withVariants(t.Affects, r.Variants)
		if err != nil {
			return Tables{}, err
		}
		t.Affects = a
	}
	if r.Groups != nil {
		t.Groups = r.Groups
	}
	set(&t.Library, r.Library)
	t.Treasure, t.Appraisal = r.Treasure, r.Appraisal
	if s := r.Scroll; s != nil {
		if s.Last < 1 || len(s.Mage) == 0 || slices.ContainsFunc(s.Mage, func(n SpellID) bool { return n < 1 || int(n) > s.Last }) {
			return Tables{}, fmt.Errorf("rules1e: randomScroll needs mage spells within 1..%d, has %v", s.Last, s.Mage)
		}
		t.Scroll = s
	}
	return t, nil
}

// LoadRules is base with an imported game's data/rules.json (path) laid
// out as l, for RuleState.Tables: the run-time option the cotab family
// takes once it plays through rules1e (fold, after E2).
func (base Tables) LoadRules(path string, l Layout) (*Tables, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	t, err := base.WithRules(f, l)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &t, nil
}

// set is *dst = src when rules.json carries src.
func set[T any](dst *[]T, src []T) {
	if src != nil {
		*dst = src
	}
}

// rows cuts flat into rows of width values, row n at flat[skip+n*stride:];
// a short tail is dropped.
func rows[T any](flat []T, skip, stride, width int) [][]T {
	var out [][]T
	for i := skip; i+width <= len(flat); i += stride {
		out = append(out, flat[i:i+width])
	}
	return out
}

// toTotals turns per-level gains (COTAB, SotSB: level 1 is the creation's
// one first-level spell, levels 2.. add their row) into POR's totals by level.
func toTotals(gain [][]int16) [][]int16 {
	if len(gain) < 2 {
		return gain
	}
	out := make([][]int16, len(gain))
	sum := make([]int16, len(gain[1]))
	sum[0] = 1
	for lvl := range gain {
		if lvl == 0 {
			out[lvl] = make([]int16, len(sum))
			continue
		}
		if lvl >= 2 {
			for i := range sum {
				sum[i] += gain[lvl][i]
			}
		}
		out[lvl] = append([]int16(nil), sum...)
	}
	return out
}

// toGains turns totals by level (POD) into per-level gains (rowSum adds them up).
// Row 0 is no level (POD's table header bytes): it gains nothing.
func toGains(total [][]int16) [][]int16 {
	out := make([][]int16, len(total))
	for lvl, row := range total {
		out[lvl] = make([]int16, len(row))
		for i := range row {
			switch lvl {
			case 0:
			case 1:
				out[lvl][i] = row[i]
			default:
				out[lvl][i] = row[i] - total[lvl-1][i]
			}
		}
	}
	return out
}

// xpFromBytes decodes COTAB's 99-byte rows: entry n is the longint at byte
// 4n of the class's own row (24 entries; xpRows cuts them to the table).
func xpFromBytes(raw [][]int16) [][]int32 {
	out := make([][]int32, len(raw))
	for k, v := range raw {
		for i := 0; i+4 <= len(v); i += 4 {
			out[k] = append(out[k], int32(uint32(uint8(v[i]))|uint32(uint8(v[i+1]))<<8|uint32(uint8(v[i+2]))<<16|uint32(uint8(v[i+3]))<<24))
		}
	}
	return out
}

// xpRows cuts each row at the end of its table: the first entry from level
// 2 that is not a threshold. A -1 (the class stops there) is kept, a 0 is
// the other data the originals' rows run into (POR, COTAB) and goes.
func xpRows(xp [][]int32) [][]int32 {
	out := make([][]int32, len(xp))
	for k, row := range xp {
		end := len(row)
		for lvl := 2; lvl < len(row); lvl++ {
			if row[lvl] <= 0 {
				end = lvl
				if row[lvl] == xpNone {
					end++
				}
				break
			}
		}
		out[k] = row[:end]
	}
	return out
}

// ages reads raceAges rows: per class, a base word, dice count and sides.
func ages(raw [][]int16) [][]ageDice {
	out := make([][]ageDice, len(raw))
	for i, row := range raw {
		for c := range ageClasses {
			e := row[c*ageEntry:]
			out[i] = append(out[i], ageDice{Base: int32(uint8(e[0])) | int32(uint8(e[1]))<<8, Count: e[2], Sides: e[3]})
		}
	}
	return out
}

// raceLimits reads the cotab family's limit rows, male values (as POR).
func raceLimits(raw [][]int16) []raceLimit {
	out := make([]raceLimit, len(raw))
	for i, row := range raw {
		out[i] = raceLimit{StrMin: row[limStrMin], StrMax: row[limStrMax], Str00Max: row[limStr00]}
		for a := range out[i].Min {
			out[i].Min[a], out[i].Max[a] = row[limRest+2*a], row[limRest+2*a+1]
		}
	}
	return out
}

// byRace puts the game's race rows at raceRow's places; nil races = the
// same order. A POR row the game has no race for stays base's.
func byRace[T any](base, rows []T, races []rules.Race) []T {
	if races == nil {
		return rows
	}
	out := append([]T(nil), base...)
	for i, race := range races {
		if r, ok := raceRow[race]; ok && i < len(rows) {
			out[r] = rows[i]
		}
	}
	return out
}

// LayoutFile is a game's Layout in its data folder, written by its import.
const LayoutFile = "layout.json"

// ReadLayout reads a Layout (LayoutFile) from path.
func ReadLayout(path string) (Layout, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Layout{}, fmt.Errorf("%w (run the game's import)", err)
	}
	var l Layout
	if err := json.Unmarshal(b, &l); err != nil {
		return Layout{}, fmt.Errorf("%s: %w", path, err)
	}
	return l, nil
}

// withVariants is the affect handler table affects with a game's
// affectVariants in place of POR's handlers.
func withVariants(affects, variants map[AffectType]HandlerName) (map[AffectType]HandlerName, error) {
	a := maps.Clone(affects)
	for id, name := range variants {
		if affectHandler[name] == nil && !unported(name) {
			return nil, fmt.Errorf("rules1e: affectVariants %#x names no handler %q", id, name)
		}
		a[id] = name
	}
	return a, nil
}

// unported tells whether name is on UnportedAffectHandlers: its id plays
// nothing until it is ported (the list only shrinks).
func unported(name HandlerName) bool {
	for _, m := range UnportedAffectHandlers {
		if _, ok := m[name]; ok {
			return true
		}
	}
	return false
}
