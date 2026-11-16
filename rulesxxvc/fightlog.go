package rulesxxvc

import (
	"strconv"

	"gobox/rules"
)

// The words of cbtacts.pas ProcessAttackResult's side panel and CharHlp.PrintCreatureMessage (the creature's name, then the message).
const (
	msgAttacks    = "Attacks"
	msgBackstabs  = "-Backstabs-"
	msgFromBehind = "(from behind) "
	msgMisses     = "and Misses"
	msgHitting    = "Hitting for "
	msgPoint      = " point "
	msgPoints     = " points "
	msgOfDamage   = "of damage"
	msgGoesDown   = "goes down"
	msgDying      = "and is Dying"
	msgGotAway    = "Got Away" // cbtacts.pas AttemptFlee: PutOutOfPlayWithStatus(StFleeing, 'Got Away')
	msgKilled     = "is killed"
)

// attackLines is ProcessAttackResult's two panel lines for one attack of a on d: the attacker's, then the defender's
// with the blow (isHit: the damage shown, else "and Misses"). manner is the attack's: plain, from behind or backstab.
func attackLines(a, d *Character, backstab, behind, isHit bool, shown int) []rules.Effect {
	first := msgAttacks
	if backstab {
		first = msgBackstabs
	}
	second := ""
	if behind && !backstab {
		second = msgFromBehind
	}
	if isHit {
		points := msgPoints
		if shown == 1 {
			points = msgPoint
		}
		second += msgHitting + strconv.Itoa(shown) + points + msgOfDamage
	} else {
		second += msgMisses
	}
	return []rules.Effect{rules.Narrated{Text: a.Name + " " + first}, rules.Narrated{Text: d.Name + " " + second}}
}

// downLines is what ProcessAttackResult says of a defender that went down: "goes down", "and is Dying" for a dying
// one, "is killed" for a dead one (not for one already out of play).
func downLines(d *Character) []rules.Effect {
	out := []rules.Effect{rules.Narrated{Text: d.Name + " " + msgGoesDown}}
	switch d.Status {
	case StDying:
		out = append(out, rules.Narrated{Text: msgDying, More: true})
	case StDead, StGone:
		out = append(out, rules.Narrated{Text: d.Name + " " + msgKilled})
	}
	return out
}
