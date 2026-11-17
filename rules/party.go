package rules

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

// JoinParty puts Who at the end of the party with the first free icon
// slot, and makes it the selected member (por AddToParty).
type JoinParty struct{ Who CharID }

func (JoinParty) action() {}

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

// MoveExtremes asks the fastest (Answer.High) and slowest (Answer.Low)
// party member's movement, haste doubling and slow halving (the original's
// MoveExtremes, eclmenu.pas: the encounter menu's flee test). Not
// Census{Fact: FactMove}: CheckParty's census reads each member's movementCur
// as it is, hasted or not (EclCheckParty).
type MoveExtremes struct{}

func (MoveExtremes) check() {}

// MarkImported flags Who as a character added from another game, as the game's own rules
// mark one (COTAB: the flag the scripts read, which its Hillsfar import sets); a game with
// no such mark leaves it be.
type MarkImported struct{ Who CharID }

func (MarkImported) action() {}

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
type PickKind struct {
	Race, Gender, Class, Align string
	Weapon                     ItemID // XXVc: the warrior's weapon specialization (0: none)
	Skill                      Skill  // XXVc: the general skill the points go to ("": none)
}

func (ForgetParty) action() {}

func (PickKind) action() {}

// Picked is the creation picks PickKind asked.
type Picked struct {
	Picks map[FieldID]int
	Spend map[FieldID][]int // the points fields' options to spend on, a point each in turn until none goes (nil: none)
}

func (Picked) effect() {}

// PartySize: the members of the party (Value), extras too; core.MemberCount
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

// Purge: dropping Who from the party deletes it for good (Yes; Text is the line said as it goes), as BR's Purge Character
// (chargen.pas DropCharacterMenu does); no: the drop keeps its saved files, as the Pool's.
type Purge struct{ Who CharID }

func (Purge) check() {}

func (CanJoin) check() {}

func (InParty) check() {}

func (ControlledBefore) check() {}

// PaladinHeal is the paladin Who laying hands on Target: 2 hp per paladin
// level (CotAB PaladinHeal), and the day's use spent. The Outcome holds
// Healed when Target was healed, none when it is unaffected.
type PaladinHeal struct{ Who, Target CharID }

// PaladinCure is the paladin Who curing Target of the temple's diseases
// (CotAB PaladinCureDisease), and the day's use spent.
type PaladinCure struct{ Who, Target CharID }

func (PaladinHeal) action() {}

func (PaladinCure) action() {}

// Diseased asks whether Who has one of the diseases PaladinCure cures (Yes).
type Diseased struct{ Who CharID }

// PartyHasSkill asks whether a party member holds thief skill number N
// above 0 (ovr15:0236); Yes.
type PartyHasSkill struct{ N int }

func (Diseased) check() {}

func (PartyHasSkill) check() {}

// OwedFact is the pick the member still owes (Fact: PowerOwed; no Options: none).
type OwedFact struct{}

// PowerOwed is one choice a member owes: one of Options.
type PowerOwed struct {
	What    string // "an at-will power"
	Options ManeuverCards
}

// Toughen keeps the ruleset's Toughness in Who's maximum hit points.
type Toughen struct{ Who CharID }

// TakePower adds Power to Who's picked powers.
type TakePower struct {
	Who   CharID
	Power int
}

func (OwedFact) fact() {}

func (PowerOwed) value() {}

func (Toughen) action() {}

func (TakePower) action() {}

// FiguresFact is the lines Who's rule set adds to its sheet (Figures; nil: none).
type FiguresFact struct{}

// Figures is FiguresFact's answer.
type Figures []Figure

func (FiguresFact) fact() {}

func (Figures) value() {}

// PortraitFact is Who's portrait (Portrait).
type PortraitFact struct{}

// Portrait: a dropped PNG (nil: none, the pool's Head and Body pictures
// show; 0: not picked), Head and Body the ruleset's own picture numbers.
type Portrait struct {
	PNG        PNG
	Head, Body int
}

// SetPortrait: Who's portrait is the dropped PNG (nil: back to Head and Body).
type SetPortrait struct {
	Who CharID
	PNG PNG
}

func (PortraitFact) fact()  {}
func (Portrait) value()     {}
func (SetPortrait) action() {}

// Modifiable asks whether Who may be modified: no experience of its own,
// that is the adventure's starting experience (Yes).
type Modifiable struct{ Who CharID }

func (Modifiable) check() {}

// PartyAuto: a member of the party fights by the computer (Yes): its turns poll the keys (por aiKeys).
type PartyAuto struct{}

func (PartyAuto) check() {}

// CanQuickMagic: the ruleset has the key Alt-M during a computer turn (QuickMagic): por's, not Buck Rogers'.
type CanQuickMagic struct{}

func (CanQuickMagic) check() {}

// MoveMember: Who takes place To in the party (camp's order); ErrNotNow out of camp or out of range.
type MoveMember struct {
	Who CharID
	To  int32
}

func (MoveMember) action() {}

// PoolPic names one of a character's shared-pool picture picks.
type PoolPic int

const (
	PoolPortraitHead PoolPic = iota
	PoolPortraitBody
	PoolIconHead
	PoolIconBody
)

// poolPicNames are the picks' play.SetChoice field names (Draft.Pics).
var poolPicNames = map[string]PoolPic{"portraithead": PoolPortraitHead, "portraitbody": PoolPortraitBody, "iconhead": PoolIconHead, "iconbody": PoolIconBody}

// PoolPicNamed is the pick a play.SetChoice field names; false: not a pick.
func PoolPicNamed(field string) (PoolPic, bool) {
	p, ok := poolPicNames[field]
	return p, ok
}

// PoolPicsFact is Who's shared-pool picks (PoolPics).
type PoolPicsFact struct{}

// PoolPics are a character's shared-pool picks; "" is none.
type PoolPics struct {
	PortraitHead, PortraitBody, IconHead, IconBody PicID
}

// SetPoolPic: Who's pick Pic is ID.
type SetPoolPic struct {
	Who CharID
	Pic PoolPic
	ID  PicID
}

// IconPicsFact is Who's own combat icon (IconPics).
type IconPicsFact struct{}

// IconPics is a character's combat icon in the ruleset's own pictures: the head
// and body numbers, the size (1 small, 2 large) and the parts' colours.
type IconPics struct {
	Head, Body, Size int
	Colours          [IconParts]IconColour
	PNG              PNG // a dropped PNG in place of the pictures (nil: none)
}

// SetIconPics: Who's own combat icon is Pics (the icon designer's edits, and its undo).
type SetIconPics struct {
	Who  CharID
	Pics IconPics
}

func (SetIconPics) action() {}

// ResetIcon: Who's combat icon is the ruleset's own again: no dropped PNG,
// the default colours and the race's size (the icon editor's Remove).
type ResetIcon struct{ Who CharID }

func (ResetIcon) action() {}

func (PoolPicsFact) fact() {}
func (PoolPics) value()    {}
func (SetPoolPic) action() {}
func (IconPicsFact) fact() {}
func (IconPics) value()    {}

// PartyFact is the party in marching order (Party; Who is NoChar).
type PartyFact struct{}

// ModFact is Who's modifier of ability Ability (Number).
type ModFact struct{ Ability Ability }

// Abilities are the six abilities in the order of a sheet's rows.
var Abilities = [6]Ability{Strength, Intelligence, Wisdom, Dexterity, Constitution, Charisma}

// ClassByNumber is the ruleset's class of the original's class number N
// (Class; ClassNone outside the table).
type ClassByNumber struct{ N int }

// PaladinCan asks whether Who may lay on hands now (Cure false) or cure
// disease now (Cure true): Yes.
type PaladinCan struct {
	Who  CharID
	Cure bool
}

func (PartyFact) fact()     {}
func (ModFact) fact()       {}
func (ClassByNumber) fact() {}
func (PaladinCan) check()   {}
func (Party) value()        {}
func (Class) value()        {}

// GainExperience gives each of Whom Points experience, split evenly among
// its classes (the script's GainExperience; ruling S8: as given, any member
// whatever its state).
type GainExperience struct {
	Whom   Party
	Points int32
	// Whole: the script gave it to the whole party (not one member); a ruleset with its own words for it names which.
	Whole bool
}

// The Outcome may carry a Narrated: the ruleset's own words for the award; none: the engine's.
func (GainExperience) action() {}

// HealParty puts every party member back at full hit points and well (the dead and stoned too):
// what an ending leaves behind.
type HealParty struct{}

func (HealParty) action() {}
