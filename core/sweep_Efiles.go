package core

// Additions of sweep Efiles (fold E-R6): what por's character files ask of
// the rules. The files' bytes stay in gobox/char; a character crosses as a
// CharacterFile (Rules.Import, Export).

type (
	// MakeMonster makes a new character from the data set's monster Of (por
	// LoadMonster, the ruleset's MonsterSource); the Outcome holds
	// MonsterMade. Error: a monster the data set lacks.
	MakeMonster struct{ Of MonsterID }
	// DressNPC gives a monster joining the party its random id and a
	// portrait (por AddMonsterToParty's dice); NoChar dresses a new blank
	// character (a monster that failed to load). The Outcome holds NPCDressed.
	DressNPC struct{ Who CharID }
	// SetRoster stores the roster number in Who's file.
	SetRoster struct {
		Who    CharID
		Roster RosterID
	}
	// EmptyMember empties the ruleset's one placeholder character (the
	// selected member of an empty party); the Outcome holds EmptyMemberIs.
	EmptyMember struct{}
)

func (MakeMonster) action() {}
func (DressNPC) action()    {}
func (SetRoster) action()   {}
func (EmptyMember) action() {}

type (
	MonsterMade   struct{ Who CharID }
	NPCDressed    struct{ Who CharID }
	EmptyMemberIs struct{ Who CharID }
)

func (MonsterMade) effect()   {}
func (NPCDressed) effect()    {}
func (EmptyMemberIs) effect() {}
