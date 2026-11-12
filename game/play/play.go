// Package play is the typed game contract : what a game offers gobox.
//
// The game is purely reactive. It never does anything by itself: gobox reads
// State, switches over its type to pick the screen and controls, turns the
// player's input into one of the Actions that state publishes, and hands it
// to Do. Do refuses an action it does not allow, or applies it and returns
// the Outcome, which gobox plays (animates, voices, paces) before it reads
// State again. Questions, menus, waits, paging and screens are all gobox's.
package play

import (
	"gobox/core"
	"image"
	"image/color"
	"time"

	"gobox/game"
)

// Game is what every game implements.
type Game interface {
	State() State               // what is true now; a copy
	Do(Action) (Outcome, error) // the only way anything changes; error: refused, nothing changed
	Save() ([]byte, error)      // the whole game, for a slot file gobox keeps
	Load([]byte) error          // back to what Save gave; error: refused, nothing changed
}

// Slots is what a game with save slots of its own (the original's) may
// also implement: gobox lists them after its own and loads one by index.
type Slots interface {
	Slots() []game.SaveSlot
	LoadSlot(n int) error
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
// exactly these; Do checks again. An allowed AdvanceTimeBy{D} allows any
// duration up to D (D zero: any), so gobox may rest in steps.
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
	Adventure bool      // the party is out in the world (the game's own rule); gobox takes it as given
	Portrait  *Portrait // a speaker's portrait up (a shopkeeper), nil none; gobox composes it
	NoSave    bool      // saving is refused here (the rules say where the party may save)
	NoLoad    bool      // loading is refused here
	Party     []game.Member
	Selected  int
	Purse     map[string]int
	PartyCap  int // most items the party carries in all (a shared inventory's slots), 0 no limit
	PoolItems int // the party inventory's items held (of PartyCap)
	Where     game.Where
	Area      string      // the area's name as the player knows it
	Lore      []game.Lore // the journal, newest first
	Quests    []game.Quest
	Allowed   Allowed
	Status    string        // the rules' status line (plain text; "" none)
	Text      []string      // the rules' message lines in the text pane, oldest first
	Picture   *game.Picture // the picture or portrait up (in the view while exploring); nil none
	Layers    []game.Layer  // a picture built of layers (with frames), as Picture
	Day       bool          // daylight as the rules see it, in every state
	Automap   game.Automap  // the map of where the party is (the overland, the starmap, ...), any state
	Ask       *Question     // a question over this state (the view stays); Reply answers it

	// Overlays any state may carry: a question only the rules can put
	// (a power to choose), the grid picker and the icon designer the
	// rules fill from their art (nil: none up).
	PowerPick  *game.PowerPick  // PickPower answers
	Picker     *game.Picker     // Pick, ClosePicker answer
	IconEditor *game.IconEditor // IconSize, IconColour, IconRemove, IconPNG, EndIcon act on it
}

// Portrait is a portrait as the rules name it: the head and body pictures
// (HEAD2 3, BODY2 3), or gobox pool ids where the game has them.
type Portrait struct {
	Head, Body     game.Picture
	HeadID, BodyID string
}

// Question is what the rules ask over a state: the text and the answers
// (Reply{N}), or a typed answer (Reply{Text}) or number (Reply{N}).
type Question struct {
	Text         string
	Replies      []string
	YesNo        bool // Replies are no, yes: Reply{N: 1} is yes
	Typed        bool
	Number       bool
	Default, Max int         // Typed: the longest; Number: the default and the largest; YesNo: 1 starts on yes
	NoPick       bool        // no highlight, Enter and the arrows do not answer it; only the hotkeys (a YesNo: Y and N) do
	NoEsc        bool        // Esc does not answer it (the question stays up), as the original's ask that ignores Esc
	Continue     bool        // a key to go on (Continue answers it); Common.Status may say so
	Heading      bool        // a heading to pick (the overland): Heading{Dir}, Leave for Done
	List         bool        // Replies are a pick from a list (a roster, a town's places), not a command menu
	Keys         []rune      // the game's own hotkey per reply (0: from the label); nil: all from the labels
	Rows         []game.Line // List: per reply, its flags (a heading, dimmed, off, the spell it names); nil: plain rows
	Do           []Action    // per reply, an action gobox runs in place of Reply (a View's Gear: OpenItems{Member}); nil: Reply
	Esc          EscReply    // the reply Esc answers (the menu's Exit, as the originals'); zero: none
}

// EscReply names the reply Esc answers: EscTo(i) is Replies[i]. It is the
// index plus one so that the zero value, NoEscReply, means Esc answers no
// reply.
type EscReply int

// NoEscReply: Esc answers no reply.
const NoEscReply EscReply = 0

// EscTo is Esc answering Replies[i].
func EscTo(i int) EscReply { return EscReply(i + 1) }

// Reply is the index into Replies Esc answers; false for NoEscReply.
func (e EscReply) Reply() (int, bool) { return int(e) - 1, e != NoEscReply }

// The states. Each situation is one explicit type; similar screens share one.

// Assembling: the party is put together before play, and characters are
// made (Draft set) or changed. gobox builds the party menu from Allowed:
// one action of a kind is its entry. AddCharacter{Roster: -1} is the Add
// Character entry while the roster may be empty: gobox opens the roster
// and sends AddCharacter with the one picked.
type Assembling struct {
	Commands []Option // the game's own party commands (Do, with a label)
	Common
	Roster []string // saved characters
	Draft  *Draft   // a character being made or changed; nil none
}

// Draft is a character being made or changed: the member so far, the
// fields still to choose and their options.
type Draft struct {
	game.Member
	Fields       []game.Field
	MinHP, MaxHP int
	Minimums     map[string]int // the class's lowest ability scores, by Member.Stats key ("str", ...)
	Step         string         // the creation step the rules are on ("race", "skills", ...), "" any
	Ask          *Question      // that step's question (its options, a typed name, a number, a yes/no), answered by SetChoice{Field: Step, N or Text}; nil: the sheet's fields
	Stats        []game.StatDef // the sheet's abilities in the game's order and count (nil: STR .. CHA from Member.Stats)
	Derived      []game.Derived // the sheet's figures besides the abilities, in the game's order (nil: from Member)
}

// Exploring: the party walks a map (dungeon, town or wilderness).
type Exploring struct {
	Common
	Map  ExploreMap
	Day  bool       // daylight as the rules see it (the sky they paint)
	Door *game.Door // a door the party stands at, with the ways through (TryDoor); nil none
}

// ExploreMap is what gobox needs to walk, draw the 3D view and the
// automap. gobox walks: it asks Walk whether a side lets a step through and
// where a script waits, then sends RunEvent{Before}, Arrive and RunEvent.
type ExploreMap struct {
	Walk     game.Map // the game's answers about its map
	Overland bool     // the wilderness or city map, not the 3D view
	Walls    Walls
	Automap  game.Automap
}

// Walls is the map's walls by square and side and what they look like.
type Walls struct {
	Side                [][][4]int     // a square's wall value by side (N E S W), rows from the top (0: open)
	Fronts              map[int]string // a wall value's art name in Data.Art
	Pool                map[int]string // a wall value's remastered texture, in place of Fronts
	Floor, Sky          string
	FloorRGBA, CeilRGBA *color.RGBA
	Backdrop            *game.Picture
	Outside             bool    // under the sky
	Sun                 bool    // the sun is up in the sky the rules paint
	Mountains           bool    // the mountain range on the horizon, by day outside
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
	Units     []game.Unit
	Clouds    []game.Effect // lasting spell effects on squares
	Round     int
	Turn      int // the unit ID whose turn it is
	Dungeon   bool
	Moves     int                 // the turn's unit's moves left
	Casting   int                 // the spell the turn's unit casts over rounds, 0 none
	Weapon    *game.Weapon        // the turn's weapon in hand (its icon on Attack), nil none
	Missile   *game.Weapon        // its ranged weapon readied, nil none
	Maneuvers []game.TurnManeuver // the rule set's maneuvers, by DoManeuver ID
}

// SceneKind says what a Scene is.
type SceneKind int

const (
	SceneText     SceneKind = iota // the script's text and pictures
	SceneTalk                      // someone speaks; Common.Ask holds the answers
	SceneHandout                   // a journal handout to read; Continue leaves
	SceneTitle                     // the title (Intro); Continue ends it
	SceneSequence                  // a Data.Sequences entry full screen; Continue at its end
	SceneResults                   // the after-combat page (Results); Continue leaves
	SceneScreen                    // a typed screen (Screen: a ship sheet, a table, ...)
	SceneHeading                   // a heading to pick (overland, rocket): Heading, Leave for Done
)

// Placed is a line of text at a text cell; Tone says what it means, gobox
// picks the colour.
type Placed struct {
	At   image.Point // column, row from 1
	Text string
	Tone game.Tone
}

// Scene: a script shows something and waits for the player. The script's
// place is the game's state; Continue or a Reply moves it on.
type Scene struct {
	Common
	Kind     SceneKind
	Speaker  string
	Pictures []game.Picture
	Text     []string          // plain paragraphs; colours are gobox's
	Handout  string            // SceneHandout: the title
	Lines    []string          // SceneHandout: journal.Layout's
	Layers   []game.Layer      // pictures with their place and frames (Pictures: gobox places them)
	Placed   []Placed          // text at cells over the window (Text: the message lines)
	Intro    *game.Intro       // SceneTitle
	Sequence string            // SceneSequence: the Data.Sequences name
	Results  *game.Results     // SceneResults
	Screen   *game.Screen      // SceneScreen
	Full     bool              // the pictures are one whole-screen picture
	Steps    map[int]TitleStep // SceneTitle: Intro.Pictures' step N's pictures and tune; gobox steps through them
	Walls    *Walls            // the 3D view the scene is over (nil: the one last explored, if any)
}

// TitleStep is one title picture step: its layers and the tune it starts.
type TitleStep struct {
	Layers []game.Layer
	Song   string // the song's name, or its number as text ("3": the game's song 3, as EvSong N)
}

// DoorWay is a way to try a door.
type DoorWay int

const (
	DoorExit  DoorWay = iota // leave it
	DoorBash                 // force it
	DoorPick                 // pick the lock
	DoorKnock                // cast Knock
)

// TradeKind says which place Trading is.
type TradeKind int

const (
	TradeShop TradeKind = iota
	TradeTemple
	TradeTraining
	TradeVault
)

// Trading: a shop, temple, trainer or vault is open.
type Trading struct {
	Common
	Kind     TradeKind
	Name     string
	Goods    []game.Item
	Buys     bool
	Offers   []game.Offer // the appraisals of gems and jewels
	Services []game.Service
	Training *game.Training
	Pooled   map[string]int
	Vault    *game.Vault    // TradeVault: what is stored
	Identify int            // the price to identify, 0 none
	Unknown  []game.ItemRef // the party's items the shop can identify
	Known    *game.ItemRef  // the item last identified here, nil none
}

// Looting: treasure lies before the party.
type Looting struct {
	Common
	Items  []game.Item
	Money  map[string]int
	Status bool // LootStatus is offered (the ship sheet over the fight)
}

// Camping: the party has made camp (rest, memorize, scribe, fix, order).
// Study is the rest the party needs to finish studying its spells (0:
// none); gobox's default rest is this long when set.
type Camping struct {
	Common
	Study time.Duration
	Heals bool // a rest would heal someone (the rest's question asks only then)
}

// Over: the game has ended.
type Over struct {
	Common
	Won  bool
	Text string // the rules' last words
}

// ShipFight: a ship battle is on. The crew menu up is Menu
// (its Replies and Keys), answered with Crew{N}; Sheet a ship's record
// shown over the fight.
type ShipFight struct {
	Common
	Fight game.SpaceFight
	Sheet *game.ShipSheet
	Menu  *Question // nil: no crew menu up
}

func (Assembling) state() {}
func (Exploring) state()  {}
func (Camping) state()    {}
func (Fighting) state()   {}
func (Scene) state()      {}
func (Trading) state()    {}
func (Looting) state()    {}
func (Over) state()       {}
func (ShipFight) state()  {}

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
		From   int
		To     image.Point
		Art    game.Picture
		StepMs int // the flight's time per third of a square, the original's (0: gobox's)
	}
	CastSpell struct {
		From, Spell int
		Ats         []image.Point
		StepMs      int // a flight's time per third of a square (0: gobox's)
	}
	Engaged struct{ At Fighting }      // a fight began: it before the first turn, board and party (the step's fight events play from it)
	Struck  struct{ Unit, Damage int } // harm without an attack (a spell, a trap)
	Healed  struct{ Unit, HP int }
	Fell    struct {
		Unit int
		Body game.Picture // what it leaves on its square, as game.Event.Picture on EvDown (empty: nothing)
		Ms   int          // the replay holds this long on the fall (0: it does not stop)
	} // down or dead
	Fled  struct{ Unit int }
	Moved struct {
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
	SaveOffered struct{} // a script offers to save the game
	ShipShown   struct {
		Fight game.SpaceFight // a moment of the ship fight (a blink, a blast frame)
	}
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
func (Engaged) event()     {}
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
func (SaveOffered) event() {}
func (ShipShown) event()   {}

// The actions. Every game uses these; a game-specific verb gets its own
// type here. Which apply now is the state's Allowed.
type (
	// Anywhere a member's goods are shown.
	Ready struct {
		Member int
		Item   core.ItemHandle
		Count  int // of a stack (0: all)
	}
	Unready struct {
		Member int
		Item   core.ItemHandle
	}
	Drop struct {
		Member int
		Item   core.ItemHandle
	}
	Give struct {
		From, To int
		Item     core.ItemHandle
	}
	Halve struct {
		Member int
		Item   core.ItemHandle
	}
	Join struct {
		Member int
		Item   core.ItemHandle
	}
	UseItem struct { // Item NoItem in a fight's Allowed: the pack is open to the turn (gobox offers Items) though nothing in it is usable
		Member int
		Item   core.ItemHandle
		Target int
		Ats    []image.Point
		Spell  int // a scroll's spell, 0 the first
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
	Reroll           struct{ Minimal bool } // Minimal: the lowest roll the class allows
	KeepCharacter    struct{}
	DiscardCharacter struct{}
	ChangeClass      struct{ Member, Class int }
	SetDifficulty    struct{ N int }
	Begin            struct{}
	QuickStart       struct{}
	Quit             struct{}
	Modify           struct{ Member int } // change a member (Draft set)
	Rename           struct{ Name string }
	// The picker and the icon designer (Common's overlays).
	OpenPicker  struct{ Kind int }
	Pick        struct{ N int }
	ClosePicker struct{}
	EditIcon    struct{ Member int }
	IconSize    struct{}
	IconColour  struct{ Part, Shade, Value int }
	IconRemove  struct{}
	IconPNG     struct{ Path string }
	EndIcon     struct{ Keep bool }
	// The sheet.
	UsePower struct {
		Member        int
		Kind          game.PowerKind
		Spell, Target int
	}
	PickPower struct{ ID int }
	SetPolicy struct{ ID, Policy int }

	// Exploring.
	Arrive struct {
		From, To image.Point
		Facing   int
	} // the party steps from From to To; the game may refuse (a locked door) and runs no event
	RunEvent struct {
		Before bool // the script ahead of a step through Side, before the party moves
		Side   int
	} // run the square's event (Walk.Event, Walk.Departure said one waits)
	Face    struct{ Facing int }
	TryDoor struct{ Way DoorWay } // a way through Exploring.Door
	Mark    struct {
		At   image.Point
		Text string
	} // a note on the automap
	Unmark        struct{ At image.Point }
	Search        struct{}
	Look          struct{}
	MakeCamp      struct{}
	Call          struct{ Member int }
	AdvanceTimeBy struct{ D time.Duration } // also in Camping: rest; the outcome tells what happened meanwhile

	// Camping.
	StopRest   struct{} // the player stops a rest gobox paces (listed: the game hears of it, e.g. to end its own rest)
	Memorize   struct{ Member, Spell int }
	Forget     struct{ Member, Spell int }
	Scribe     struct{ Member, Spell int }
	Fix        struct{ Member int }
	Bandage    struct{ Member int }
	LayOnHands struct{ Member int }
	BreakCamp  struct{}
	Plan       struct{ Member, Spell, Count int } // copies to memorize at the next rest
	CampRest   struct{ Kind int }                 // one of Data.Rests

	// Fighting. EndTurn's outcome carries the monsters' moves up to the
	// next member's turn.
	MoveTo struct{ To image.Point }
	Attack struct {
		Target int
		At     *image.Point // aimed (Allowed.Aimed, squares): the square picked, nil by Target
	}
	Shoot      struct{ Ats []image.Point }
	DoManeuver struct {
		ID  int
		Ats []image.Point
	}
	TurnUndead struct{}
	Guard      struct{}
	Delay      struct{}
	Quick      struct{ All bool }
	// Manual takes every member back from the computer while it fights for them (the Gold Box's Space).
	Manual struct{}
	// QuickMagic switches whether the computer casts spells for the members it fights for (the Gold Box's Alt-M).
	QuickMagic struct{}
	Flee       struct{}
	EndTurn    struct{}
	// Turn verbs the rules answer by hotkey (classic answerVerbs).
	Target     struct{} // pick a target to fire at (the rules ask on)
	ViewTarget struct{} // look over the targets
	// ViewMember views member Member over a pick that lets the player look
	// first (the rules' own View menu; the question stays and comes back).
	// Allowed offers one per member while it may.
	ViewMember struct{ Member int }
	Intimidate struct{}
	Aid        struct{} // help a downed member next to the unit
	Sprint     struct{}
	Dodge      struct{}
	Jetpack    struct{}

	// Scene.
	Continue struct{}
	Heading  struct{ Dir int } // SceneHeading: 0 N, 2 E, ... odd diagonals

	// ShipFight.
	Crew  struct{ N int } // the crew menu's N
	Reply struct {
		N    int
		Text string
	} // Text: a typed answer

	// Trading.
	Buy struct {
		Member int
		Item   core.ItemHandle
	}
	Sell struct {
		Member int
		Item   core.ItemHandle
	}
	Identify struct {
		Member int
		Item   core.ItemHandle
	}
	Appraise struct {
		Member int
		Offer  game.OfferKind
		Keep   bool
	}
	Repair struct {
		Member int
		Item   core.ItemHandle
	}
	Reload struct {
		Member int
		Item   core.ItemHandle
	}
	Train         struct{ Member int }
	Heal          struct{ Service, Member int }
	PoolCoins     struct{ Member int }
	OpenTraining  struct{ Member int }
	CloseTraining struct{}
	Learn         struct{ Spell int }

	// Looting, and Trading where the shop holds pooled coins (Take, LootShare).
	Take struct {
		Item core.ItemHandle // NoItem: the money only
		All  bool            // the money and every item
	}

	LootDetect struct{}
	LootShare  struct{} // divide the money among the party
	LootStatus struct{}

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
func (Arrive) action()           {}
func (RunEvent) action()         {}
func (Face) action()             {}
func (Search) action()           {}
func (Look) action()             {}
func (MakeCamp) action()         {}
func (Call) action()             {}
func (AdvanceTimeBy) action()    {}
func (StopRest) action()         {}
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
func (ViewMember) action()       {}
func (Quick) action()            {}
func (Manual) action()           {}
func (QuickMagic) action()       {}
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
func (QuickStart) action()       {}
func (Quit) action()             {}
func (Modify) action()           {}
func (Rename) action()           {}
func (OpenPicker) action()       {}
func (Pick) action()             {}
func (ClosePicker) action()      {}
func (EditIcon) action()         {}
func (IconSize) action()         {}
func (IconColour) action()       {}
func (IconRemove) action()       {}
func (IconPNG) action()          {}
func (EndIcon) action()          {}
func (UsePower) action()         {}
func (PickPower) action()        {}
func (SetPolicy) action()        {}
func (TryDoor) action()          {}
func (Mark) action()             {}
func (Unmark) action()           {}
func (Plan) action()             {}
func (CampRest) action()         {}
func (Heading) action()          {}
func (Crew) action()             {}
func (OpenTraining) action()     {}
func (CloseTraining) action()    {}
func (Learn) action()            {}
func (LootDetect) action()       {}
func (LootShare) action()        {}
func (LootStatus) action()       {}
func (Target) action()           {}
func (ViewTarget) action()       {}
func (Intimidate) action()       {}
func (Aid) action()              {}
func (Sprint) action()           {}
func (Dodge) action()            {}
func (Jetpack) action()          {}

// AppraiseActions are what the offers allow: valuing each one, and with one
// valued, selling it and (room for it) keeping it.
func AppraiseActions(offers []game.Offer) []Action {
	var l []Action
	for _, o := range offers {
		l = append(l, Appraise{Member: o.Member, Offer: o.Offer})
		if o.Value > 0 && o.CanKeep {
			l = append(l, Appraise{Member: o.Member, Offer: o.Offer, Keep: true})
		}
	}
	return l
}

// CommonOf is the Common of any State (none: the zero Common).
func CommonOf(ps State) Common {
	switch t := ps.(type) {
	case Assembling:
		return t.Common
	case Exploring:
		return t.Common
	case Camping:
		return t.Common
	case Fighting:
		return t.Common
	case Scene:
		return t.Common
	case Trading:
		return t.Common
	case Looting:
		return t.Common
	case Over:
		return t.Common
	case ShipFight:
		return t.Common
	}
	return Common{}
}

// ByPosition gives the items of s (the party's, the goods, the vault's, the
// loot) the handles of their places in their lists (game.ByPosition).
// ponytail: the play boundary of the games whose items have no handle of
// their own (cotab, sotsb, pod, buckrogers); it goes when they switch to pool.
func ByPosition(s State) State {
	for _, m := range CommonOf(s).Party {
		game.ByPosition(m.Items)
	}
	switch t := s.(type) {
	case Trading:
		game.ByPosition(t.Goods)
		if t.Vault != nil {
			game.ByPosition(t.Vault.Items)
		}
	case Looting:
		game.ByPosition(t.Items)
	}
	return s
}
