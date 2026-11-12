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
		Goods []ShopItem
		Terms ShopTerms
	}
	// AtTemple: services for money.
	AtTemple struct {
		Party    Party
		Services []ServicePrice
	}
	// Training: a hall is open.
	Training struct {
		Party     Party
		Price     Copper  // per level
		Allows    []Class // WordTrainable
		DualClass bool    // the adventure lets a dual-class change start here
	}
	// Looting: treasure on the ground.
	Looting struct {
		Party Party
		Items []ItemRef
		Money Copper
	}
	// AtDoor: a door to force.
	AtDoor struct {
		Party Party
		Door  Door
	}
	// Scripted: a script's command, no screen of its own.
	Scripted struct{ Party Party }
)

// ShopTerms is the zone's pricing a ruleset prices goods by (BR zone1
// priceMultiplier, sixteenths of the value).
type ShopTerms struct {
	PriceMultiplier int32
}

// ShopItem is what the adventure's shop sells, at its prices.
type ShopItem struct {
	Handle    ItemHandle // names the good (Buy's Item); the engine gives each its own
	Item      ItemRef
	Buy, Sell Copper
}

// ServicePrice is a temple service at this temple's price.
type ServicePrice struct {
	Service ServiceID
	Price   Copper
}

// DoorState is a door as the adventure defines it.
type DoorState int8

const (
	DoorOpen DoorState = iota + 1
	DoorClosed
	DoorLocked
	DoorMagic // magically locked
)

// Door: the engine offers Pick, Bash and casting whatever the party can; the
// rules decide; only the script's Opened and Failed decide what follows. A
// knock on a door is a script's option on that door, not a door rule.
type Door struct {
	State  DoorState
	Opened AreaEvent // script label on success
	Failed AreaEvent
}

// AreaEvent is a label in the current area's script.
type AreaEvent int32

// FightTerms are the script words a fight depends on.
type FightTerms struct {
	Morale     int32
	Surprised  []Side // WordSurprise
	MonsterHit int32
	PartyHit   int32
	PartyMove  int32
	NoMagic    bool
	Duel       bool
	NoExpAward bool // WordNoExpAward
	// BR zone1: the downed party stands up with 1 HP (revivesParty); the
	// beaten keep their items (skipsCombatItems).
	RevivesParty bool
	NoItemLoot   bool
	Area         AreaKind // the sort of place the fight is in
	// The party's map facing when the fight began (BR partyPosition.facing:
	// a fleeing monster runs from it).
	PartyFacing Facing
	PartyAt     image.Point // the party's map square when the fight began (por DS:6AAD/6AAE: the placement's origin)
}

// AreaKind is the sort of place a fight is in, as a ruleset tests it (BR
// zone0 areaKind). The zero value is a place with no special kind.
type AreaKind uint8

const (
	AreaPlain      AreaKind = iota
	AreaWilderness          // BR: Venus or Mars wilderness (C_Area), where a jetpack flies
)

// Fight is a battle as the engine holds it: a snapshot, value by value.
type Fight struct {
	Party Party
	Board Board
	Round int32
	Actor CharID // NoChar: between turns
	Terms FightTerms
}

func (Fight) situation()       {}
func (ShipFight) situation()   {}
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
	Size image.Point
	// Terrain is each square's terrain number, row by row (Size.X*Size.Y),
	// as the ruleset numbers it (BR: the arena cell, gdconsts.pas
	// combatTerrain's index); the ruleset says what blocks, costs and stops a
	// line. Empty is open ground.
	Terrain []TerrainID
	Height  []int8       // the same; game.Ground.Height
	Ground  []GroundKind // the same; BR bareGround
	ZeroG   bool         // BR
	Free    int32        // por FreeSteps: steps before the encounter closes
	Units   []BoardUnit
	Clouds  []Cloud
	Dungeon bool
	// Straight: the board is drawn deskewed, north is north (por's Deskew
	// option); the original's leans its north-south rows into the perspective.
	Straight bool
}

// MapSides is the area map's sides a fight's placement reads (por WallAt and
// DoorAt: the wall value 0..15 and the 2-bit door state of a cell's side).
type MapSides interface {
	WallAt(x, y, dir int) byte
	DoorAt(x, y, dir int) byte
	MapCell(x, y int) byte // the cell byte (its flags: rubble)
}

// TerrainID is a square's terrain number; a ruleset gives it meaning.
type TerrainID uint8

// GroundKind is what a square is made of; a ruleset gives it meaning.
type GroundKind int8

// BoardUnit is a combatant's place.
type BoardUnit struct {
	ID     CharID
	At     BoardPos
	Size   int32
	Facing Facing
	Side   Side
}

// Cloud is a lasting spell on a square.
type Cloud struct {
	Spell SpellID
	At    BoardPos
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
	Exp     int32        // per member
	Items   []ItemHandle // the loot's items (treasure), by handle
	Money   Purse        // the loot's money (BR: credits in Copper)
	Verdict Verdict      // how it ended: won, fled, lost; FightOn for neither (por: treasure found)
	Foes    bool         // there were foes to beat (por's "The party has won.")
	Shared  bool         // hired NPCs took their shares of the money (por hiredShare's page)
	Sharers []CharID     // who took one
}

// Turn is what BeginTurn says of the actor.
type Turn struct {
	Skipped  bool // delay gone, helpless (por takeTurn false)
	Computer bool // Decide picks its actions
	Effects  []Effect
}
