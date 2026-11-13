package rules1e

// spellEffectCotabFirst runs the effect of the unified spells 68..103, the ids
// cotab has first (spellids_games.go), ported from its spellFx (fold T3f).
// false: the spell is not one of them (or not ported yet).
func (e *RuleState) spellEffectCotabFirst(spell SpellID) bool {
	switch spell {
	case spCauseSeriousWounds:
		e.applySpell(int(e.CurSpell), 0, 0, e.RollDice(2, 8)+1, DmgMagic, "") // cotab O0C2:4185
	case spPoison:
		e.fxPoison()
	case spCauseCriticalWounds:
		e.applySpell(int(e.CurSpell), 0, 0, e.RollDice(3, 8)+3, DmgMagic, "") // cotab O0C2:4431
	case spAnimateDeadPower:
		e.spellAnimateDead() // cotab O0C2:2E19, POR's ovr23 Animate Dead
	case spScrollOfProtection, spUnnamed96, spPotionOfInvisibility:
		e.applySpell(int(e.CurSpell), 0, 0, 0, 0, "") // cotab O0C2:569E, 56CC, 56FA: the spell record alone
	case spUnnamed98:
		e.fxSpell98()
	case spBestowCurseMage:
		e.applySpell(int(e.CurSpell), 0, 0, 0, 0, "has been cursed!") // Felix's ruling; cotab O0C2:371B
	case spEntangle:
		e.fxEntangle()
	case spCloudKill:
		e.createCloud(CloudPoison)
	default:
		return false
	}
	return true
}

// Entangle's duration (cotab O0C2:46BB): 6 + 3 rounds per caster level.
const (
	entangleBase     = 6
	entanglePerLevel = 3
)

// fxEntangle = pod O0BB:4365 (Felix's ruling): outdoors only, each target
// that fails its save vs spell is Befuddled (loses its turns) for cotab's
// duration.
func (e *RuleState) fxEntangle() {
	if e.Dungeon {
		return
	}
	rounds := entangleBase + entanglePerLevel*e.casterLevel(int(e.CurSpell)) // POR's casterLevel knows no druids
	for i := 1; i <= e.NSpellTargets; i++ {
		if t := e.tgt(i); t != nil {
			saved := e.SavingThrow(t, SaveVsSpell, 0)
			e.AddAffectMsg(t, affBefuddledPod, rounds, 0, 0, saveNegate, saved, "is entangled")
		}
	}
}

// fxPoison = cotab O0C2:42D0: the spell's affect on the targets; then the
// caster's own target is tried against the poison sting: when nothing there
// blocks it, the sting handler runs for the caster.
func (e *RuleState) fxPoison() {
	e.applySpell(int(e.CurSpell), 0, 0, 0, DmgMagic, "")
	sel := e.Acting()
	e.Rv.AffVal = affPoisonBite
	e.AffectGroup(sel.Combat.Target, grpAddAffect)
	if e.Rv.AffVal == affPoisonBite {
		e.callAffect(affPoisonBite, sel, nil, 0) // the affect argument is an unset local
	}
}
