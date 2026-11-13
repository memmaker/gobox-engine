package rules1e

// Score limits of the CotAB stat items (O0E3:1771 statAffects).
const (
	scoreUpLimit   Score = 18 // a score-up item adds 1 only below it
	feebleScore    Score = 7  // the low-intelligence item's intelligence
	lowScore       Score = 3  // the low strength-and-mind item's scores; feeblemind's
	noFixedScore   Score = 0
	dexDownBy      Score = 2
	dexUpLow       Score = 6  // free action: +4 up to it
	dexUpMid       Score = 13 // +2 up to it, else +1
	constitutionUp Score = 1
)

// ownScore is c's own score, before items: the record's base, or (POR, a
// character made here) the score in use.
func ownScore(c *Character, stat int) Score {
	if c.Base[stat] != 0 {
		return c.Base[stat]
	}
	return c.Stats[stat]
}

// statAffects = CotAB O0E3:1771 (char, stat): the score in use from the
// character's own and its readied stat items. Strength from spells keeps
// POR's model (SetStrength, bestStrength): with a strength affect on, only a
// fixed low strength applies. Constitution works the hit points and
// regeneration out again (conHitPoints).
func (e *RuleState) statAffects(c *Character, stat int) {
	if c.Base[stat] == 0 {
		c.Base[stat] = c.Stats[stat]
		if stat == StatStr {
			c.Str00Base = c.Str00
		}
	}
	own := ownScore(c, stat)
	value, fixed := own, noFixedScore
	for _, it := range c.Items {
		if it.Readied == 0 {
			continue
		}
		arg := it.Spells[1]
		switch AffectType(it.Spells[2]) {
		case itemPowerScoreUp:
			if int(arg) == stat && own < scoreUpLimit {
				value++
			}
		case itemPowerConstitution:
			switch stat {
			case StatCon:
				value += constitutionUp
			case StatCha:
				value -= constitutionUp
			}
		case itemPowerFreeActionDex:
			if stat == StatDex {
				switch {
				case own <= dexUpLow:
					value += 4
				case own <= dexUpMid:
					value += 2
				default:
					value++
				}
			}
		case itemPowerDexterityDown:
			if stat == StatDex {
				value -= dexDownBy
			}
		case itemPowerLowIntelligence:
			if stat == StatInt {
				fixed = feebleScore
			}
		case itemPowerLowStrengthMind:
			if stat == StatStr || stat == StatInt {
				fixed = lowScore
			}
		}
	}
	if (stat == StatInt || stat == StatWis) && fixed > feebleScore && e.FindAffect(c, affFeebleminded) != nil {
		fixed = lowScore
	}
	if fixed != noFixedScore {
		value = fixed
	}
	if stat == StatStr {
		if fixed == noFixedScore && e.strengthAffected(c) {
			return // a strength spell holds the score (POR's model)
		}
		c.Str00 = c.Str00Base
		if fixed != noFixedScore {
			c.Str00 = 0
		}
	}
	c.Stats[stat] = value
	if stat == StatCon {
		e.conHitPoints(c)
	}
}

// strengthAffected: a strength spell or item sets c's strength now.
func (e *RuleState) strengthAffected(c *Character) bool {
	for _, a := range c.Affects {
		if a.Type == affEnlarged || a.Type == affStrength {
			return true
		}
	}
	return false
}

// The CotAB stat items (itemReadyEffect): readying or putting one away
// works the scores it touches out again.
func (e *RuleState) constitutionItemCotab(c *Character, a *Affect, flag byte) {
	e.statAffects(c, StatCon)
	e.statAffects(c, StatCha)
}

func (e *RuleState) scoreUpItemCotab(c *Character, a *Affect, flag byte) {
	if s := int(affectItem(a).Spells[1]); s <= StatCha { // the original's case covers 0..5 only
		e.statAffects(c, s)
	}
}

func (e *RuleState) freeActionDexCotab(c *Character, a *Affect, flag byte) {
	e.statAffects(c, StatDex)
	e.thiefSkills(c)
}

func (e *RuleState) dexterityDownItemCotab(c *Character, a *Affect, flag byte) {
	e.statAffects(c, StatDex)
}

func (e *RuleState) lowIntelligenceItemCotab(c *Character, a *Affect, flag byte) {
	e.statAffects(c, StatInt)
}

func (e *RuleState) lowStrengthMindItemCotab(c *Character, a *Affect, flag byte) {
	e.statAffects(c, StatStr)
	e.statAffects(c, StatInt)
}
