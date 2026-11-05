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
	"image/color"
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

// Outcome is what an action did: the facts, in the order they happened.
// gobox decides how to show them, how long each takes and what waits for a
// key; nothing here says.
type Outcome struct {
	Events []Event
}

// Allowed is what a state lets the player do now. Plain is every action
// that needs nothing picked on the map, with its parameters filled in (one
// Attack per target, one Buy per member and good, ...). Aimed are those that
// take squares; Moves the squares a combatant may walk to. gobox offers
// exactly these; Do checks again.
type Allowed struct {
	Plain []Action
	Aimed []Aimed
	Moves []game.Reach
}

// Aimed is an action that takes squares: gobox lets the player pick Targets
// of Squares (each Area squares round), then sends Action with Ats filled.
type Aimed struct {
	Action  Action // Cast, UseItem, Shoot or DoManeuver with Ats empty
	Squares []image.Point
	Targets int // squares to pick, 1 for one
	Area    int // radius hit round each, 0 the square alone
}

// Common is what every state shows besides its own view.
type Common struct {
	Party    []Member
	Selected int
	Purse    map[string]int
	Where    game.Where
	Area     string      // the area's name as the player knows it
	Lore     []game.Lore // the journal, newest first
	Quests   []game.Quest
	Allowed  Allowed
}

// Member is one character as the rules know it.
type Member struct {
	Name, Race, Class, Alignment, Status string
	Levels                               []int
	Stats, BaseStats                     map[string]int // "str", "int", ...
	Str00                                int            // exceptional strength: 0 none, 1..100
	HP, MaxHP, AC, THAC0, Move           int
	Exp, LevelExp, NextExp               int
	Items                                []game.Item
	ItemCap                              int // 0 no limit
	Spells                               []game.SpellRef
	Slots, SlotsMax                      map[string]int // by kind/level
	Effects                              []string       // spells on the member, as the rules word them
	PortraitHeadID, PortraitBodyID       string         // gobox's pool picks, stored back as given
	IconHeadID, IconBodyID               string
	IconColours                          []byte
	NPC, Okay, Foe                       bool
}

// The states. Each situation is one explicit type; similar screens share one.

// Assembling: the party is put together before play, and characters are
// made (Draft set) or changed.
type Assembling struct {
	Common
	Roster []string // saved characters
	Draft  *Draft   // a character being made or changed; nil none
}

// Draft is a character being made or changed: the member so far, the
// fields still to choose and their options.
type Draft struct {
	Member
	Fields       []game.Field
	MinHP, MaxHP int
	Minimums     string // the rules' text on stat minimums for the class
}

// Exploring: the party walks a map (dungeon, town or wilderness).
type Exploring struct {
	Common
	Map ExploreMap
}

// ExploreMap is what gobox needs to draw the 3D view and the automap.
// Whether a step goes through is the rules' call (Do(Step)).
type ExploreMap struct {
	Walls   Walls
	Automap game.Automap
}

// Walls is the map's walls by square and side and what they look like.
type Walls struct {
	Side                [][][4]int     // a square's wall value by side (N E S W), rows from the top (0: open)
	Fronts              map[int]string // a wall value's art name in Data.Art
	Pool                map[int]string // a wall value's remastered texture, in place of Fronts
	Floor, Sky          string
	FloorRGBA, CeilRGBA *color.RGBA
	Backdrop            *game.Picture
	Outside             bool    // under the sky: gobox draws sun and mountains by Where's hour
	SunAz, SunEl        float64 // the sun's bearing and height
}

// Fighting: a battle is on.
type Fighting struct {
	Common
	Field Battlefield
}

// Battlefield is the battle as the rules hold it.
type Battlefield struct {
	Size      image.Point
	Squares   []game.Ground
	Tiles     []string // Ground.Tile's Data.Art names
	FloorTile string
	Floor     color.RGBA
	Units     []Unit
	Clouds    []game.Effect // lasting spell effects on squares
	Round     int
	Turn      int // the unit ID whose turn it is
	Dungeon   bool
}

// Unit is a combatant.
type Unit struct {
	ID, Member      int // Member: party index, -1 a monster
	Name            string
	At              image.Point
	Squares         []image.Point
	Size, Facing    int
	Foe, Dead, Gone bool
	HP, MaxHP, AC   int
	Status, Dmg     string
	Hit             int // the turn's unit's chance in percent to hit it, 0 not a target
	Art             game.Picture
}

// SceneKind says what a Scene is.
type SceneKind int

const (
	SceneText    SceneKind = iota // the script's text and pictures
	SceneTalk                     // someone speaks; Replies are the answers
	SceneHandout                  // a journal handout to read; Continue leaves
)

// Scene: a script shows something and waits for the player. The script's
// place is the game's state; Continue or a Reply moves it on.
type Scene struct {
	Common
	Kind     SceneKind
	Speaker  string
	Pictures []game.Picture
	Text     []string // plain paragraphs; colours are gobox's
	Replies  []string // SceneTalk: the answers, Reply{N}
	Typed    bool     // the reply is typed text (Reply{Text})
	Handout  string   // SceneHandout: the title
	Lines    []string // SceneHandout: journal.Layout's
}

// Trading: a shop, temple, trainer or vault is open.
type Trading struct {
	Common
	Name     string
	Goods    []game.Item
	Prices   []int
	Buys     bool
	Offers   []game.Offer // the appraisals of gems and jewels
	Services []game.Service
	Training *game.Training
	Pooled   map[string]int
}

// Looting: treasure lies before the party.
type Looting struct {
	Common
	Items []game.Item
	Money map[string]int
}

// Camping: the party has made camp (rest, memorize, scribe, fix, order).
type Camping struct{ Common }

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

// Event is one fact in an Outcome. The set is closed like State's.
type Event interface{ event() }

type (
	Narrated    struct{ Text string }          // a line of the rules' text
	Said        struct{ Who, Text string }     // someone speaks
	Shown       struct{ Picture game.Picture } // a picture comes up
	Sounded     struct{ Sound string }         // a sound plays
	MusicChange struct{ Song string }          // the music changes
	Walked      struct {
		Unit int
		Path []game.Step
	} // a unit walks (combat)
	Turned   struct{ Unit, Facing int }
	Attacked struct {
		From, To int
		Hit      bool
		Damage   int
	}
	Shot struct {
		From int
		To   image.Point
		Art  game.Picture
	}
	CastSpell struct {
		From, Spell int
		Ats         []image.Point
	}
	Struck struct{ Unit, Damage int } // harm without an attack (a spell, a trap)
	Healed struct{ Unit, HP int }
	Fell   struct{ Unit int } // down or dead
	Fled   struct{ Unit int }
	Moved  struct {
		To     image.Point
		Facing int
	} // the party, exploring
	TimePassed struct{ D time.Duration }
	Gained     struct {
		Member int
		What   string
	} // an item, money, experience, a level
	Lost struct {
		Member int
		What   string
	}
	JournalNote struct{ ID int }     // a Lore entry was added
	Refused     struct{ Why string } // a step blocked by the rules (a wall, a locked door)
	Ended       struct{ Won bool }
)

func (Narrated) event()    {}
func (Said) event()        {}
func (Shown) event()       {}
func (Sounded) event()     {}
func (MusicChange) event() {}
func (Walked) event()      {}
func (Turned) event()      {}
func (Attacked) event()    {}
func (Shot) event()        {}
func (CastSpell) event()   {}
func (Struck) event()      {}
func (Healed) event()      {}
func (Fell) event()        {}
func (Fled) event()        {}
func (Moved) event()       {}
func (TimePassed) event()  {}
func (Gained) event()      {}
func (Lost) event()        {}
func (JournalNote) event() {}
func (Refused) event()     {}
func (Ended) event()       {}

// The actions. Every game uses these; a game-specific verb gets its own
// type here. Which apply now is the state's Allowed.
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
	Shoot      struct{ Ats []image.Point }
	DoManeuver struct {
		ID  int
		Ats []image.Point
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
