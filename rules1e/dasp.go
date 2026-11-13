package rules1e

import (
	"gobox/rules"
)

// DASP (extra): the Dubtrain Angband Sound Pack, in data/assets/dasp (its
// WAVs and sound.cfg, which names the files of each Angband sound event),
// plays in place of the original's effects (daspFx) and on events the
// original kept silent (shops, training, doors, treasure, affects ...),
// while the Esc menu's "DASP Sfx" is on (Engine.Dasp). Each event plays one
// of its files at random, mixed over the speaker, Tandy or Amiga output on
// the effects' volume; the music plays on under it. Ctrl-S silences it too.

// daspAffects names the event of an affect a character gets.
var daspAffects = map[AffectType]string{
	AffBless:           "blessed",
	affPrayer:          "blessed",
	affPrayerBonus:     "blessed",
	AffCurse:           "cursed",
	AffBestowCurse:     "cursed",
	AffProtEvil:        "prot_evil",
	affProtGood:        "prot_evil",
	affProtEvilRadius:  "prot_evil",
	affProtGoodRadius:  "prot_evil",
	affResistCold:      "res_cold",
	affResistFire:      "res_fire",
	AffCharmed:         "confused",
	affEnlarged:        "hero",
	affStrength:        "hero",
	affShield:          "shield",
	affProtMissiles:    "shield",
	affSilenced:        "stun",
	affDetectInvisible: "see_invis",
	affInvisible:       "teleport",
	affBlink:           "teleport",
	affEnfeebled:       "drain_stat",
	affWeakened:        "drain_stat",
	affCauseDisease:    "drain_stat",
	affCoughing:        "drugged",
	affHelpless:        "paralyzed",
	AffHeld:            "paralyzed",
	affHeldFast:        "paralyzed",
	affParalyse2d8:     "paralyzed",
	affParalyseNotElf:  "paralyzed",
	affParalyse1d9:     "paralyzed",
	affAsleep:          "slow",
	AffSlow:            "slow",
	AffBlind:           "blind",
	AffHaste:           "speed",
	affRepulsed:        "flee",
	AffPoisoned:        "poisoned",
	affPoisonDamage:    "poisoned",
}

// hitEvent is event, or a monster's hit when the target is one of the party.
func (e *RuleState) hitEvent(target *Character, event string) string {
	if e.partyMember(target) {
		return "mon_hit"
	}
	return event
}

// fallEvent: a monster's death cry, or the party member's death.
func (e *RuleState) fallEvent(c *Character) string {
	if e.partyMember(c) {
		return "death"
	}
	return "kill"
}

// MoneySound plays the coins of picking up p, by its gold.
func (e *RuleState) MoneySound(p Purse) {
	switch {
	case p.Gold == 0 && p.Copper == 0:
	case p.Gold < 200:
		e.Emit(rules.Sounded{Cue: rules.SoundCue("money1")})
	case p.Gold < 600:
		e.Emit(rules.Sounded{Cue: rules.SoundCue("money2")})
	default:
		e.Emit(rules.Sounded{Cue: rules.SoundCue("money3")})
	}
}
