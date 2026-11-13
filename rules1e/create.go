package rules1e

import (
	"fmt"
	"gobox/rules"
	"slices"
	"strconv"
	"strings"
)

// Character creation and dropping, ported from ovr17.
//
// extra: the creation is one screen taking decisions, the character sheet
// of a sample character (View.Creation): its gender, race, alignment and
// class are drop-downs (an option that does not fit forces the dependants
// to their first valid value: SetField), the name and the ability
// scores are typed (Rename, SetStat), the portrait's head and body and
// the combat icon open pickers (OpenPicker, EditIcon), and Roll rolls,
// to minimums too (Roll), Keep keeps (EndCreation). The drop-downs, the
// typed lines and the questions are the presenter's. The original's four
// menus, name prompt, Head/Body stepping and icon editor states are gone.

// Kind is what the player chooses; the rest of the record follows from it.
type Kind struct {
	Race  rules.Race
	Sex   byte
	Class ClassID
	Align byte
}

// The fields of a kind, for enforce.
const (
	FRace = iota
	FSex
	FClass
	FAlign
)

// PickRaces are the races offered, in list order (no half-orcs, as the original).
var PickRaces = []rules.Race{RaceDwarf, RaceElf, RaceGnome, RaceHalfElf, RaceHalfling, RaceHuman}

// ClassFor: the race may be the class (DS:3E17).
func ClassFor(race rules.Race, class ClassID) bool {
	return slices.Contains(RaceCareers(race), Career(class))
}

// AlignFor: the class may be of the alignment (DS:3FE7).
func AlignFor(class ClassID, align byte) bool {
	return slices.Contains(CareerAlignments(Career(class)), AlignmentNumber(align))
}

// PickClasses are the classes any offered race may be, in class order.
func PickClasses() []ClassID {
	var out []ClassID
	for cl := ClassID(0); cl < ClassID(len(ClassNames)); cl++ {
		if slices.ContainsFunc(PickRaces, func(r rules.Race) bool { return ClassFor(r, cl) }) {
			out = append(out, cl)
		}
	}
	return out
}

// Enforce applies the rules after field f changed: a dependant that
// became invalid takes its first valid value. A class picked for a race
// that may not be it changes the race; an alignment the class may not be
// changes the class (every race has the fighter, who may be anything).
func (k Kind) Enforce(f int) Kind {
	switch f {
	case FClass:
		if !ClassFor(k.Race, k.Class) {
			k.Race = PickRaces[slices.IndexFunc(PickRaces, func(r rules.Race) bool { return ClassFor(r, k.Class) })]
		}
	case FAlign:
		if rc := RaceCareers(k.Race); !AlignFor(k.Class, k.Align) {
			k.Class = ClassID(rc[slices.IndexFunc(rc, func(cl Career) bool { return AlignFor(ClassID(cl), k.Align) })])
		}
		return k
	}
	if !ClassFor(k.Race, k.Class) {
		k.Class = ClassID(RaceCareers(k.Race)[0])
	}
	if !AlignFor(k.Class, k.Align) {
		k.Align = byte(CareerAlignments(Career(k.Class))[0])
	}
	return k
}

// Fresh builds a new character of kind k: the record the original filled
// in as its menus were answered. keep carries the name, portrait, icon and
// items over from the character it replaces (the abilities are rolled again:
// their limits are the race's and class's).
func (e *RuleState) Fresh(k Kind, keep *Character) *Character {
	c := e.KindRecord(k, keep)
	c.RandomID = byte(e.Random(0x100))
	c.Age = uint16(e.startAge(c))
	return c
}

// KindRecord is fresh without its two dice (RandomID, the age): what
// follows from the kind alone, so a kept draft rebuilds it (porCreation.Make).
func (e *RuleState) KindRecord(k Kind, keep *Character) *Character {
	c := &Character{}
	for i := 1; i <= 6; i++ {
		t := defaultIconColours[i]
		c.IconColour[i-1] = Colours((t+8)<<4 + t)
	}
	c.BaseAC = AcTen
	c.Thac0Base = 0x28
	c.Status = StatusOkay
	c.Okay = 1
	c.PortHead = 1
	c.PortBody = 1
	c.Size = 1
	c.SetName("No-One")
	if keep != nil {
		c.CharName, c.PortHead, c.PortBody, c.Portrait = keep.CharName, keep.PortHead, keep.PortBody, keep.Portrait
		c.IconHead, c.IconBody, c.IconColour, c.Icon = keep.IconHead, keep.IconBody, keep.IconColour, keep.Icon
		c.PoolIDs = keep.PoolIDs
		c.Items = keep.Items // Modify: the gear stays with the character
	}
	c.Race, c.Sex, c.Class, c.Alignment = k.Race, k.Sex, k.Class, k.Align
	switch c.Race {
	case RaceHalfling:
		c.IconSize = 1
		e.AddAffect(c, affSaveBonusVsType42, 0, 0xFF, 0) // 0x5A's paralysis bonus is in the stored save (ClassStats)
	case RaceDwarf:
		c.IconSize = 1
		for _, t := range []AffectType{affSaveBonusVsType42, affHitBonusVsTargetNames, affMinus4VsNames} {
			e.AddAffect(c, t, 0, 0xFF, 0)
		}
	case RaceGnome:
		c.IconSize = 1
		for _, t := range []AffectType{affSaveBonusVsType42, affHitBonusVsNames, affMinus4VsNames, affMinus4VsBugbearGnoll} {
			e.AddAffect(c, t, 0, 0xFF, 0)
		}
	case RaceElf:
		c.IconSize = 2
		e.AddAffect(c, affBlock90, 0, 0xFF, 0)
	case RaceHalfElf:
		c.IconSize = 2
		e.AddAffect(c, affBlock30, 0, 0xFF, 0)
	default:
		c.IconSize = 2
	}
	c.MaxLevel = 1
	class := int(int8(c.Class))
	switch {
	case class >= ClsCleric && class <= ClsMonk:
		c.Levels[class] = 1
	case class == clsClericFighter:
		c.Levels[ClsCleric], c.Levels[ClsFighter] = 1, 1
	case class == clsClericFighterMage:
		c.Levels[ClsCleric], c.Levels[ClsFighter], c.Levels[ClsMage] = 1, 1, 1
	case class == ClsClericRanger:
		c.Levels[ClsCleric], c.Levels[ClsRanger] = 1, 1
	case class == clsClericMage:
		c.Levels[ClsCleric], c.Levels[ClsMage] = 1, 1
	case class == clsClericThief:
		c.Levels[ClsCleric], c.Levels[ClsThief] = 1, 1
	case class == clsFighterMage:
		c.Levels[ClsFighter], c.Levels[ClsMage] = 1, 1
	case class == ClsFighterThief:
		c.Levels[ClsFighter], c.Levels[ClsThief] = 1, 1
	case class == clsFighterMageThief:
		c.Levels[ClsFighter], c.Levels[ClsMage], c.Levels[ClsThief] = 1, 1, 1
	case class == clsMageThief:
		c.Levels[ClsMage], c.Levels[ClsThief] = 1, 1
	}
	switch class { // Felix's R3 ruling 10, the later games' for all
	case ClsPaladin:
		c.PaladinCures = 1
		e.AddAffect(c, AffProtEvil, 0, 0xFF, 0)
	case ClsRanger, ClsClericRanger:
		e.AddAffect(c, AffRangerVsGiants, 0, 0xFF, 0)
	}
	if int8(c.Levels[ClsThief]) > 0 {
		e.thiefSkills(c) // ovr24:0329
	}
	c.ClassMask = 0
	c.Thac0Base = 0
	for k := 0; k <= 7; k++ {
		if int8(c.Levels[k]) > 0 {
			if t := byte(e.T().Thac0At(PorClass[k], int(int8(c.Levels[k])))); t > c.Thac0Base {
				c.Thac0Base = t
			}
			c.ClassMask += byte(e.T().ClassMask[k])
		}
	}
	if keep == nil {
		defaultIconParts(c)
	}
	e.savingThrows(c) // ovr24:0255
	return c
}

// startAge rolls c's starting age.
func (e *RuleState) startAge(c *Character) int {
	class := int(int8(c.Class))
	// Starting age: single classes roll the dice, multi-classes use the
	// maximum.
	age := 0
	if class <= ClsMonk {
		base, count, sides := e.T().StartingAge(c.Race, Career(class))
		age = base + int(byte(e.Dice(count, sides)))
	} else {
		o := -1
		switch class {
		case 8, 9, 0xA, 0xB, 0xC: // 0xA (cleric/ranger) as in CotAB
			o = ClsCleric
		case 0xD, 0xF, 0x10:
			o = ClsMage
		case 0xE:
			o = ClsFighter
		}
		if o >= 0 {
			base, count, sides := e.T().StartingAge(c.Race, Career(o))
			age = base + count*sides
		}
	}
	return age
}

// LimitStat keeps v within the race limits and the class minimum
// (DS:3DA3.., DS:3F81..); a cleric multi-class has wisdom 13 at least.
func LimitStat(c *Character, i int, v Score) Score {
	l, h := AbilityRange(c.Race, Ability(i))
	lo, hi := Score(l), Score(h)
	if v < lo {
		v = lo
	}
	if v > hi {
		v = hi
	}
	if m := CareerMinimum(Career(c.Class), Ability(i)); int(v) < m {
		v = Score(m)
	}
	if i == StatWis && v < 0xD && c.Class >= clsClericFighter && c.Class <= clsClericThief { // set [8..12] at cs:0525
		v = 0xD
	}
	return v
}

// AdjustStat: age, race limits and class minimums on a rolled stat.
func AdjustStat(c *Character, i int, v Score) Score {
	// Age steps DS:3F31 + race*10 (4 words) against the age word.
	older := func(j int) bool { return AgeStep(c.Race, j+1) < int(c.Age) }
	switch i {
	case 0:
		if v != 0 && older(0) {
			v++
		}
		if v != 0 && older(1) {
			v--
		}
		if v != 0 && older(2) {
			v -= 2
		}
		if v != 0 && older(3) {
			v--
		}
	case 1:
		if older(1) {
			v++
		}
		if older(3) {
			v++
		}
	case 2:
		if !older(0) {
			v--
		}
		if older(1) {
			v++
		}
		if older(2) {
			v++
		}
		if older(3) {
			v++
		}
	case 3:
		if older(2) {
			v -= 2
		}
		if older(3) {
			v--
		}
	case 4:
		if older(1) {
			v--
		}
		if older(2) {
			v--
		}
		if older(3) {
			v--
		}
	}
	return LimitStat(c, i, v)
}

// bestStats (extra: Minimal stats): the most each stat can reach, from a
// roll of 18, in the units of minStats (STR as strScore).
func bestStats(c *Character) [6]int {
	var best [6]int
	for i := range best {
		v := AdjustStat(c, i, 18)
		best[i] = int(v)
		if i == 0 {
			best[0] = StrScore(v, 0)
			if v == 0x12 && FighterKind(c) {
				best[0] = StrScore(v, min(100, byte(ExceptionalStrengthCap(c.Race))))
			}
		}
	}
	return best
}

// rollStats rolls the abilities: the best of six 3d6+1 each (Felix's R3
// ruling 1, the cotab family's die), then the age,
// race and class limits. (The original's race adjustments on the dice
// never applied: their switch read the race menu's loop variable, left at
// 7, so they are not here.)
func (e *RuleState) rollStats(c *Character) {
	e.SetBaseStats(c, func(int) (st byte) { // each ability rolled as its turn comes: STR's percentile falls before DEX's dice
		for n := 1; n <= 6; n++ {
			if roll := byte(e.Dice(3, 6) + 1); st < roll {
				st = roll
			}
		}
		return st
	})
}

// SetBaseStats gives c the base abilities base(0..5) (rolled, or bought:
// extra), in order, with the race and class rules applied, as after a roll.
func (e *RuleState) SetBaseStats(c *Character, base func(i int) byte) {
	c.Str00 = 0
	for i := 0; i <= 5; i++ {
		st := &c.Stats[i]
		*st = AdjustStat(c, i, Score(base(i)))
		if i == 0 && *st == 0x12 && FighterKind(c) {
			c.Str00 = byte(e.Random(100) + 1)
			if m := byte(e.T().ExceptionalStrengthCap(c.Race)); m < c.Str00 {
				c.Str00 = m
			}
		}
	}
}

// RollHP rolls the starting money, then the hit points: per class the
// better of two rolls of its first-level dice, summed and divided by the
// class count, with the constitution adjustment per die (Felix's R3 ruling
// 2, the sotsb/pod form; POR's two-thirds minimum and cotab's discarded die
// go); then the start's experience, trained silently (ruling 7).
func (e *RuleState) RollHP(c *Character) {
	c.HP = c.MaxHP
	count := e.FirstLevel(c)
	e.startMoney(c, count)
	rolled := 0
	for k := 0; k <= 7; k++ {
		if int8(c.Levels[k]) > 0 {
			n, sides := e.T().TrainingDie(PorClass[k])
			a := e.Dice(n, sides)
			rolled += max(a, e.Dice(n, sides))
		}
	}
	e.setHitRoll(c, rolled, count)
	e.startExperience(c, count, true)
}

// setHitRoll gives c its hit points from the hit-dice roll before
// adjustments: roll and the constitution's per class, at least 1.
func (e *RuleState) setHitRoll(c *Character, roll int, count byte) {
	c.HPRoll = byte(max(1, roll/int(count)))
	c.MaxHP = byte(max(1, (roll+e.conHPBonus(c))/int(count)))
	c.HP = c.MaxHP
}

// FirstLevel is rollHP without its dice: the base attack, the first spell
// slots and books (wisdom's bonus included); the count of classes.
func (e *RuleState) FirstLevel(c *Character) (count byte) {
	c.BaseAttack[0], c.BaseDice[0], c.BaseSides[0], c.StrBonus, c.BaseMove = 2, 1, 2, 1, 0xC
	for k := 0; k <= 7; k++ {
		if int8(c.Levels[k]) <= 0 {
			continue
		}
		if k == 0 {
			c.SpellSlots[SpellClassCleric][0] = startClericSlots
		} else if k == 5 {
			c.SpellSlots[SpellClassMage][0] = 1
		}
		count++
		if k == 0 {
			e.wisdomSlots(c) // ovr24:01AF
			for s := 1; s <= LastSpell; s++ {
				if rec := e.T().Spells[s][:2]; SpellClass(rec[0]) == SpellClassCleric && rec[1] == 1 {
					c.SpellBook[s-1] = 1
				}
			}
		}
	}
	e.startSpells(c)
	return count
}

// startClericSlots are a new cleric's first-level slots before wisdom's
// bonus (Felix's R3 ruling 9: sotsb/pod's 2 for all).
const startClericSlots = 2

// RollCharacter is one roll: abilities, then hit points.
func (e *RuleState) RollCharacter(c *Character) {
	e.rollStats(c)
	e.RollHP(c)
}

// SetMinStats takes the minimums typed (as minimumsText; a field that is
// not a value keeps the last one).
func (e *RuleState) SetMinStats(text string) {
	for i, f := range strings.Fields(text) {
		if v, ok := parseMinStat(f, i == 0); ok && i < 6 {
			e.MinStats[i] = v
		}
	}
}

// RollMinimal (extra: Minimal stats) rolls on until the minimums are met;
// impossible minimums are refused (*core.MinimumError).
func (e *RuleState) RollMinimal(c *Character) error {
	best := bestStats(c)
	for i := range statNames {
		if e.MinStats[i] > best[i] {
			return MinimumError(i, e.MinStats[i], best[i], 0)
		}
	}
	const most = 200000 // ponytail: a cap in place of "any key stops"; minimums this rare are given up
	for tries := 1; ; tries++ {
		e.RollCharacter(c)
		if e.MeetsMinStats(c) {
			return nil
		}
		if tries == most {
			return MinimumError(0, 0, 0, most)
		}
	}
}

// MinimumError is rollMinimal's refusal: minimum want of stat i above best
// (strength as strScore), or not met in tries rolls.
func MinimumError(i, want, best, tries int) *rules.MinimumError {
	m := &rules.MinimumError{Ability: Abilities[i], Tries: tries}
	m.Want, m.WantPct = unStrScore(i, want)
	m.Best, m.BestPct = unStrScore(i, best)
	return m
}

// unStrScore is strScore's inverse for stat i (strength), else v as it is.
func unStrScore(i, v int) (st, pct int32) {
	if i != 0 {
		return int32(v), 0
	}
	if v > 1800 && v <= 1900 {
		return 18, int32(v - 1800)
	}
	return int32(v / 100), 0
}

var statNames = [6]string{"STR", "INT", "WIS", "DEX", "CON", "CHA"}

// SetStat gives ability i the score st (STR: x its 18/xx, 100 = 18/00),
// kept within the race and class limits, with wisdom's spell slots and
// constitution's hit points along.
func (e *RuleState) SetStat(c *Character, i int, st Score, x byte) {
	c.Stats[i] = LimitStat(c, i, st)
	if i == 0 {
		if c.Stats[0] != 18 || !FighterKind(c) {
			x = 0
		}
		c.Str00 = min(x, byte(e.T().ExceptionalStrengthCap(c.Race)))
	}
	switch i {
	case StatWis:
		if c.SpellSlots[SpellClassCleric][0] > 0 {
			c.SpellSlots[SpellClassCleric][0] = startClericSlots
		}
		e.wisdomSlots(c)
	case StatCon:
		c.MaxHP = min(max(c.MaxHP, e.MinHitPoints(c)), e.MaxHitPoints(c))
		c.HP = c.MaxHP
	}
}

// SheetChar is the character on the creation sheet: c, of kind k.
type SheetChar struct {
	C *Character
	K Kind
}

// Options are field f's option names, the one c holds and which fit.
func (cr *SheetChar) Options(f int) ([]string, int, func(int) bool) {
	c := cr.C
	switch f {
	case FieldSex:
		return GenderNames[:], int(c.Sex), func(int) bool { return true }
	case FieldRace:
		names := make([]string, len(PickRaces))
		for i, r := range PickRaces {
			names[i] = RaceName(r)
		}
		return names, slices.Index(PickRaces, c.Race), func(i int) bool { return ClassFor(PickRaces[i], c.Class) }
	case FieldAlign:
		return AlignmentNames[:], int(c.Alignment), func(i int) bool { return AlignFor(c.Class, byte(i)) }
	}
	classes := PickClasses()
	names := make([]string, len(classes))
	for i, cl := range classes {
		names[i] = ClassNames[cl]
	}
	return names, slices.Index(classes, c.Class), func(i int) bool { return ClassFor(c.Race, classes[i]) && AlignFor(classes[i], c.Alignment) }
}

// SetSheetField gives field f its option x on the creation sheet.
func (e *RuleState) SetSheetField(cr *SheetChar, f, x int) {
	c := cr.C
	// rekind replaces c by a fresh character of kind k, rolled.
	rekind := func() {
		c = e.Fresh(cr.K, c)
		e.RollCharacter(c)
		cr.C = c // the caller swaps the selection (creationReplaced)
	}
	if f > FieldClass {
		e.SetPower(c, f-FieldClass-1, x)
		return
	}
	if f < 0 {
		return
	}
	names, cur, _ := cr.Options(f)
	if x < 0 || x >= len(names) || x == cur {
		return
	}
	switch f {
	case FieldSex:
		c.Sex = byte(x)
	case FieldRace:
		cr.K.Race = PickRaces[x]
		cr.K = cr.K.Enforce(FRace)
		rekind()
	case FieldClass:
		cr.K.Class = PickClasses()[x]
		cr.K = cr.K.Enforce(FClass)
		rekind()
	case FieldAlign:
		cr.K.Align = byte(x)
		if cr.K = cr.K.Enforce(FAlign); cr.K.Class != c.Class {
			rekind()
		} else {
			c.Alignment = cr.K.Align
		}
	}
}

// powerSlot is one of the creation's power picks (extra: rules4e): the
// pick, and which of its options the character holds (-1: none yet).
type powerSlot struct {
	PowerOwed
	Pick int
}

// PowerSlots are c's power picks, in the order the rule set owes them
// (Picks, already of c's classes only): those c made, then those owed.
func (e *RuleState) PowerSlots(c *Character) []powerSlot {
	held := c.Powers
	defer func() { c.Powers = held }()
	c.Powers = nil
	var out []powerSlot
	for {
		ps, took := e.Rules().Picks(e, c), false
		for _, p := range ps {
			i := -1 // the first power held, in the order chosen, that this pick offers
			for _, id := range held {
				if i < 0 && !slices.Contains(c.Powers, id) {
					i = slices.IndexFunc(p.Options, func(m Maneuver) bool { return m.ID == id })
				}
			}
			if i >= 0 {
				out, took = append(out, powerSlot{p, i}), true
				c.Powers = append(c.Powers, p.Options[i].ID)
				break
			}
		}
		if !took {
			for _, p := range ps {
				out = append(out, powerSlot{p, -1})
			}
			return out
		}
	}
}

// SetPower gives c's power pick i its option x: in place of the power it
// held, else a new one; a power c holds in another pick does not fit.
func (e *RuleState) SetPower(c *Character, i, x int) {
	ss := e.PowerSlots(c)
	if i < 0 || i >= len(ss) || x < 0 || x >= len(ss[i].Options) {
		return
	}
	s, id := ss[i], ss[i].Options[x].ID
	switch {
	case slices.Contains(c.Powers, id):
	case s.Pick >= 0:
		c.Powers[slices.Index(c.Powers, s.Options[s.Pick].ID)] = id
	default:
		c.Powers = append(c.Powers, id)
	}
}

// DefaultPowers answers each of c's open power picks (extra: rules4e)
// with its first option c does not hold yet.
func (e *RuleState) DefaultPowers(c *Character) {
	for i := 0; ; i++ {
		ss := e.PowerSlots(c)
		if i >= len(ss) {
			return
		}
		if ss[i].Pick >= 0 {
			continue
		}
		for x, m := range ss[i].Options {
			if !slices.Contains(c.Powers, m.ID) {
				e.SetPower(c, i, x)
				break
			}
		}
	}
}

// StrScore folds 18/xx into one number: 18/01 = 1801, 18/00 = 1900.
func StrScore(st Score, x byte) int {
	if st == 18 && x > 0 {
		return 1800 + int(x)
	}
	return int(st) * 100
}

func (e *RuleState) MeetsMinStats(c *Character) bool {
	for i, m := range e.MinStats {
		v := int(c.Stats[i])
		if i == 0 {
			v = StrScore(c.Stats[0], c.Str00)
		}
		if v < m {
			return false
		}
	}
	return true
}

func fmtStr(v int) string {
	if v <= 1800 || v > 1900 {
		return strconv.Itoa(v / 100)
	}
	return fmt.Sprintf("18/%02d", (v-1800)%100)
}

func parseMinStat(s string, str bool) (int, bool) {
	a, b, slash := strings.Cut(s, "/")
	n, err := strconv.Atoi(a)
	if err != nil || n < 0 || n > 25 {
		return 0, false
	}
	if !slash {
		if str {
			return n * 100, true
		}
		return n, true
	}
	x, err := strconv.Atoi(b)
	if !str || n != 18 || err != nil || x < 0 || x > 99 {
		return 0, false
	}
	if x == 0 {
		x = 100
	}
	return StrScore(18, byte(x)), true
}
