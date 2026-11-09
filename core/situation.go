package core

import "image"

// Situation is where an action happens: everything of the engine's the rules
// need, passed in, never asked back for. The set is closed.
type Situation interface{ situation() }

// Party is the characters in order. The engine keeps the order and who is
// selected; por's e.party mixed monsters in, here it does not.
type Party []CharID

type (
	// Assembling: at the guild, building the party.
	Assembling struct{ Party Party }
	// Creating: a Draft is open (Rules.Draft).
	Creating struct{ Party Party }
	// Adventuring: walking; items and spells outside a fight.
	Adventuring struct {
		Party   Party
		NoMagic bool // WordNoMagic
	}
	// Camping: camp is made.
	Camping struct {
		Party   Party
		NoMagic bool
	}
	// Shopping: a shop is open with the adventure's goods.
	Shopping struct {
		Party Party
		Goods []Item
		Terms ShopTerms
	}
	// AtTemple: services for money.
	AtTemple struct {
		Party Party
		Terms ShopTerms
	}
	// Training: a hall is open.
	Training struct {
		Party  Party
		Terms  ShopTerms
		Allows []Class // WordTrainable
	}
	// Looting: treasure on the ground.
	Looting struct {
		Party Party
		Items []Item
		Money Purse
	}
	// AtDoor: a door to force.
	AtDoor struct{ Party Party }
	// Scripted: a script's command, no screen of its own.
	Scripted struct{ Party Party }
)

// ShopTerms are the script words a price depends on.
type ShopTerms struct {
	PriceFactor int32 // WordPriceFactor: shop pricing stays per ruleset (RULE2.md (e))
}

// FightTerms are the script words a fight depends on.
type FightTerms struct {
	Morale     int32
	Surprised  []Side // WordSurprise
	MonsterHit int32
	PartyHit   int32
	PartyMove  int32
	NoMagic    bool
	Duel       bool
}

// Fight is a battle as the engine holds it: a snapshot, value by value.
type Fight struct {
	Party Party
	Board Board
	Round int32
	Actor CharID // NoChar: between turns
	Terms FightTerms
}

func (Fight) situation()       {}
func (Assembling) situation()  {}
func (Creating) situation()    {}
func (Adventuring) situation() {}
func (Camping) situation()     {}
func (Shopping) situation()    {}
func (AtTemple) situation()    {}
func (Training) situation()    {}
func (Looting) situation()     {}
func (AtDoor) situation()      {}
func (Scripted) situation()    {}

// Board is the squares and who stands where. The engine builds it (por
// combatmap.go, eclbr combatmap.go) and moves units on it by Effects.
type Board struct {
	Size    image.Point
	Blocked []bool // row by row, Size.X*Size.Y
	Units   []BoardUnit
	Clouds  []Cloud
	Dungeon bool
}

// BoardUnit is a combatant's place.
type BoardUnit struct {
	ID     CharID
	At     image.Point
	Size   int32
	Facing Facing
	Side   Side
}

// Cloud is a lasting spell on a square.
type Cloud struct {
	Spell SpellID
	At    image.Point
}

// Verdict is how a fight stands (por sidesCheck, BR endFight).
type Verdict int

const (
	FightOn Verdict = iota
	PartyWon
	PartyLost
	PartyFled
)

// FightResult is what the rules made of a finished fight (por AfterCombat
// giveExp, the BR loot page). Script words the engine sets from it: Beaten, CombatEnd.
type FightResult struct {
	Effects []Effect
	Beaten  int32
	Exp     int32 // per member
	Items   []Item
	Money   Purse
}

// Turn is what BeginTurn says of the actor.
type Turn struct {
	Skipped  bool // delay gone, helpless (por takeTurn false)
	Computer bool // Decide picks its actions
	Effects  []Effect
}
