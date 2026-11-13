package rules1e

import "gobox/core/goldbox"

// OldLevelRule: how a dual-classed human's old class levels count toward
// the level a spell is cast at, while the earlier class is active.
type OldLevelRule byte

const (
	OldLevelsNone   OldLevelRule = iota // POR: old levels never count
	OldLevelsAdd                        // CotAB, SotSB: added to the current level
	OldLevelsHigher                     // PoD (EffectiveClassLevel): the higher of current and old
)

// CasterLevelRule is a game's rule for the level a spell is cast at
// (casterLevel; rules.json casterLevel). The zero value is POR's.
type CasterLevelRule struct {
	NoCasterFallback    bool         `json:"noCasterFallback"`    // one of no casting class casts at itemCasterLevel (CotAB, SotSB, PoD)
	PaladinFirstCasting int8         `json:"paladinFirstCasting"` // the paladin level that makes a caster for the fallback check; 0: paladins do not count (SotSB, PoD)
	RangerFirstCasting  int8         `json:"rangerFirstCasting"`  // the same for rangers; 0: rangers do not count
	Hybrids             bool         `json:"hybrids"`             // paladins cast cleric spells (level - paladinClericOffset, Felix's ruling for all games), rangers druid and mage spells
	OldLevels           OldLevelRule `json:"oldLevels"`
	DualScanClasses     int          `json:"dualScanClasses"` // the classes the earlier-class test scans for the current one (CotAB 7, SotSB/PoD 6)
	Cap                 int          `json:"cap"`             // cappedCasterLevel's most; 0: none (PoD 31)
}

// The later games' caster-level rules (each importer writes its own).
var (
	CotabCasterLevel = CasterLevelRule{NoCasterFallback: true, PaladinFirstCasting: paladinFirstCasting, RangerFirstCasting: rangerFirstCasting,
		Hybrids: true, OldLevels: OldLevelsAdd, DualScanClasses: ClsMonk}
	SotsbCasterLevel = CasterLevelRule{NoCasterFallback: true, RangerFirstCasting: rangerFirstCasting,
		Hybrids: true, OldLevels: OldLevelsAdd, DualScanClasses: ClsThief}
	PodCasterLevel = CasterLevelRule{NoCasterFallback: true, RangerFirstCasting: rangerFirstCasting,
		Hybrids: true, OldLevels: OldLevelsHigher, DualScanClasses: ClsThief, Cap: maxSpellLevelPod}
)

const (
	paladinClericOffset = 8
	rangerDruidOffset   = 7
	rangerMageOffset    = 8
	innateCasterLevel   = 12
	itemCasterLevel     = 6
	paladinFirstCasting = 9
	rangerFirstCasting  = 8
)

// earlierClassActive: a dual-classed human's current class (the first of
// the first scan classes it has levels in) has passed the level it changed
// at, so the old class counts again.
func earlierClassActive(c *Character, scan int) bool {
	var first int8
	if c.Race == RaceHuman {
		k := 0
		for k < scan && c.Levels[k] == 0 {
			k++
		}
		first = int8(c.Levels[k])
	}
	return int8(c.MulticlassLevel) < first
}

// casterLevel = POR 00BA:26C2, CotAB O0ED:286F, SotSB/PoD GetCasterLevel
// (spell): the level the selected character casts the spell at, by the
// game's CasterLevelRule. Items cast cleric, druid and mage spells at 6,
// powers at 12.
func (e *RuleState) casterLevel(spell int) int {
	r := e.T().CasterLevel
	c := e.Acting()
	level := func(k int) int8 {
		l := int8(c.Levels[k])
		if r.OldLevels == OldLevelsNone || !earlierClassActive(c, r.DualScanClasses) {
			return l
		}
		if r.OldLevels == OldLevelsAdd {
			return l + int8(c.OldLevels[k])
		}
		return max(l, int8(c.OldLevels[k]))
	}
	check := func(k int) int8 { // PoD's test reads the effective levels, the others the current
		if r.OldLevels == OldLevelsHigher {
			return level(k)
		}
		return int8(c.Levels[k])
	}
	class := e.SpellClass(SpellID(spell))
	if e.UsingItem && class != SpellClassPower {
		return itemCasterLevel
	}
	if r.NoCasterFallback && check(ClsCleric) == 0 && check(ClsMage) == 0 &&
		(r.PaladinFirstCasting == 0 || check(ClsPaladin) < r.PaladinFirstCasting) &&
		(r.RangerFirstCasting == 0 || check(ClsRanger) < r.RangerFirstCasting) {
		return itemCasterLevel
	}
	var lvl int8
	switch class {
	case SpellClassCleric:
		lvl = level(ClsCleric)
		if r.Hybrids {
			lvl = max(lvl, level(ClsPaladin)-paladinClericOffset)
		}
	case SpellClassDruid:
		if r.Hybrids {
			lvl = max(level(ClsRanger)-rangerDruidOffset, 0)
		}
	case SpellClassMage:
		lvl = level(ClsMage)
		if r.Hybrids {
			lvl = max(lvl, level(ClsRanger)-rangerMageOffset)
		}
	case SpellClassPower:
		lvl = innateCasterLevel
	}
	return int(byte(lvl))
}

// cappedCasterLevel: casterLevel at most the game's cap (PoD's capped calls).
func (e *RuleState) cappedCasterLevel(spell int) int {
	lvl := e.casterLevel(spell)
	if c := e.T().CasterLevel.Cap; c > 0 {
		lvl = min(lvl, c)
	}
	return lvl
}

// AffectRule is a game's rule for laying and ending affects (rules.json
// affectRule). The zero value is POR's.
type AffectRule struct {
	Down           []AffectType `json:"downAffects"`    // the affects going down ends; nil: POR's combatAffects
	CharmedBerserk AffectType   `json:"charmedBerserk"` // a charmed one with this affect goes back to the party's side when down; 0: none (POR)
}

var (
	CotabAffectRule = AffectRule{Down: downAffectsCotab[:], CharmedBerserk: affBerserkCotab}
	PodAffectRule   = AffectRule{Down: downAffectsPod, CharmedBerserk: affBerserk}
)

// AddAffectMsg = POR ovr25:0F6E, the later games' isUnaffected (char, type,
// time, data, call, saveMode, saved, message): the affect unless blocked
// or negated by the save; one already there is replaced when the new one lasts longer or for good, else kept (no stacking).
func (e *RuleState) AddAffectMsg(c *Character, typ AffectType, time int, data, call, saveMode byte, saved bool, msg string) {
	msg = goldbox.Cut(msg, msgMax)
	e.Rv.AffVal = typ
	e.AffectGroup(c, grpAddAffect)
	if e.Rv.AffVal == 0 || saved && saveMode == saveNegate {
		e.CombatSay(c, "is Unaffected", true)
		return
	}
	// Felix's ruling, all games: the longer one wins (POR); no stacking.
	if a := e.FindAffect(c, typ); a == nil {
		e.AddAffect(c, typ, time, data, call)
	} else if a.Time > 0 && time > int(a.Time) || time == 0 {
		e.RemoveAffect(c, typ, a)
		e.AddAffect(c, typ, time, data, call)
	}
	if msg != "" {
		e.hitMessage(c, true, msg)
	}
}

// removeCombatAffects: going down ends the game's down affects; a charmed
// berserker goes back to the party's side.
func (e *RuleState) removeCombatAffects(c *Character) {
	r := e.T().AffectRule
	down := r.Down
	if down == nil {
		down = combatAffects[:]
	}
	for _, t := range down {
		e.RemoveAffect(c, t, nil)
	}
	if r.CharmedBerserk != 0 && e.FindAffect(c, r.CharmedBerserk) != nil && c.Control == controlCharmed {
		c.Side = 0
	}
}
