// Package game is the UI-free contract between a Gold Box game and gobox.
// A game states domain facts (State), asks domain questions (Ask), accepts
// rule verbs (Decision) and reports Events. It never says how anything is
// drawn: no colours, cells, rects, inks or positions. Modelled on pod's
// rules.State / Decision / Apply / Pending / Event. Stdlib only.
package game

import (
	"image"
	"image/color"
)

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
	Clear        int    // with Pixels: the index drawn transparent, -1 none (ignored without Pixels)
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
	// Allowed lists the item verbs the rules take now, by Decision type
	// name ("UseItem", "Sell", "Trade", "Halve", "Join", ...); gobox offers
	// only these.
	Allowed                map[string]bool
	RestHeals, RestStudies bool        // what resting now does
	TempleOffers           []Offer     // the temple\'s appraisals, as Shop.Offers
	PartyCap               int         // most items the party carries in all, 0 no limit
	Bestiary               []Beast     // the monsters the party has met
	Picker                 *Picker     // a grid picker is up
	IconEditor             *IconEditor // the combat icon designer is up
	Page                   *Page       // the text page: the game's scripted text and story pictures
}

// Page is the text a game wrote and the pictures it showed: scripted text,
// story pages and the messages in the text window. Lines hold runs in the
// colours the game chose; the front lays them out.
type Page struct {
	Pane   [][]TextRun // the messages' lines (nil: no message shown)
	Prompt []TextRun   // the prompt row's text
	View   [][]TextRun // text over the view window, rows from 1
	Layers []Layer     // the pictures, bottom first
}

// TextRun is text in one colour on a line from column X (Upcase: the
// glyphs of a game that upcases what it writes).
type TextRun struct {
	X      int
	Text   string
	Fg, Bg color.RGBA
	Upcase bool
}

// Layer is a picture placed on the page: At in pixels of the page.
type Layer struct {
	Picture
	At image.Rectangle
}

// Picker is a grid picker up (the creation's heads and bodies, the icon's
// parts, the saved characters): every picture it holds, in order; the front
// lays them out in pages.
type Picker struct {
	Cell      image.Point   // a picture's size in text cells
	Cur       int           // the picture it opens on
	Pics      []Picture     // one per picture; Pixels nil where it has none (W, H in pixels, the DAX ref as loaded)
	Labels    []string      // the pictures' names (saved characters)
	Classes   []string      // ... their classes ("": the files are gone)
	Portraits []image.Image // ... their PNG portraits (nil: none; the DAX head is among Pics)
	Ask       string        // the question before a pick (%s: the label), "" none
}

// IconEditor is the combat icon designer: the icon built so far and the
// colours its six parts take. Colours name palette entries (Looks), in the
// same index space as Picture.Pixels.
type IconEditor struct {
	Large, Custom bool           // Custom: the poses are a dropped PNG (Icons), no parts or colours
	Poses         [2]Picture     // ready and action poses (Pixels nil: none)
	Icons         [2]image.Image // Custom: the PNG's poses
	Colour        [6][2]int      // each part's two colour numbers (0..15)
	Cells         [16]int        // the colour number in each palette cell, row-major (-1: empty)
	Looks         [16]byte       // the palette entry each colour number shows as
}

// Beast is a bestiary entry.
type Beast struct {
	Name, Text string
	Danger     int     // 1..8 (I..VIII), 0 unrated
	Art        Picture // the fight picture
	Icon       Picture // the combat icon
}

// ScrollSpell is one spell on a scroll.
type ScrollSpell struct {
	ID       int
	Scribing bool // being scribed now
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
	Hall      bool           // the party is where characters may join or leave it (a guild hall, an inn)
	Detail    map[string]int // game-specific domain numbers (block ids, wall sets)
}

// Member is one character.
type Member struct {
	Name, Race, Class, Alignment, Status string
	Levels                               []int
	Stats                                map[string]int // "str", "int", ...
	Str00                                int            // exceptional strength: 0 none, 1..100 for 18/01..18/00
	HP, MaxHP, AC, THAC0, Move           int
	Exp                                  uint32
	Items                                []Item // in the rules' order; gobox keeps it
	Spells                               []SpellRef
	Portrait                             Picture
	Icon                                 map[string]int // icon designer choices as numbers
	BaseStats                            map[string]int // stats before magic and age
	IconColors                           []byte         // icon palette indices
	ClassText, LevelText                 string         // e.g. "old/new", "3/3"
	Slots                                map[string]int // spell slots left by kind/level
	Okay                                 bool
	ItemCap                              int            // most items carried, 0 no limit
	SlotsMax                             map[string]int // spell slots in all, keys as Slots
	CanCast                              bool           // may cast now (status, no-magic area)
	Powers                               []Option       // the member\'s special powers menu (Decisions)
}

// SpellRef is a spell a member knows or has memorized.
type SpellRef struct {
	ID        int
	Name      string
	Level     int
	Memorized bool
	Planned   int    // copies to memorize at the next rest
	Group     string // the level group it is listed under, e.g. "cleric 1"
	Range     int    // aim: squares from the caster, 0 touch
	Targets   int    // aim: squares picked, 0 or 1 for one
	Area      int    // aim: radius hit around each square, 0 just that square
}

// Item is one item.
type Item struct {
	Name                        string
	Readied, Identified, Cursed bool
	Charges, Price, Weight      int
	Type                        int            // item type id (picks the icon)
	Kind                        string         // "weapon", "armor", "ring", ...
	Count                       int            // stack size, 0 or 1 for one
	Slot                        string         // readied in: weapon, shield, armor, helm, belt, robe, cloak, boots, gauntlets, ring, ring2, arrows, bolts; "" none
	Wrong                       bool           // the holder may not use it (class, race)
	Text                        string         // what Look tells of it
	Spell                       int            // a scroll's first spell id, 0 none (all of them: Scroll)
	Scroll                      []ScrollSpell  // a scroll's spells
	Scribe                      bool           // the holder may scribe the scroll's spell
	Usable                      bool           // UseItem takes it now
	Words                       []string       // the name\'s words as the original builds them
	Hidden                      string         // an unidentified item\'s real name, shown once identified
	Sale                        map[string]int // what a shop pays for it, by money kind
}

// Quest is a quest fact.
type Quest struct {
	ID, Text string
	Done     bool
}

// Automap is what the party knows of the map.
type Automap struct {
	Area      string
	Size      image.Point
	Known     []image.Point // squares seen; empty: nothing is known
	Marked    []image.Point
	Notes     []string      // the player's text for each Marked square
	NoteKinds []string      // a Label.Kind for each Marked square, \"\" none
	Walls     []Wall        // by square
	Doors     []Wall        // State says what each door is
	Walked    []image.Point // squares walked on (Known also holds those only seen)
	Labels    []Label       // the map's own labels; the player's are Marked/Notes
	// Terrain is the overland map's terrain id per square, row by row over
	// Size; its art is Data.Art["terrain/<id>"] (a game numbers each
	// remapped variant as its own id and gives its Pixels), the party marker
	// Data.Art["marker/<facing>"]. Known and Walked apply as on a dungeon
	// map. Empty for a dungeon map.
	Terrain []int
}

// Label is a named place on the automap.
type Label struct {
	At   image.Point
	Text string
	Kind string // "shop", "temple", "inn", "civic", ...
}

// Wall is a wall or door on one side of a square. Side as Facing; Kind is the game's wall/door kind.
type Wall struct {
	At         image.Point
	Side, Kind int
	State      DoorState // a door's state; walls leave it zero
}

// DoorState is what a door on the automap is.
type DoorState int

const (
	DoorClosed DoorState = iota
	DoorOpen
	DoorLocked
	DoorWizardLocked
)

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
	Aims    []SpellRef // spells that take a square, with their Range, Targets and Area
	Targets []int      // unit IDs the turn's Attack takes; while aiming, the foes Tab cycles through
	Panel   int        // the unit ID the status row shows (who acts, or is being hit), 0 none
	// Marks are the squares framed for a choice: with Area, a spell's area
	// (AreaAt the squares it covers round the aimed square, Skipped the allies
	// in it that it passes over), else the units an aim would hit.
	Marks, AreaAt, Skipped []image.Point
	Area                   bool
	Why                    string // aiming: why the aimed square cannot be taken, "" when it can
}

// Reach is a square the unit can walk to and the steps it costs.
type Reach struct {
	At    image.Point
	Steps int
}

// Ground is one battlefield square.
type Ground struct {
	At     image.Point
	Art    Picture
	Block  bool
	Height int // the square's height (a wall, a tree)
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
	Dmg        string // the unit's damage, e.g. "1d8+6"
	Member     int    // party index of a party unit, -1 for a monster
	Active     bool   // it is the unit the player acts for
	Hit        int    // the active unit's chance in percent to hit it with one blow, 0 not a target
}

// Turn is whose turn it is and what they may do.
type Turn struct {
	Unit    int
	Moves   int
	Casting int    // the spell the unit is casting over rounds, 0 none
	Aim     bool   // aiming: Aim takes a square (Cast: Ats), Cancel gives up
	Spell   string // aiming a spell, its name
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
// Offer is the shop's price for one member's gem or jewel (Appraise).
type Offer struct {
	Member, Item int
	Kind         string // Gems or Jewelry
	Value        int
	CanKeep      bool // Appraise.Keep is allowed
}

// ItemRef names one member's item.
type ItemRef struct{ Member, Item int }

type Shop struct {
	Offers        []Offer
	Name          string
	Goods         []Item
	Prices        []int
	Buys          bool // the shop buys from the party
	IdentifyPrice int
	Unknown       []ItemRef // the party's unidentified items the shop can identify
	Identified    *ItemRef  // the item last identified here, nil none
	Identifies    int       // identifications done here so far
}

type Service struct {
	Name  string
	Price int
	Needs []int          // party indices the service would help; the rest have no need of it
	Not   map[int]string // party index to why the service is of no use ("is not blind.")
}

type Training struct {
	Ready    []int // party indices able to level
	Cost     int
	Trains   map[int]string   // party index to what it trains to
	Becomes  map[int][]string // party index to its new level per class
	Why      map[int]string   // party index to why it may not train
	Learning bool             // a spell is being chosen (Learn), not training offered
	Learn    []SpellRef       // the spells the member may learn now
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
	Fields       []Field  // the draft's fields with their options (race, gender, class, alignment ...)
	Roster       []string // saved characters that can be added
	MinHP, MaxHP int
	Minimums     string // the rules' text on stat minimums for the chosen class
}

// Field is one creation choice: its options, which fit the draft so far,
// and the current pick (-1 none).
type Field struct {
	Name    string
	Options []string
	Fits    []bool
	Pick    int
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

// Ask is a domain question. The answer is an Answer decision. Options on an
// Ask of another kind are extra moves the player may make instead of an
// answer (a move question's Done): gobox applies the Option's Do as is.
type Ask struct {
	Kind    AskKind
	Prompt  string
	Text    string   // the question's body text
	Options []Option // Choice: labels with the original hotkeys
	Default int      // default option or number
	Max     int      // Name length, Number upper bound
	Lines   []Line   // PickItem: the list's rows
	Typed   string   // Name, Number: what is typed so far
	Ms      int      // Continue: a timed pause of Ms milliseconds (0: wait for a key)
}

// Line is one row of a PickItem list.
type Line struct {
	Text     string
	Header   bool  // a heading, not pickable
	Dim, Off bool  // shown dimmed; not pickable
	Spell    int   // the spell the row names, 0 if none
	Hides    []int // Ask.Options hidden while this row is chosen
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

// Adjust raises or lowers one of the creation draft's values by By (the
// original's modify step); Stat indexes StatKeys.
type Adjust struct{ Stat, By int }

// StatKeys is the order of the values Adjust names: the Member.Stats keys,
// then hit points, then exceptional strength (Member.Str00).
var StatKeys = []string{"str", "int", "wis", "dex", "con", "cha", "hp", "str00"}

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
		Member, Spell, Target int // Member: outside combat, who casts
		At                    image.Point
		Ats                   []image.Point // one square per target, when the spell takes several
	}
	UseItem struct {
		Member, Item, Target int
		Ats                  []image.Point // one square per target
		Spell                int           // a scroll's spell to read, 0 the first
	}
	Ready    struct{ Member, Item int }
	Drop     struct{ Member, Item int }
	Trade    struct{ From, To, Item int }
	Buy      struct{ Member, Item int } // Member: who buys and carries it
	Sell     struct{ Member, Item int }
	TakeLoot struct {
		Item int  // -1: the money only
		All  bool // the money and every item
	}
	Rest   struct{ Minutes int } // 0: the default rest
	Camp   struct{}
	Train  struct{ Member int }
	Heal   struct{ Service, Member int }
	Create struct {
		Field string
		N     int
		Text  string
	}
	AddChar struct {
		Roster int
		From   string // a roster the game imports from (Data.Imports), "" its own
	}
	DropChar struct{ Member int }
	Begin    struct{}
	Wait     struct{} // combat: end turn / hold
	Flee     struct{}
	Save     struct {
		Slot int
		Name string // the slot's name, gobox's; empty keeps the slot's own
	}
	Load struct {
		Slot int
		From string // as AddChar.From
	}

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
	Quick      struct{ All bool } // All: the whole party, else the unit whose turn it is
	Bandage    struct{ Member int }
	TurnUndead struct{}
	Aim        struct{ At image.Point }
	Ranged     struct{} // combat: shoot or throw the missile
	Items      struct{} // combat: open the inventory
	Maneuver   struct{} // combat: pick one of the rule set's maneuvers
	// Items and money.
	Unready  struct{ Member, Item int }
	Halve    struct{ Member, Item int } // split a stack
	Join     struct{ Member, Item int } // merge stacks
	Identify struct{ Member, Item int }
	Appraise struct {
		Member, Item int
		Keep         bool // keep the gem or jewel rather than sell it at the offer
	}
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
func (Ranged) decision()           {}
func (Items) decision()            {}
func (Maneuver) decision()         {}
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

// Data is declarative overrides only; no drawing, no positions. A game's
// rules package exports it as func Data() Data, and its cmd hands gobox
// the rules and the data together: app.Run(rules, data).
type Data struct {
	Series     string
	ID         string // game id
	Texts      map[string]string
	Credits    []string
	SpellRules map[int]string // what each spell does, by spell id (the spell tip text)
	Art        map[string]Picture
	Sounds     map[string]string
	Palette    []byte   // palette bytes
	Commands   []Option // extra party commands: Decisions with labels
	Imports    []string // other games' rosters and saves this game reads, by game id
}
