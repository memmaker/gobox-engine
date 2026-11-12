package core

// Sweep Sviews2: what por's rules package asks of the ruleset through
// Fact and Check instead of reaching the ruleset's own records.

// PartyFact is the party in marching order (Party; Who is NoChar).
type PartyFact struct{}

// SpellByNumber is the ruleset's id of the original's spell number N
// (SpellID; "" outside the table).
type SpellByNumber struct{ N int }

// ItemByNumber is the ruleset's id of the original's item type N
// (ItemID; the Fact errs when the ruleset has none).
type ItemByNumber struct{ N int }

// ModFact is Who's modifier of ability Ability (Number).
type ModFact struct{ Ability Ability }

// PaladinCan asks whether Who may lay on hands now (Cure false) or cure
// disease now (Cure true): Yes.
type PaladinCan struct {
	Who  CharID
	Cure bool
}

func (PartyFact) fact()     {}
func (SpellByNumber) fact() {}
func (ItemByNumber) fact()  {}
func (ModFact) fact()       {}
func (PaladinCan) check()   {}

func (Party) value()   {}
func (SpellID) value() {}
func (ItemID) value()  {}

// Abilities are the six abilities in the order of a sheet's rows.
var Abilities = [6]Ability{Strength, Intelligence, Wisdom, Dexterity, Constitution, Charisma}
