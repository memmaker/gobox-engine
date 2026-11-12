package core

// Additions of sweep Eids6 (fold E-R6): casting out of a fight.

// CastAim is how a spell out of a fight finds its target before it is cast.
type CastAim int32

const (
	CastAimSelf       CastAim = iota // no question (the caster, the party, or none)
	CastAimCombatOnly                // a combat-only spell or item: "Lose it?" / "Use it?"
	CastAimMember                    // a member it goes on: "Cast Spell on whom"
)

// CastCardFact is spell number Number as a cast out of a fight sees it
// (CastCard); Who is NoChar. Number is the ruleset's spell number (SpellLine.Number).
type CastCardFact struct{ Number int }

// CastCard: the spell's Name ("" for an item's power) and its Aim.
type CastCard struct {
	Name string
	Aim  CastAim
}

// OpenCast: Who opens camp's Cast (por CastMenu's start); refused
// (ErrNotNow) when Who cannot cast here, as the rules word it to the party.
type OpenCast struct{ Who CharID }

// CampCast: Who casts memorized spell Spell out of a fight; Sure answers
// a combat-only spell's "Lose it?", Target is the member it goes on (NoChar: none).
type CampCast struct {
	Who, Target CharID
	Spell       SpellID
	Sure        bool
}

func (CastCardFact) fact() {}
func (CastCard) value()    {}
func (OpenCast) action()   {}
func (CampCast) action()   {}
