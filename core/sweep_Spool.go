package core

// Additions of sweep Spool (fold E-R6): the paladin's powers, the door menu's
// questions and the area map's terrain.

type (
	// PaladinHeal is the paladin Who laying hands on Target: 2 hp per paladin
	// level (CotAB PaladinHeal), and the day's use spent. The Outcome holds
	// Healed when Target was healed, none when it is unaffected.
	PaladinHeal struct{ Who, Target CharID }
	// PaladinCure is the paladin Who curing Target of the temple's diseases
	// (CotAB PaladinCureDisease), and the day's use spent.
	PaladinCure struct{ Who, Target CharID }
)

func (PaladinHeal) action()  {}
func (PaladinCure) action() {}

type (
	// Diseased asks whether Who has one of the diseases PaladinCure cures (Yes).
	Diseased struct{ Who CharID }
	// PartyHasSkill asks whether a party member holds thief skill number N
	// above 0 (ovr15:0236); Yes.
	PartyHasSkill struct{ N int }
	// MapTerrain asks the area map's terrain number at map square (X, Y) (Value).
	MapTerrain struct{ X, Y int }
)

func (Diseased) check()      {}
func (PartyHasSkill) check() {}
func (MapTerrain) check()    {}
