package rules1e

import "gobox/core"

// After combat, ported from ovr06: the fight's outcome, experience, the
// hired NPCs' share and the treasure screen. ECL COMBAT without monsters
// runs this directly, which is how scripts hand out treasure.

// CombatParty returns the party members before the first combat monster
// (Combat +0x13 = 1).
func (e *RuleState) CombatParty() []*Character {
	for i, c := range e.Party {
		if c.Combat != nil && c.Combat.Extra == 1 {
			return e.Party[:i]
		}
	}
	return e.Party
}

// GiveExp = ovr06:0355 (exp): every member still standing gets it, with
// a tenth more for a high prime requisite; multi-class characters split it.
func (e *RuleState) GiveExp(exp int32) {
	for _, c := range e.Party {
		if c.Okay == 0 || c.Status == StatusAnimated {
			continue
		}
		g := exp
		prime := -1
		switch c.Class {
		case ClsCleric:
			prime = StatWis
		case ClsFighter:
			prime = StatStr
		case ClsMage:
			prime = StatInt
		case ClsThief:
			prime = StatDex
		case ClsPaladin, ClsRanger: // CotAB: paladin str+wis, ranger str+int+wis over 15
			if c.Stats[StatStr] > 15 && c.Stats[StatWis] > 15 && (c.Class == ClsPaladin || c.Stats[StatInt] > 15) {
				g += g / 10
			}
		case clsClericFighter, ClsClericRanger, clsClericMage, clsClericThief, clsFighterMage, ClsFighterThief, clsMageThief:
			g = exp / 2
		case clsClericFighterMage, clsFighterMageThief:
			g = exp / 3
		}
		if prime >= 0 && c.Stats[prime] > 15 {
			g += g / 10
		}
		c.Exp = uint32(int32(c.Exp) + g)
	}
}

// GainExperience = sotsb O050:37F4's share: points split evenly among c's
// classes (a level above 0 is a class); a character with none gets nothing
// (the original divides by zero). The training state follows (pod O040:21BF,
// for all: the coordinator's 04:05 superset).
func (e *RuleState) GainExperience(c *Character, points int32) {
	classes := int32(0)
	for _, l := range c.Levels {
		if int8(l) > 0 {
			classes++
		}
	}
	if classes > 0 {
		c.Exp += uint32(points / classes)
	}
	c.CanTrain = e.CanTrain(c)
}

// LootDetect is the Detect Magic the selected member could cast on the
// treasure, 0 for none.
func (e *RuleState) LootDetect() byte {
	c := e.Acting()
	if len(e.Treasure) == 0 || c == nil {
		return 0
	}
	for i := 0; i <= 0x14; i++ {
		if s := c.Memorized[i]; s == SpDetectMagicCl || s == spDetectMagicMU {
			return byte(s)
		}
	}
	return 0
}

// HarmfulAffect (extra) names affect t when it hurts its bearer: one a
// hostile spell gives (spell table +7 is 0, as Aim reads it) or one of the
// named harms; "" otherwise.
func (e *RuleState) HarmfulAffect(t AffectType) string {
	switch t {
	case affHelpless, affSick, affMummyDisease, affRepulsed, AffPoisoned:
		return affectNames[t]
	}
	if listedAffect(t) {
		for s := 1; s <= LastSpell; s++ {
			if AffectType(e.T().Spells[s][10]) == t && e.T().Spells[s][7] == 0 {
				return e.SpellName(s)
			}
		}
	}
	return ""
}

// CanTrain = pod's canTrain: c has the experience for the next level in a
// class it still rises in (not dormant, not at its race's limit).
func (e *RuleState) CanTrain(c *Character) bool {
	for i := 0; i <= 7; i++ {
		lv := c.Levels[i]
		if int8(lv) <= 0 || Dormant(c, i) || LevelBlocked(c.Race, c.Stats[StatStr], i, lv) {
			continue
		}
		if need := e.T().XPForLevel(PorClass[i], int(int8(lv+1))); need > 0 && int32(c.Exp) >= need {
			return true
		}
	}
	return false
}

// OldAgeRule is a game's death of old age after a fight in which someone
// aged (rules.json oldAge; SotSB O089:0387, PoD O05F:03E1; POR and CotAB
// have none): those past their race's limit die, ten years before it they
// look very old.
type OldAgeRule struct {
	Limits  map[RaceID]uint16 `json:"limits"`
	Warning uint16            `json:"warning"`
}

var (
	SotsbOldAge = OldAgeRule{Limits: map[RaceID]uint16{RaceElf: 1600, RaceHalfElf: 325, RaceDwarf: 525, RaceGnome: 750, RaceHalfling: 199, RaceHuman: 120}, Warning: 10}
	PodOldAge   = OldAgeRule{Limits: map[RaceID]uint16{RaceElf: 1600, RaceHalfElf: 325, RaceDwarf: 525, RaceGnome: 525, RaceHalfling: 199, RaceHuman: 120}, Warning: 10}
)

// oldAgeDeaths: after a fight in which someone aged, the members past the
// age limit of their race die and the others mourn them; the animated dead
// (combat block Extra 1) are spared. The original's "press enter" waits are
// the fresh lines.
func (e *RuleState) oldAgeDeaths() {
	r := e.T().OldAge
	e.ageCheckDue = false
	if r == nil {
		return
	}
	for _, c := range e.Party {
		limit := r.Limits[c.Race]
		if c.Status >= StatusDead || c.Combat != nil && c.Combat.Extra == 1 || limit == 0 {
			continue
		}
		if c.Age <= limit {
			if c.Age > limit-r.Warning {
				e.CombatSay(c, "is looking very old", true)
			}
			continue
		}
		e.Emit(core.Narrated{Text: "The strain of this last battle has worn heavily on " + c.Name() + ".  "})
		pronoun := "He"
		if c.Sex != 0 {
			pronoun = "She"
		}
		e.Emit(core.Narrated{Text: pronoun + " breathes deeply one last time and crumples to the ground."})
		e.killWithMessage(c, StatusGone, "dies from age")
		mourners := 0
		var mourner *Character
		for _, m := range e.Party {
			if m.Okay == 1 && m.Status == StatusOkay && m.Side != SideMonsters {
				mourners++
				mourner = m
			}
		}
		switch mourners {
		case 0:
			e.Emit(core.Narrated{Text: "Sadly, no party member witnesses the passing of the champion."})
		case 1:
			e.Emit(core.Narrated{Text: mourner.Name() + " weeps at the loss of such a venerable companion."})
		default:
			e.Emit(core.Narrated{Text: " The remaining party members are overcome with grief as they search " + c.Name() + "'s body for valuables."})
		}
	}
}
