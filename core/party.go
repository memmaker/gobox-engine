package core

// The party and its roster: members, monsters, hired NPCs, reactions, creation picks.

// MemberCardFact is what the party card shows of a member the sheet does
// not (MemberCard).
type MemberCardFact struct{}

// MemberCard: Party: a member's card (not a monster loaded for a fight),
// Class as the card words it, Ready spells memorized of Slots, the
// dropped Portrait (nil: none).
type MemberCard struct {
	Party        bool
	Class        string
	Ready, Slots int
	Portrait     PNG
}

func (MemberCardFact) fact() {}

func (MemberCard) value() {}

// MakeMonster makes a new character from the data set's monster Of (por
// LoadMonster, the ruleset's MonsterSource); the Outcome holds
// MonsterMade. Error: a monster the data set lacks.
type MakeMonster struct{ Of MonsterID }

// DressNPC gives a monster joining the party its random id and a
// portrait (por AddMonsterToParty's dice); NoChar dresses a new blank
// character (a monster that failed to load). The Outcome holds NPCDressed.
type DressNPC struct{ Who CharID }

// SetRoster stores the roster number in Who's file.
type SetRoster struct {
	Who    CharID
	Roster RosterID
}

// EmptyMember empties the ruleset's one placeholder character (the
// selected member of an empty party); the Outcome holds EmptyMemberIs.
type EmptyMember struct{}

func (MakeMonster) action() {}

func (DressNPC) action() {}

func (SetRoster) action() {}

func (EmptyMember) action() {}

type MonsterMade struct{ Who CharID }

type NPCDressed struct{ Who CharID }

type EmptyMemberIs struct{ Who CharID }

func (MonsterMade) effect() {}

func (NPCDressed) effect() {}

func (EmptyMemberIs) effect() {}

// Reaction is one of the rule set's reactions a character has, with its policy.
type Reaction struct {
	ID     int
	Name   string
	Policy Policy
}

// Reactions are Who's reactions (ReactionsFact).
type Reactions []Reaction

// ReactionsFact is Who's reactions and its policy for each (Reactions).
type ReactionsFact struct{}

func (Reactions) value() {}

func (ReactionsFact) fact() {}

// PartyMoveRange asks the fastest (Answer.High) and slowest (Answer.Low)
// party member's movement, haste doubling and slow halving (por PartyHpRange).
type PartyMoveRange struct{}

func (PartyMoveRange) check() {}

// EmptyParty sets the party's member count to 0 (the roster stays).
type EmptyParty struct{}

func (EmptyParty) action() {}

// ForgetParty drops the party and the pool of characters (a load fills
// them again).
type ForgetParty struct{}

// PickKind asks the creation picks (the field of Template.Picks to the
// option's place) of a kind named as the game's data names it (Race,
// Gender, Class, Align); the Outcome holds Picked. Error: a name the
// ruleset does not know.
type PickKind struct{ Race, Gender, Class, Align string }

func (ForgetParty) action() {}

func (PickKind) action() {}

// Picked is the creation picks PickKind asked.
type Picked struct{ Picks map[FieldID]int }

func (Picked) effect() {}

// PartySize: the members of the party (Value), extras too; core.WordPartyCount
// is the record's count, which they are not in.
type PartySize struct{}

// CanJoin: Who may join the party (Yes); the Refusal says why not is the rules' own.
type CanJoin struct{ Who CharID }

// InParty: a character of Who's name and random id is in the party (Yes).
type InParty struct{ Who CharID }

// ControlledBefore: the members the player controls (hired ones not)
// ahead of the first InParty match of Who, or all of them (Value).
type ControlledBefore struct{ Who CharID }

func (PartySize) check() {}

func (CanJoin) check() {}

func (InParty) check() {}

func (ControlledBefore) check() {}
