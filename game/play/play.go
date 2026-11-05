// Package play is the typed game contract (draft, 2026-11-05). It replaces
// game.Rules once every game has moved over; see MIGRATION.md.
//
// The game is purely reactive. It never does anything by itself: gobox reads
// State, switches over its type to pick the screen and controls, turns the
// player's input into one of the Actions that state publishes, and hands it
// to Do. Do refuses an action it does not allow, or applies it and returns
// the Outcome, which gobox plays (animates, voices, paces) before it reads
// State again. Questions, menus, waits, paging and screens are all gobox's.
package play

import (
	"image"
	"time"

	"gobox/game"
)

// Game is what every game implements.
type Game interface {
	State() State               // what is true now; a copy
	Do(Action) (Outcome, error) // the only way anything changes; error: refused, nothing changed
}

// State is one situation the party is in. The set is closed: gobox handles
// every type, and a new situation means a new type here first.
type State interface{ state() }

// Action is one thing the player does. Each State lists the ones it allows
// now; Do still checks.
type Action interface{ action() }

// Outcome is what an action did, in order: gobox plays the events, then
// shows the new State.
type Outcome struct {
	Events []game.Event
}

// Choice is an allowed action as gobox offers it: the rules' words and the
// original's hotkey. gobox builds its menus and dialogs from these.
type Choice struct {
	Label string
	Key   rune
	Do    Action
}

// Common is what every state shows besides its own view.
type Common struct {
	Party    []game.Member
	Selected int
	Purse    map[string]int
	Where    game.Where
	Area     string      // the area's name as the player knows it
	Lore     []game.Lore // the journal, newest first
	Choices  []Choice    // the actions allowed now
}

// The states. Each situation is one explicit type; similar screens share one.

// Assembling: the party is put together before play, and characters are
// made (Draft set) or changed.
type Assembling struct {
	Common
	Roster []string        // saved characters
	Draft  *game.SheetPage // a character being made or changed; nil none
}

// Exploring: the party walks a map (dungeon, town or wilderness).
type Exploring struct {
	Common
	Map ExploreMap
}

// ExploreMap is what gobox needs to draw and walk the 3D view and the
// automap. Whether a step goes through is the game's call (Do(Step)).
type ExploreMap struct {
	Size    image.Point
	Walls   *game.Walls
	Automap game.Automap
}

// Camping: the party has made camp (rest, memorize, scribe, fix, order).
type Camping struct{ Common }

// Fighting: a battle is on.
type Fighting struct {
	Common
	Combat game.Combat // the field, the units and whose turn it is
}

// Scene: a script shows text and pictures, someone speaks, or a handout is
// read. The script's place is the game's state; Continue or a Reply moves it.
type Scene struct {
	Common
	Speaker game.Picture
	Text    []game.TextRun
	Picture []game.Picture
	Handout string   // a journal handout's title; Lines is its layout
	Lines   []string // journal.Layout's
}

// Trading: a shop, temple, trainer or vault is open.
type Trading struct {
	Common
	Shop   game.Shop
	Offers []game.Offer // the appraisals of gems and jewels
}

// Looting: treasure lies before the party.
type Looting struct {
	Common
	Loot game.Loot
}

// Over: the game has ended.
type Over struct {
	Common
	Won bool
}

func (Assembling) state() {}
func (Exploring) state()  {}
func (Camping) state()    {}
func (Fighting) state()   {}
func (Scene) state()      {}
func (Trading) state()    {}
func (Looting) state()    {}
func (Over) state()       {}

// The actions. Every game uses these; a game-specific verb gets its own
// type here. Which ones apply now is the state's Choices.
type (
	// Anywhere a member's goods are shown.
	Ready   struct{ Member, Item, Count int } // Count of a stack (0: all)
	Unready struct{ Member, Item int }
	Drop    struct{ Member, Item int }
	Give    struct{ From, To, Item int }
	Halve   struct{ Member, Item int }
	Join    struct{ Member, Item int }
	UseItem struct {
		Member, Item, Target int
		Ats                  []image.Point
		Spell                int // a scroll's spell, 0 the first
	}
	Cast struct {
		Member, Spell, Target int
		Ats                   []image.Point
	}
	SelectMember struct{ Member int }
	Order        struct{ From, To int }
	AddNote      struct{ Text string }
	RemoveNote   struct{ ID int }

	// Assembling.
	AddCharacter struct {
		Roster int
		From   string
	} // From: another game's roster, "" this one's
	RemoveCharacter struct{ Roster int }
	DropMember      struct{ Member int }
	NewCharacter    struct{}
	SetChoice       struct {
		Field string
		N     int
		Text  string
	} // race, class, alignment, portrait and icon pool IDs, colours ...
	SetStat struct {
		Stat int
		Text string
	}
	Adjust           struct{ Stat, By int }
	PointBuy         struct{ Scores []int }
	Reroll           struct{}
	KeepCharacter    struct{}
	DiscardCharacter struct{}
	ChangeClass      struct{ Member, Class int }
	SetDifficulty    struct{ N int }
	Begin            struct{}

	// Exploring.
	Step          struct{ Facing int } // one square; refused at a wall
	Face          struct{ Facing int }
	Search        struct{}
	Look          struct{}
	MakeCamp      struct{}
	Call          struct{ Member int }
	AdvanceTimeBy struct{ D time.Duration } // also in Camping: rest; the outcome tells what happened meanwhile

	// Camping.
	Memorize   struct{ Member, Spell int }
	Forget     struct{ Member, Spell int }
	Scribe     struct{ Member, Spell int }
	Fix        struct{ Member int }
	Bandage    struct{ Member int }
	LayOnHands struct{ Member int }
	BreakCamp  struct{}

	// Fighting. EndTurn's outcome carries the monsters' moves up to the
	// next member's turn.
	MoveTo     struct{ To image.Point }
	Attack     struct{ Target int }
	Shoot      struct{ At image.Point }
	DoManeuver struct {
		ID int
		At image.Point
	}
	TurnUndead struct{}
	Guard      struct{}
	Delay      struct{}
	Quick      struct{ All bool }
	Flee       struct{}
	EndTurn    struct{}

	// Scene.
	Continue struct{}
	Reply    struct {
		N    int
		Text string
	} // Text: a typed answer

	// Trading.
	Buy      struct{ Member, Item int }
	Sell     struct{ Member, Item int }
	Identify struct{ Member, Item int }
	Appraise struct {
		Member, Item int
		Keep         bool
	}
	Repair    struct{ Member, Item int }
	Reload    struct{ Member, Item int }
	Train     struct{ Member int }
	Heal      struct{ Service, Member int }
	PoolCoins struct{ Member int }

	// Looting.
	Take struct {
		Item int  // -1: the money only
		All  bool // the money and every item
	}

	// Trading, Looting, Scene's handout.
	Leave struct{}

	// Over and anywhere in play: back to Assembling.
	EndGame struct{}
)

func (Ready) action()            {}
func (Unready) action()          {}
func (Drop) action()             {}
func (Give) action()             {}
func (Halve) action()            {}
func (Join) action()             {}
func (UseItem) action()          {}
func (Cast) action()             {}
func (SelectMember) action()     {}
func (Order) action()            {}
func (AddNote) action()          {}
func (RemoveNote) action()       {}
func (AddCharacter) action()     {}
func (RemoveCharacter) action()  {}
func (DropMember) action()       {}
func (NewCharacter) action()     {}
func (SetChoice) action()        {}
func (SetStat) action()          {}
func (Adjust) action()           {}
func (PointBuy) action()         {}
func (Reroll) action()           {}
func (KeepCharacter) action()    {}
func (DiscardCharacter) action() {}
func (ChangeClass) action()      {}
func (SetDifficulty) action()    {}
func (Begin) action()            {}
func (Step) action()             {}
func (Face) action()             {}
func (Search) action()           {}
func (Look) action()             {}
func (MakeCamp) action()         {}
func (Call) action()             {}
func (AdvanceTimeBy) action()    {}
func (Memorize) action()         {}
func (Forget) action()           {}
func (Scribe) action()           {}
func (Fix) action()              {}
func (Bandage) action()          {}
func (LayOnHands) action()       {}
func (BreakCamp) action()        {}
func (MoveTo) action()           {}
func (Attack) action()           {}
func (Shoot) action()            {}
func (DoManeuver) action()       {}
func (TurnUndead) action()       {}
func (Guard) action()            {}
func (Delay) action()            {}
func (Quick) action()            {}
func (Flee) action()             {}
func (EndTurn) action()          {}
func (Continue) action()         {}
func (Reply) action()            {}
func (Buy) action()              {}
func (Sell) action()             {}
func (Identify) action()         {}
func (Appraise) action()         {}
func (Repair) action()           {}
func (Reload) action()           {}
func (Train) action()            {}
func (Heal) action()             {}
func (PoolCoins) action()        {}
func (Take) action()             {}
func (Leave) action()            {}
func (EndGame) action()          {}
