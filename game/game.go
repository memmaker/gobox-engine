// Package game is the UI-free contract between a Gold Box game and gobox.
// A game states domain facts (State), asks domain questions (Ask), accepts
// rule verbs (Decision) and reports Events. It never says how anything is
// drawn: no colours, cells, rects, inks or positions. Modelled on pod's
// rules.State / Decision / Apply / Pending / Event. Stdlib only.
package game

import "image"

// Rules is a game's rules engine. It never blocks and starts no goroutines.
type Rules interface {
	State() State          // a copy; changing it changes nothing
	Apply(Decision) Result // take what the player decided, run on until it waits
	Pending() *Ask         // the question the game waits on, nil if none
	Step() Result          // advance time with no decision (animations, AI, scripts)
}

// Result is what one Apply or Step did. Err refuses the decision; nothing changed.
type Result struct {
	Events []Event
	Err    error
}

// Refusal is an Err that carries the reason and, if any, the question to put instead.
type Refusal struct {
	Text string
	Ask  *Ask
}

func (r *Refusal) Error() string { return r.Text }

// Picture is an art ref: a DAX File/Block/Frame (Block -1: none), or a Name,
// plus optionally raw indexed pixels (Pixels are palette indices, W*H long).
type Picture struct {
	File         string
	Block, Frame int
	Name         string
	W, H         int
	Pixels       []byte // palette indices, W*H
	Clear        int    // the index drawn transparent, -1 none
}

// Mode says what the game waits for.
type Mode int

const (
	ModeNone Mode = iota
	ModeTitle
	ModeExplore
	ModeAsk
	ModeParty
	ModeCreate
	ModeCamp
	ModeShop
	ModeTemple
	ModeTraining
	ModeVault
	ModeLoot
	ModeCombat
	ModeStory
)

// State is the domain superset across the games. Fields a game lacks stay zero.
type State struct {
	Mode       Mode
	Pending    *Ask
	Party      []Member
	Selected   int // index into Party
	Where      Where
	Picture    Picture // story / view picture
	Story      []Picture
	LastText   string
	Purse      map[string]int // money by kind ("gold", "platinum", ...)
	Quests     []Quest
	Map        Automap
	Combat     *Combat
	Shop       *Shop
	Temple     []Service
	TempleName string
	Training   *Training
	Loot       *Loot
	Door       *Door
	Vault      *Vault
	Saves      []SaveSlot
	Creation   *Creation
}

// Where is the party's place and time.
type Where struct {
	Area      string
	X, Y      int
	Facing    int // 0 N, 2 E, 4 S, 6 W
	Inside    bool
	Day       int
	Hour      int
	Minute    int // of the hour
	Month     int
	Year      int
	Mode      string // travel mode: "dungeon", "wilderness", "town", ...
	Searching bool
	Detail    map[string]int // game-specific domain numbers (block ids, wall sets)
}

// Member is one character.
type Member struct {
	Name, Race, Class, Alignment, Status string
	Levels                               []int
	Stats                                map[string]int // "str", "int", ...
	HP, MaxHP, AC, THAC0, Move           int
	Exp                                  uint32
	Items                                []Item
	Spells                               []SpellRef
	Portrait                             Picture
	Icon                                 map[string]int // icon designer choices as numbers
	BaseStats                            map[string]int // stats before magic and age
	IconColors                           []byte         // icon palette indices
	ClassText, LevelText                 string         // e.g. "old/new", "3/3"
	Slots                                map[string]int // spell slots left by kind/level
	Okay                                 bool
}

// SpellRef is a spell a member knows or has memorized.
type SpellRef struct {
	ID        int
	Name      string
	Memorized bool
}

// Item is one item.
type Item struct {
	Name                        string
	Readied, Identified, Cursed bool
	Charges, Price, Weight      int
}

// Quest is a quest fact.
type Quest struct {
	ID, Text string
	Done     bool
}

// Automap is what the party knows of the map.
type Automap struct {
	Area   string
	Size   image.Point
	Known  []image.Point // squares seen
	Marked []image.Point
	Walls  []Wall // by square
	Doors  []Wall
}

// Wall is a wall or door on one side of a square. Side as Facing; Kind is the game's wall/door kind.
type Wall struct {
	At         image.Point
	Side, Kind int
}

// Combat is a fight. Squares hold art refs, not cells.
type Combat struct {
	Size    image.Point
	Squares []Ground
	Units   []Unit
	Turn    *Turn
	Round   int
	Dungeon bool
	Reach   []Reach    // squares the turn's unit may walk to
	Spells  []SpellRef // castable by the turn's unit
	Aims    []SpellRef // spells that take a square
}

// Reach is a square the unit can walk to and the steps it costs.
type Reach struct {
	At    image.Point
	Steps int
}

// Ground is one battlefield square.
type Ground struct {
	At    image.Point
	Art   Picture
	Block bool
}

// Unit is a combatant.
type Unit struct {
	ID         int
	Name       string
	At         image.Point
	Facing     int
	Foe        bool
	HP, MaxHP  int
	Art        Picture
	Dead, Gone bool
	Size       int           // footprint edge in squares
	Squares    []image.Point // squares covered
	AC         int
	Status     string
}

// Turn is whose turn it is and what they may do.
type Turn struct {
	Unit    int
	Moves   int
	Options []Option
}

// Option is a rule verb the player may pick, with its original hotkey.
type Option struct {
	Label string
	Key   rune
	Do    Decision
}

// Shop, Service, Training, Loot, Door, Vault: the place an area script has
// opened (Mode says which); nil when none is open.
type Shop struct {
	Name   string
	Goods  []Item
	Prices []int
	Buys   bool // the shop buys from the party
}

type Service struct {
	Name  string
	Price int
}

type Training struct {
	Ready  []int // party indices able to level
	Cost   int
	Trains map[int]string // party index to what it trains to
}

type Loot struct {
	Items []Item
	Money map[string]int
}

type Door struct {
	Locked, Stuck, Secret bool
}

// Vault is the party's storage an area script has opened, like a shop.
type Vault struct {
	Items []Item
	Money map[string]int
}

// Creation is the character creation draft.
type Creation struct {
	Draft        Member
	Step         string // "race", "class", "stats", "name", ...
	Choices      []string
	Roster       []string // saved characters that can be added
	MinHP, MaxHP int
}

// SaveSlot is a save slot: its name and info text.
type SaveSlot struct {
	Name, Info string
}

// AskKind says what an Ask wants.
type AskKind int

const (
	YesNo AskKind = iota
	PickMember
	PickSpell
	PickItem
	Choice
	Number
	Name
	Square
	Direction
	Continue
)

// Ask is a domain question. The answer is an Answer decision.
type Ask struct {
	Kind    AskKind
	Prompt  string
	Text    string   // the question's body text
	Options []Option // Choice: labels with the original hotkeys
	Default int      // default option or number
	Max     int      // Name length, Number upper bound
}

// Decision is a typed rule verb. Screen verbs are not Decisions.
type Decision interface{ decision() }

// Answer answers the pending Ask, the same way in every game, by its Kind:
//   - Choice: N is the index into Ask.Options
//   - YesNo: N is 1 for yes, 0 for no
//   - PickMember: N is the party index; PickItem: the item's index in its list;
//     PickSpell: the spell id
//   - Number: N; Name: Text
//   - Square: At
//   - Direction: N is the facing (0 north, 2 east, 4 south, 6 west; odd: diagonals)
//   - Continue: no fields
//
// Backing out of a question is Cancel, never a special N.
type Answer struct {
	N    int
	Text string
	At   image.Point
}

// Cancel backs out of the pending Ask (the original's Esc / Exit).
type Cancel struct{}

// Adjust raises or lowers one of the creation draft's stats by By (the
// original's stat modify step); Stat indexes Member.Stats.
type Adjust struct{ Stat, By int }

// Money kinds: Purse, Loot.Money and Vault.Money use these names, the same in
// every game.
const (
	Copper   = "copper"
	Silver   = "silver"
	Electrum = "electrum"
	Gold     = "gold"
	Platinum = "platinum"
	Gems     = "gems"
	Jewelry  = "jewelry"
	Credits  = "credits"
)

type (
	Forward struct{}
	TurnBy  struct{ By int } // 2 right, 6 left, 4 about
	Search  struct{}
	Look    struct{}
	Select  struct{ N int }
	Move    struct{ To image.Point } // combat
	Attack  struct{ Target int }
	Cast    struct {
		Spell, Target int
		At            image.Point
	}
	UseItem  struct{ Member, Item, Target int }
	Ready    struct{ Member, Item int }
	Drop     struct{ Member, Item int }
	Trade    struct{ From, To, Item int }
	Buy      struct{ Item int }
	Sell     struct{ Member, Item int }
	TakeLoot struct{ Item int }    // Item -1: all
	Rest     struct{ Minutes int } // 0: the default rest
	Camp     struct{}
	Train    struct{ Member int }
	Heal     struct{ Service, Member int }
	Create   struct {
		Field string
		N     int
		Text  string
	}
	AddChar  struct{ Roster int }
	DropChar struct{ Member int }
	Begin    struct{}
	Wait     struct{} // combat: end turn / hold
	Flee     struct{}
	Save     struct{ Slot int }
	Load     struct{ Slot int }

	// Camp.
	Fix       struct{ Member int }
	Memorize  struct{ Member, Spell int }
	Forget    struct{ Member, Spell int }
	Scribe    struct{ Member, Spell int }
	Order     struct{ From, To int }
	LeaveCamp struct{}
	// Combat.
	Guard      struct{}
	EndTurn    struct{}
	Quick      struct{}
	Bandage    struct{ Member int }
	TurnUndead struct{}
	Aim        struct{ At image.Point }
	// Items and money.
	Unready   struct{ Member, Item int }
	Halve     struct{ Member, Item int } // split a stack
	Join      struct{ Member, Item int } // merge stacks
	Identify  struct{ Member, Item int }
	Appraise  struct{ Member, Item int }
	Repair    struct{ Member, Item int }
	PoolCoins struct{ Member int }
	Leave     struct{} // leave the open shop, vault, loot or temple
	// Paladin.
	LayOnHands struct{ Member int }
	Cure       struct{ Member, Service int }
	// Party.
	Reroll           struct{}
	DiscardCharacter struct{}
	RemoveCharacter  struct{ Roster int }
	ChangeClass      struct{ Member, Class int }
	DropMember       struct{ Member int }
	SetSpeed         struct{ N int }
	SetDifficulty    struct{ N int }
	SetIcon          struct {
		Member int
		Field  string
		N      int
	}
	Call struct{ Member int }
)

func (Answer) decision()   {}
func (Cancel) decision()   {}
func (Adjust) decision()   {}
func (Forward) decision()  {}
func (TurnBy) decision()   {}
func (Search) decision()   {}
func (Look) decision()     {}
func (Select) decision()   {}
func (Move) decision()     {}
func (Attack) decision()   {}
func (Cast) decision()     {}
func (UseItem) decision()  {}
func (Ready) decision()    {}
func (Drop) decision()     {}
func (Trade) decision()    {}
func (Buy) decision()      {}
func (Sell) decision()     {}
func (TakeLoot) decision() {}
func (Rest) decision()     {}
func (Camp) decision()     {}
func (Train) decision()    {}
func (Heal) decision()     {}
func (Create) decision()   {}
func (AddChar) decision()  {}
func (DropChar) decision() {}
func (Begin) decision()    {}
func (Wait) decision()     {}
func (Flee) decision()     {}
func (Save) decision()     {}
func (Load) decision()     {}

func (Fix) decision()              {}
func (Memorize) decision()         {}
func (Forget) decision()           {}
func (Scribe) decision()           {}
func (Order) decision()            {}
func (LeaveCamp) decision()        {}
func (Guard) decision()            {}
func (EndTurn) decision()          {}
func (Quick) decision()            {}
func (Bandage) decision()          {}
func (TurnUndead) decision()       {}
func (Aim) decision()              {}
func (Unready) decision()          {}
func (Halve) decision()            {}
func (Join) decision()             {}
func (Identify) decision()         {}
func (Appraise) decision()         {}
func (Repair) decision()           {}
func (PoolCoins) decision()        {}
func (Leave) decision()            {}
func (LayOnHands) decision()       {}
func (Cure) decision()             {}
func (Reroll) decision()           {}
func (DiscardCharacter) decision() {}
func (RemoveCharacter) decision()  {}
func (ChangeClass) decision()      {}
func (DropMember) decision()       {}
func (SetSpeed) decision()         {}
func (SetDifficulty) decision()    {}
func (SetIcon) decision()          {}
func (Call) decision()             {}

// EventKind says what an Event reports.
type EventKind int

const (
	EvMessage EventKind = iota
	EvText
	EvSound
	EvMusic
	EvPicture
	EvPortrait
	EvQuit
	EvAttack
	EvHit
	EvMiss
	EvDamage
	EvDown
	EvSpell
	EvMove
	EvMissile
	EvSaid    // Text spoken by Role
	EvSong    // Name starts playing
	EvRefused // Text says why
	EvReset   // the state changed wholesale (load, new game): redraw from State
)

// Event is something that happened.
type Event struct {
	Kind         EventKind
	Text         string
	Role         string // EvSaid: "narrator", an NPC name, ...
	N            int
	Wait         bool
	Clear        bool
	Picture      Picture
	Sound        string // sound ref for EvSound / song ref for EvSong
	Unit, Target int
	At, To       image.Point
	Sprite       []Picture
}

// Data is declarative overrides only; no drawing, no positions.
type Data struct {
	Series     string
	ID         string // game id
	Texts      map[string]string
	Credits    []string
	SpellRules map[string]int
	Art        map[string]Picture
	Sounds     map[string]string
	Palette    []byte   // palette bytes
	Commands   []Option // extra party commands: Decisions with labels
}
