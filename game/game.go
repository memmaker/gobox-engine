// Package game is the UI-free contract between a Gold Box game and gobox.
// A game states domain facts (State), asks domain questions (Ask), accepts
// rule verbs (Decision) and reports Events. It never says how anything is
// drawn: no colours, cells, rects, inks or positions. Modelled on pod's
// rules.State / Decision / Apply / Pending / Event. Stdlib only.
package game

import (
	"image"
	"image/color"
	"io/fs"
	"time"
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
	RestHeals, RestStudies bool            // what resting now does
	TempleOffers           []Offer         // the temple\'s appraisals, as Shop.Offers
	PartyCap               int             // most items the party carries in all, 0 no limit
	Bestiary               []Beast         // the monsters the party has met
	Picker                 *Picker         // a grid picker is up
	IconEditor             *IconEditor     // the combat icon designer is up
	IconPalette            []color.RGBA    // the colour of each icon colour index 0..15 as the art set shows it (empty: the EGA colours)
	Page                   *Page           // the text page: the game's scripted text and story pictures
	Flags                  *Flags          // the plain status flags of the screen
	Walls                  *Walls          // the 3D view the rules build from their own map (nil: the remastered map)
	Intro                  *Intro          // the title waits to be played
	Sequence               string          // the Data.Sequences entry playing full screen ("": none); gobox sends SequenceDone at its end or on a skip
	Note                   Note            // the status line's toast
	Results                *Results        // the after-combat page
	Read                   *Read           // journal text the game waits to have read
	Cells                  image.Point     // the page's size in text cells
	Ink                    [16]color.RGBA  // the text colours (EGA 0..15) now (all zero: the EGA's own)
	Pacing                 *Pacing         // what goes on with time
	Window                 Window          // what the view window shows
	Holding                Hold            // what the game waits on out of a question
	Lore                   []Lore          // the journal's notes, newest first
	PowerPick              *PowerPick      // a member must choose a power
	PoolItems              int             // the party items held (Purse's room is PartyCap)
	Sheet                  *Sheet          // the selected member's sheet offers
	Items                  ItemsUp         // the inventory screen is up, and whose goods it shows beside the member's
	SheetPage              *SheetPage      // the character sheet's page is up
	Portrait               image.Rectangle // the sheet's portrait box in text cells, Max inclusive (zero elsewhere)
	IconBox                image.Rectangle // the creation sheet's combat icon box, as Portrait
	Screen                 *Screen         `json:",omitempty"` // a typed screen is up (nil: none); Pending, when set, is asked on its prompt rows

}

// Walls are the facts of a 3D view seen from the party's square: a wall
// face on every side. Fronts name the wall art (Data.Art), read live.
type Walls struct {
	Key                 string         // what the scene is built from: a new Key builds it again
	Side                [][][4]int     // a square's wall value by side (N E S W), rows from the top (0: open)
	Fronts              map[int]string // a wall value's art name in Data.Art
	Pool                map[int]string // a wall value's remastered texture name, in place of Fronts
	Floor, Sky          string         // the flats' pool names
	FloorRGBA, CeilRGBA *color.RGBA    // plain flat colours (nil: none)
	Sun                 bool           // by day outside: the sun up, at SunAz
	SunAz, SunEl        float64        // the sun's bearing (0 north, pi/2 east) and height
	Mountains           bool           // by day outside: the mountain range on the horizon
	// Backdrop, when set, is drawn opaque over the whole 3D view behind the
	// walls, stretched with nearest-neighbour filtering, in place of the sky
	// and floor flats (nil: none, the flats and CeilRGBA/FloorRGBA fills show).
	Backdrop *Picture `json:",omitempty"`
}

// Intro is the title waiting to be played.
type Intro struct {
	Full      bool  // picture 2 and the credits follow picture 1 (when Pictures is nil)
	Pictures  []int // the steps: N > 0 picture N, -N credits page N-1 (nil: 1, with Full 1 and 2; with no -N step the credits follow the last picture)
	Lightning bool  // picture 2 flashes
}

// Sequence is a full-screen animation gobox plays (State.Sequence names
// it): Anim the base name of its files in Data.Files (e.g. "INTRO": INTRO.ACF
// and INTRO.BMA), Music the songs it cues.
type Sequence struct {
	Name  string
	Anim  string
	Music []string
}

// Note is a toast for the status line: Text, put up at At (zero: none).
type Note struct {
	Text string
	At   time.Time
}

// Results is the after-combat page: the hired NPCs hiding their shares,
// else the outcome, or the party destroyed.
type Results struct {
	Shares  []string
	Outcome string
	Exp     string
	Taken   string
	Harmed  []Harm
	End     string
	EndAt   [4]int // the box's cells: X1, Y1, X2, Y2 (inclusive)
}

// Harm is a member's line on the after-combat page.
type Harm struct{ Name, What string }

// Read is journal text the game shows for the player to read.
type Read struct {
	Title string   // the box's title
	Lines []string // journalLayout's: "\x00" a title, "\x01" a picture, "\x02" its rows, "\x03key\x03" a voiced paragraph
	Page  int      // the page to start on
}

// Pace is the loop that goes on with time while the game waits.
type Pace int

const (
	PaceNone  Pace = iota
	PaceFight      // the fight's steps
	PaceRest       // camp's rest
	PaceStory      // the story's steps
	PaceWalk       // the auto-walk
	PaceAnim       // the picture's frames
)

// Pacing is what goes on with time.
type Pacing struct {
	Pace    Pace
	Resting bool   // camp's rest goes on: a key asks to stop it
	Seq     uint64 // counts the states published: a front redraws on a new one
}

// Window says what the view window shows.
type Window int

const (
	WindowScene   Window = iota // the 3D view
	WindowMap                   // the wilderness or city map
	WindowPicture               // a picture, portrait or approaching sprite
	WindowText                  // text in the whole window
)

// ItemsUp says whether the inventory screen is up and what it shows with the
// selected member's items: nothing, the shop's goods or the treasure.
type ItemsUp int

const (
	ItemsNone ItemsUp = iota // not up
	ItemsOwn                 // the member's own
	ItemsShop                // with the shop's goods
	ItemsLoot                // with the treasure
)

// Hold is what the game waits on besides a question.
type Hold int

const (
	HoldNone Hold = iota
	HoldShop
	HoldTemple
	HoldFight
	HoldStory
	HoldLoot
	HoldDoor
)

// Lore is a journal note: a story text read or a note the player wrote
// (ID its index, -1 for a story text).
type Lore struct {
	Title, Place string
	Text         []string
	ID           int
}

// PowerOption is one of the maneuvers a power offers; Aim is 0 a foe with the
// weapon, 1 none, 2 a foe in Range, 3 an ally in Range, 4 a square in Range
// with a burst of Area; Move the squares first walked.
type PowerOption struct {
	ID               int
	Name, Tip        string
	Aim, Range, Area int
	Move             int
}

// PowerPick is a member choosing a power: what it is and the maneuvers.
type PowerPick struct {
	Name, What string
	Options    []PowerOption
}

// Entry is one choice of a member's menu; Verb says what it does: "view",
// "items", "cast", "magic", "heal", "cure", "memorize", "scribe", "icon", "drop".
type Entry struct{ Name, Verb string }

// Reaction is one of the rule set's reactions a member has; Policy is 0
// always, 1 never, 2 only on a hit.
type Reaction struct {
	ID     int
	Name   string
	Policy int
}

// Sheet is what the character sheet offers besides the page.
type Sheet struct {
	Powers    []Entry
	Reactions []Reaction
}

// Stat is one of a rule set's figures on the sheet, with its tooltip lines.
type Stat struct {
	Name, Value string
	Help        []string
}

// SheetPage is what the character sheet's page shows.
type SheetPage struct {
	Name                        string
	NameColour                  int
	NPC                         bool
	Spells                      string
	Sex, Race                   string
	Age                         int
	Alignment, Class            string
	Stats                       [6]int
	Str00                       string
	Creating                    bool
	Levels                      string
	Exp, Next                   int
	AC, HP, MaxHP, THAC0        int
	Damage                      string
	Move                        int
	Saves                       [5]int
	Mods                        [6]int
	Rules                       []Stat
	Status                      string
	Weapon, Armour              string
	Encumbrance, Coins, Effects string
	PortraitAt                  image.Rectangle // as State.Portrait
	Portrait                    image.Image     // a dropped PNG in the box
	Icons                       [2]image.Image  // the creation's icon poses from a PNG
}

// Flags are the plain yes/no facts of the screen the front shows or takes
// decisions by.
type Flags struct {
	NoSave     bool // the rules refuse saving here (gobox's menus offer no Save)
	NoLoad     bool // ... and loading
	Exploring  bool // the party explores: the adventuring decisions are taken
	Sweeping   bool // ... at the slow, searching pace
	Interact   bool // ... on a square whose event waits for an interaction
	CanCast    bool // ... and the selected member has a spell to cast
	Adventure  bool // the party is out in the world
	Day        bool // 6:00 to 20:00
	Inside     bool // the party stands on an inside square
	Ordering   bool // the camp's party order bar is up
	Credits    bool // the title's credits page
	CreditPage int  `json:",omitempty"` // which one (Credit.Page) while Credits
	Full       bool // the pictures are one whole-screen picture
	IconDrop   bool // the combat icon designer is up
}

// Page is the text a game wrote and the pictures it showed: scripted text,
// story pages and the messages in the text window. Lines hold runs in the
// colours the game chose; the front lays them out.
type Page struct {
	Pane     [][]TextRun // the messages' lines (nil: no message shown)
	Prompt   []TextRun   // the prompt row's text
	View     [][]TextRun // text over the view window, rows from 1
	Layers   []Layer     // the pictures, bottom first
	PaneRows int         // >0: the front shows this many Pane lines at a time and waits for a key before the next lot (0: no paging)
}

// TextRun is text in one colour on a line from column X (Upcase: the
// glyphs of a game that upcases what it writes).
type TextRun struct {
	X      int
	Text   string
	Fg, Bg color.RGBA
	Upcase bool
	Blink  bool // the front alternates the run's ink with its highlight colour a few times, ending on normal
}

// Layer is a picture placed on the page: At in pixels of the page.
type Layer struct {
	Picture
	At      image.Rectangle
	Frames  []int // with FrameMs: the front cycles the picture through these frames (empty: static)
	FrameMs int   // the milliseconds each of Frames shows (0: static)
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
	Square    *image.Point // the square the 3D view stands on where X, Y is another map's (the overland); nil: X, Y
	Facing    int          // 0 N, 2 E, 4 S, 6 W
	Inside    bool
	Day       int
	Hour      int
	Minute    int // of the hour
	Month     int
	Year      int
	Mode      string // travel mode: "dungeon", "wilderness", "town", ...
	Searching bool
	Hall      bool           // the party is where characters may join or leave it (a guild hall, an inn)
	Music     string         // the place's own music trigger ("new-phlan": data/music/new-phlan), over gobox's generic ones; "" none
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
	PortraitHeadID, PortraitBodyID       string         // gobox's opaque shared-pool picks, one pool name each ("head_NNN", "body_NNN"); the game only stores and states them back ("": none, gobox falls back to Portrait)
	IconHeadID, IconBodyID               string         // likewise the combat icon's ("combat_head_NNN", "combat_body_NNN"; "": none, gobox falls back to Icon)
	BaseStats                            map[string]int // stats before magic and age
	IconColors                           []byte         // icon palette indices
	ClassText, LevelText                 string         // e.g. "old/new", "3/3"
	Slots                                map[string]int // spell slots left by kind/level
	Okay                                 bool
	ItemCap                              int            // most items carried, 0 no limit
	SlotsMax                             map[string]int // spell slots in all, keys as Slots
	CanCast                              bool           // may cast now (status, no-magic area)
	Powers                               []Option       // the member\'s special powers menu (Decisions)
	Chosen                               bool           // the card is marked: the selected member, or the candidate while one is picked
	Colour                               int            // the roster colour of the name (EGA number)
	Head                                 string         // the head the player chose: a Data.Art name, "" none
	Dropped                              image.Image    // the PNG dropped on the sheet in place of the head, nil none
	Card                                 []Entry        // the card's context menu out of combat
	LevelExp, NextExp                    int            // the experience the level took and the next takes (0: none)
	SpellsReady, Memo                    string         // spells ready over the slots ("0/3"), in camp too
	Effects                              []Line         // the sheet's list of spell effects on the member, as the game words it
	Foe                                  bool           // on the monsters' side (gobox colours its name as such); last, for parity's unsetLate
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
	ID, Text     string
	Done         bool
	Giver, Place string
	Brief        []string // Text by paragraph
}

// Automap is what the party knows of the map.
type Automap struct {
	Area      string // the area's name as the player knows it ("New Phlan", "Valjevo Castle 3"): the map shows it top left
	Size      image.Point
	Known     []image.Point // squares seen; empty: nothing is known
	Marked    []image.Point
	Notes     []string      // the player's text for each Marked square
	NoteKinds []string      // a Label.Kind for each Marked square, \"\" none
	Walls     []Wall        // by square
	Doors     []Wall        // State says what each door is
	Walked    []image.Point // squares walked on (Known also holds those only seen)
	Labels    []Label       // the map's own labels; the player's are Marked/Notes
	Marker    string        // the party marker's art prefix on this map (Data.Art "<Marker>/<facing>"); "" is "marker"
	// Terrain is the overland map's terrain id per square, row by row over
	// Size; its art is Data.Art["terrain/<id>"] (a game numbers each
	// remapped variant as its own id and gives its Pixels), the party marker
	// Data.Art["marker/<facing>"]. Known and Walked apply as on a dungeon
	// map. Empty for a dungeon map.
	Terrain []int
	// NoMap marks the area unmappable: while the player's "No Map Zones"
	// option is on, gobox shows no minimap and no full map here.
	NoMap bool `json:",omitempty"`
}

// Label is a named place on the automap.
type Label struct {
	At   image.Point
	Text string
	Kind string // "shop", "temple", "inn", "civic", ...
}

// Map is the area the party walks, as the game knows it; gobox moves the
// party on it and tells the game with Arrive and Face. A game whose Rules
// also implement Mapped gives it.
type Map interface {
	Name() string      // the area's name as the player knows it (the map screen's title)
	Size() image.Point // in squares
	// Edge is side (as Where.Facing: 0 N, 2 E, 4 S, 6 W) of square at: whether
	// a step goes through it.
	Edge(at image.Point, side int) Edge // at in Where.X, Y (the overland's too)
	// Departure reports whether a script would run ahead of a step through
	// side of square at (before the party moves); gobox runs it with
	// RunEvent{Before: true, Side: side}, then takes the step only if the
	// party still stands there, facing as it did, with no question up.
	Departure(at image.Point, side int) bool
	// Event reports whether an event would run by itself on square at now
	// (a step script, an encounter); gobox runs it with RunEvent.
	Event(at image.Point) bool
}

// Mapped is Rules with a walkable map: Map is nil while the party walks
// none (a menu, a shop, combat) or may not walk now (a question is up).
type Mapped interface {
	Map() Map
}

// Edge is one side of a square to a step.
type Edge struct {
	Kind  EdgeKind
	State DoorState // a door's (EdgeDoor)
}

// EdgeKind is what stands on a side of a square.
type EdgeKind int

const (
	EdgeOpen EdgeKind = iota // nothing: a step goes through
	EdgeWall                 // a step does not
	EdgeDoor                 // a door: gobox sends the step and the game opens it, or refuses (Refusal.Ask)
)

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

	// The ground as a view draws it: Squares' Tile numbers index Tiles, the
	// Data.Art names of the tile set (read live; "" a tile with no picture),
	// Floor is the bare ground's colour.
	Tiles     []string
	FloorTile string `json:",omitempty"` // the Tiles name (a series combat sprite) of the plain floor, tiled on every floor square; "": the bare ground's Floor colour
	Floor     color.RGBA
	Effects   []Effect    // sprites over the map
	Cursor    image.Point // the square the cursor was last drawn on
	CursorOn  bool        // the cursor icon is up (aiming, looking)
	CursorAt  []image.Point
	View      image.Point // top left square of the scrolled view window
	Strike    Strike      // the last weapon attack
	Log       []string    // the battle's messages, oldest first
}

// Effect is a sprite over the combat map: its top left corner in thirds of a
// square.
type Effect struct {
	X, Y int
	Art  Picture
}

// Strike is the last weapon attack; Seq counts attacks, From and To are unit IDs.
type Strike struct {
	Seq, From, To int
	Hit           bool
	Damage        int
}

// Step is one square of a walk; Free: leaving it lets an enemy take a free attack.
type Step struct {
	At   image.Point
	Free bool
}

// Reach is a square the unit can walk to and the steps it costs.
type Reach struct {
	At    image.Point
	Steps int
	Path  []Step // the walk there from the unit's own square
}

// Ground is one battlefield square.
type Ground struct {
	At     image.Point
	Art    Picture
	Block  bool
	Height int // the square's height (a wall, a tree)
	Tile   int // the tile number (Combat.Tiles)
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
	Dmg        string      // the unit's damage, e.g. "1d8+6"
	Member     int         // party index of a party unit, -1 for a monster
	Active     bool        // it is the unit the player acts for
	Hit        int         // the active unit's chance in percent to hit it with one blow, 0 not a target
	Pose       int         // its icon's pose: 0 ready, 1 action
	Icon       image.Image // a member's own PNG icon, in its pose and facing (nil: composed from its icon picks)
}

// Turn is whose turn it is and what they may do.
type Turn struct {
	Unit      int
	Moves     int
	Casting   int    // the spell the unit is casting over rounds, 0 none
	Aim       bool   // aiming: Aim takes a square (Cast: Ats), Cancel gives up
	Spell     string // aiming a spell, its name
	Options   []Option
	Weapon    *Weapon        // the weapon in hand (its icon on Attack), nil none
	Missile   *Weapon        // the ranged weapon readied (its icon on Ranged), nil none
	Maneuvers []TurnManeuver // the rule set's maneuvers (Maneuver picks one); none: no Maneuver
}

// Weapon is a Turn's weapon: its icon (Type, Words) and tooltip.
type Weapon struct {
	Name        string
	Type        int
	Words       []string
	Look, Rules []string // what it is, how it attacks
}

// TurnManeuver is one of a Turn's maneuvers.
type TurnManeuver struct {
	ID          int
	Name, Tip   string
	Aim         int // front.Aim
	Range, Area int
	Shift       int // AimWeapon: squares the member may move first
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
	Buys          bool           // the shop buys from the party
	Reloads       bool           // the shop refills clips (Reload)
	Pooled        map[string]int // money pooled with the shop, by coin (nil: none)
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

// Loot is the treasure up (Mode ModeLoot). The contract, the same in every
// game: while Loot is set and Items is ItemsNone, gobox shows its one loot
// screen (the member's inventory beside the treasure: Money and Items) and
// draws and reads it all; the game publishes Loot and answers:
//   - TakeLoot{Item: i}: take Items[i] for the selected member;
//   - TakeLoot{Item: -1}: take the money; TakeLoot{Item: -1, All: true}: all;
//   - LootDetect{} (offered while there are items and Allowed["LootDetect"],
//     or when unset while the selected member has Detect Magic memorized),
//     LootShare{}, PoolCoins{}, LootStatus{} (Status set);
//   - LootLeave{Sure: false}: leave; while treasure is left, refuse it with
//     a Refusal whose Ask is YesNo (Prompt the question): gobox asks it and
//     sends LootLeave{Sure: true} on Yes.
//
// Select and ShowMember pick whose inventory it shows. A game needs no View
// or Take of its own (LootLook, LootTake are not sent by this screen) and
// clears Loot when the party leaves.
type Loot struct {
	Items  []Item
	Money  map[string]int
	Status bool `json:",omitempty"` // the menu offers Status (LootStatus); the game answers by publishing a screen (the ship sheet over the fight)
}

type Door struct {
	Locked, Stuck, Secret bool
	Bash, Pick, Knock     bool // the ways to get through a locked door
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
	// Stats are the abilities the sheet shows, in the game's order and
	// count (nil: STR INT WIS DEX CON CHA from Draft.Stats, with 18/xx).
	Stats []StatDef `json:",omitempty"`
	// Derived are the figures the sheet shows besides the abilities
	// (age, AC, THAC0, ...), in the game's order. When set they are the only
	// source: a figure not given is not shown. nil: the figures come from
	// Draft.
	Derived []Derived `json:",omitempty"`
}

// CreationFlow picks gobox's generic creation sheet layout and step
// sequence; the game's Fields, Stats and Derived carry the meaning.
type CreationFlow int

const (
	// CreationADnD1 is the Pool games' sheet (the default): four fields at
	// fixed places, six abilities with 18/xx, the Minimal roll and the
	// power picks after the fourth field.
	CreationADnD1 CreationFlow = iota
	// CreationADnD2BR is Buck Rogers' sheet: every field one per row in the
	// game's order (no power picks), Stats and Derived drawn as given.
	CreationADnD2BR
)

// StatDef is one ability on the creation sheet.
type StatDef struct {
	Key      string // the Member.Stats key ("str", "tech", ...), for the game
	Label    string // what the sheet shows ("STR", "TECH", ...)
	Min, Max int    // the range the rules allow it (0, 0: not given)
	Value    int    // its current value
	Extra    string // shown after the value, e.g. "(00)" for an exceptional strength
}

// Derived is one figure on the creation sheet: its label and its value as
// the sheet shows it.
type Derived struct {
	Label string // as the sheet shows it ("Age", "AC", ...)
	Value string // "21", "10", "3d4", ...
}

// Field is one creation choice: its options, which fit the draft so far,
// and the current pick (-1 none).
type Field struct {
	Name    string
	Options []string
	Fits    []bool
	Pick    int
}

// SaveSlot is a save slot: its name, info text, when it was saved (as the
// slot list shows it) and the highest level in its party (0: none read).
type SaveSlot struct {
	Name, Info string
	Time       string
	Level      int
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
	Yes     bool     `json:",omitempty"` // YesNo: the highlight starts on Yes (false: on No)
	Max     int      // Name length, Number upper bound
	Lines   []Line   // PickItem: the list's rows
	Typed   string   // Name, Number: what is typed so far
	Ms      int      // Continue: a timed pause of Ms milliseconds (0: wait for a key)
	Break   int      `json:",omitempty"` // Choice: >0 puts Options[Break:] on a second bar row (the original's '@'); the bar takes rows 23-24 and Up/Down switch rows
	View    bool     // while set, the player may open the View menu on the highlighted option's member (game.ViewMember{N}) and the ask stays up; used by member picks
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
	// Arrive: gobox moved the party from From onto the next square To
	// (Map.Edge let it through), facing Facing (a step back or aside keeps
	// it). To may lie off the map (Map.Size): the game wraps or leaves the
	// area. The game only moves the party (and the step's time): it runs no
	// event. A Refusal's Ask is the question put instead (a locked door's
	// Bash / Pick / Knock). Then gobox asks Map.Event of the square and
	// sends RunEvent when it says one waits.
	Arrive struct {
		From, To image.Point
		Facing   int // as Where.Facing
	}
	// RunEvent: run an event Map says waits: Before, the script the game
	// runs ahead of a step through Side (as Where.Facing) from the party's
	// square, before gobox moves it (Map.Departure: it may hold the step,
	// turn or move the party, or ask); else the event on the party's square
	// after Arrive (Map.Event: a step script, an encounter). gobox's call.
	RunEvent struct {
		Before bool
		Side   int
	}
	// Face: gobox turned the party to Facing (as Where.Facing) on its square.
	Face   struct{ Facing int }
	Search struct{}
	Look   struct{}
	Select struct{ N int }
	Move   struct{ To image.Point } // combat
	Attack struct{ Target int }
	Cast   struct {
		Member, Spell, Target int // Member: outside combat, who casts
		At                    image.Point
		Ats                   []image.Point // one square per target, when the spell takes several
	}
	UseItem struct {
		Member, Item, Target int
		Ats                  []image.Point // one square per target
		Spell                int           // a scroll's spell to read, 0 the first
	}
	Ready    struct{ Member, Item, Count int } // Count of a stack to ready (0: all of it)
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
		Field string // "portraithead", "portraitbody", "iconhead" or "iconbody": Text is gobox's opaque pool name, stored as Member.PortraitHeadID .. IconBodyID
		N     int
		Text  string
	}
	SetStat struct {
		Stat int    // 0..5: STR .. CHA
		Text string // the value as typed ("18/50" for an exceptional strength)
	} // creation: set an ability by hand
	// PointBuy: creation: the abilities bought on points (pointbuy.go),
	// offered only when Data.PointBuy. Scores are the base values in
	// Creation.Stats' order (Stats nil: STR .. CHA, StatKeys 0..5), each
	// PointBuyMin..PointBuyMax and within PointBuyBudget; the game applies
	// its rules to them (race, class minimums, exceptional strength) as
	// after a roll, or refuses.
	PointBuy  struct{ Scores []int }
	KeepDraft struct{} // creation and Modify: keep the draft (save the new character, apply the changes)
	AddChar   struct {
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
	Fix        struct{ Member int }
	Memorize   struct{ Member, Spell int }
	Forget     struct{ Member, Spell int }
	Scribe     struct{ Member, Spell int }
	Order      struct{ From, To int }
	EndGame    struct{}              // the main menu: end the game in progress, back to the party menu
	AddNote    struct{ Text string } // a player journal note of Text (the line gobox shows), with the place; a duplicate is refused
	RemoveNote struct{ ID int }      // delete player note ID
	LeaveCamp  struct{}
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
	Reload    struct{ Member, Item int } // refill a clip (the shop's Ammo / Fill Clip)
	Leave     struct{}                   // leave the open shop, vault, loot or temple
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
		Field  string // "portraithead" .. "iconbody" with ID: store that opaque pool name
		N      int
		ID     string
	}
	Call struct{ Member int }
)

func (Answer) decision()    {}
func (Cancel) decision()    {}
func (Adjust) decision()    {}
func (Arrive) decision()    {}
func (Face) decision()      {}
func (RunEvent) decision()  {}
func (Search) decision()    {}
func (Look) decision()      {}
func (Select) decision()    {}
func (Move) decision()      {}
func (Attack) decision()    {}
func (Cast) decision()      {}
func (UseItem) decision()   {}
func (Ready) decision()     {}
func (Drop) decision()      {}
func (Trade) decision()     {}
func (Buy) decision()       {}
func (Sell) decision()      {}
func (TakeLoot) decision()  {}
func (Rest) decision()      {}
func (Camp) decision()      {}
func (Train) decision()     {}
func (Heal) decision()      {}
func (Create) decision()    {}
func (SetStat) decision()   {}
func (PointBuy) decision()  {}
func (KeepDraft) decision() {}
func (AddChar) decision()   {}
func (DropChar) decision()  {}
func (Begin) decision()     {}
func (Wait) decision()      {}
func (Flee) decision()      {}
func (Save) decision()      {}
func (Load) decision()      {}

func (Fix) decision()              {}
func (Memorize) decision()         {}
func (Forget) decision()           {}
func (Scribe) decision()           {}
func (Order) decision()            {}
func (EndGame) decision()          {}
func (AddNote) decision()          {}
func (RemoveNote) decision()       {}
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
func (Reload) decision()           {}
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
	EvSaves   // gobox puts its save screen up (a script's "save the game?" choice)
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
	Ms           int     // >0 with Page: the front shows Page for Ms milliseconds (a timed moment outside a fight)
	Page         *Page   // the whole page as it stood for this moment (nil: none)
	Pose         int     // >0 with Ms in a fight: Unit's icon pose (1 action) while this moment shows
	Screen       *Screen // with Ms: the typed screen as it stood for this moment (a blink, a blast frame), shown for Ms milliseconds
}

// Data is declarative overrides only; no drawing, no positions. A game's
// rules package exports it as func Data() Data, and its cmd hands gobox
// the rules and the data together: app.Run(rules, data).
type Data struct {
	SaveDir      string // where gobox writes slot files ("": saves)
	Series       string
	ID           string          // game id
	CreationFlow CreationFlow    // the creation sheet's layout and steps
	PointBuy     bool            `json:",omitempty"` // the game takes PointBuy: creation asks Roll or Point Buy for each new character
	VerbKeys     map[string]rune `json:",omitempty"` // the series' hotkeys of combat verbs by label ("Aid": 'H'), over gobox's
	NoHotkeys    bool            `json:",omitempty"` // the series' menus are picked by arrows only: gobox gives its menus no hotkeys
	PartyMenu    []MenuKey       `json:",omitempty"` // the series' party menu: gobox's function labels in its order with its hotkeys (nil: gobox's)
	Texts        map[string]string
	Credits      []string
	SpellRules   map[int]string // what each spell does, by spell id (the spell tip text)
	Art          map[string]Picture
	Sounds       map[string]string
	Palette      []byte                        // palette bytes
	Commands     []Option                      // extra party commands: Decisions with labels
	MenuItems    []MenuItem                    // the game's own main menu lines, after gobox's
	MusicPlaces  []string                      // every Where.Music trigger the game sets (gobox makes their data/music folders)
	Imports      []string                      // other games' rosters and saves this game reads, by game id
	QuickParty   []QuickPick                   // the party menu's Quick Start party (every game offers it while the party is empty)
	Clues        []Clue                        // the clue book's chapters (nil: the game has none)
	Scans        func(name string) image.Image // the picture a Read's line "\x01name" shows (nil: none)
	Font         string                        // the ui/fonts pool name of the game's font ("": gobox's default)
	Fonts        []Typeface                    // the game's own text fonts (non-nil, even empty: gobox offers the font menu, Ctrl-F and keeps the choice; the first is listed first)
	Border       *Border                       // the window border's pieces in the game's DAX symbols (nil: gobox's default)

	Sequences    []Sequence     // the full-screen animations State.Sequence names (nil: none)
	DaxMask      map[string]int `json:",omitempty"` // a DAX file's transparent colour index, as the original's loader takes it (File, e.g. "CHARS", "CPIC3"; no entry: none)
	Files        string         // the game's DOS files: the DAX archives Picture File and Block, Border and AdLib name ("": none)
	SoundTables  []byte         // the sound driver's tables (gobox/audio NewSynth's data; nil: none)
	AdLib        *AdLib         // the AdLib music and effects (nil: none)
	AmigaMusic   *AmigaMusic    // the Amiga's Music-X songs and GLIB effects, played instead of the DOS sound (nil: none)
	DungeonTiles int            // the tiles at the start of the combat set
	Rests        []RestKind     // camp's rests, in the menu's order (nil: none)
	CreditLines  []Credit       // the title's credits page, placed (nil: Credits as plain lines)
	CreditGaps   []int          // the credits' rows a blank row follows on a tall page (nil: 4, 5, 6)
	CreditFrame  string         // the credits' frame: "" two boxes (title, lines) with CreditGaps and the remaster's lines; "edges" one frame on the page's edges, the lines at their cells
	IconParts    []string       // the combat icon designer's parts
	ReadyMark    string         // shown before a readied item's name ("": none)
	LabelInks    []int          // a map label's EGA colour by its kind
	MapCell      int            // pixels a square of the area map (0: gobox's)

	DoorInk, SelectInk color.RGBA     // a door's colour on the maps; the frame round the selected box (zero: gobox's)
	Logo               *Logo          // how the main menu's lettering is cut from the title picture (nil: none)
	RemasterCredit     [2]string      // the remaster's own lines at the bottom of the credits page
	TitleBackdrop      int            // a title picture N (TitlePicture n) not among Intro.Pictures' timed steps: gobox puts it up once the steps and credits are done, keeps it fitted under the main menu and zooms the menu in over it, then TitleDone (0: none)
	AreaMaps           map[int]string // the remastered 3D map (Remaster maps/NAME.map) of each area id
	Remaster           fs.FS          // the remastered art: maps/NAME.map and assets/walls/walls.txt
	OpenAt             string         // debug: map:NAME:X:Y:D shows that remastered map alone
	Bindings           []Binding      // the game's own help rows, after the Any time Ctrl-F row
}

// Logo is how the title picture's lettering is cut out: EGA colours Keep
// above row Sub stay, from Sub to Bottom the SubInk pixels are drawn in Gold
// with an Outline round them; everything else is transparent.
type Logo struct {
	Sub, Bottom   int
	Keep          []uint8
	SubInk        uint8
	Gold, Outline color.RGBA
}

// Binding is one row of the help: in context Ctx, Keys (and Mouse, Pad) do Desc.
type Binding struct {
	Ctx, Desc        string
	Keys, Mouse, Pad string
}

// Typeface is an 8x8 text font: 256 glyphs of 8 rows, bit 7 leftmost.
type Typeface struct {
	Name string
	Bits []byte
}

// AdLib is the AdLib songs (file Music, song N its block N) and effects
// (file Sounds, effect N its block Effects[N]), unsigned 8-bit PCM at Rate Hz.
// Tempo is the sequencer's fixed tempo divisor (SotSB's SOUND.DAX: 0x76); 0
// takes each song header's tempo word / 84 (PoD).
type AdLib struct {
	Music, Sounds string
	Effects       map[int]int
	Rate          int
	Tempo         int
}

// AmigaMusic is an Amiga Music-X player's files in the Amiga data directory
// (data/amiga beside Files; amigaimport Spec.Files copies them off the disks):
// Songs the MXTX song file, Effects the GLIB library of DIG8 samples ("": none).
// EvSong n plays the file's song SongOf[n] (0-based; nil: n-1, 0 stops); the
// game's effect n (PlaySound n, 2 up) plays GLIB entry EffectOf[n-2].
type AmigaMusic struct {
	Songs, Effects string
	SongOf         map[int]int
	EffectOf       []int
}

// RestKind is one of camp's rests: Text with its Key rests the N way. Fix: offered
// only while someone heals or studies. Night, NightKey: its text and key at
// night ("": Text's). Duration: the player sets how long first.
type RestKind struct {
	Text     string
	Key      byte
	N        int
	Fix      bool
	Night    string
	NightKey byte
	Duration bool
}

// Credit is one line of the title's credits page at cell X, Y in EGA colour Colour.
type Credit struct {
	X, Y, Colour int
	Text         string
	Page         int // the credits page it is on (0: the first)
}

// Border is the window border as frames of block Block of DAX file File:
// Pieces by place (top-left, top-right, bottom-left, bottom-right, top,
// bottom, left, right; below 0: none), Caps the 4 end caps, Mask the colour
// index left transparent (nil: none), and Edges an edge's cells one by one
// ("top", "bottom", "rule", "left", "right", "divider"; see gobox's present).
type Border struct {
	File   string
	Block  int
	Pieces [8]int
	Caps   []int
	Mask   *uint8
	Edges  map[string][]int
}

// QuickPick is one Quick Start member: the creation answers by the names the
// game's sheet shows, the abilities rerolled for (StatKeys key -> minimum),
// and the name and portrait it is saved under (Head and Body are pool
// portrait names, sprites series "pool", kind "portraits", e.g. "head_003").
type QuickPick struct {
	Race, Gender, Class, Align string
	Need                       map[string]int
	Name                       string
	Head, Body                 string
}

// Clue is a clue book chapter: its title, and the book read from it to its
// end, laid out as a Read's Lines are.
type Clue struct {
	Title string
	Read  Read
}

// Tone is what a line or figure means; gobox picks its colour.
type Tone int

const (
	ToneText   Tone = iota // plain text
	ToneOwn                // own crew and ship events, own labels
	ToneBad                // dead, destroyed, zero
	TonePrompt             // prompts, table titles
	ToneWarn               // the foe, dying, damaged
	ToneLoud               // headers, values
)

// ScreenKind is the Buck Rogers-style screen a game asks gobox to put up
// (State.Screen). The kind comes from the game's state, never from its id.
type ScreenKind int

const (
	ScreenNone        ScreenKind = iota
	ScreenSpaceFight             // the ship fight: the view, both ships' systems, the weapons, the message window
	ScreenShipSheet              // one ship's record
	ScreenResultTable            // medic / repair columns
	ScreenStarmap                // the system map with the rocket
	ScreenIconGrid               // the 6x6 body icons
	ScreenSkillPoints            // the skill table
	ScreenWeaponSpec             // the Weapon Specialization list
	ScreenTrainOffer             // "NAME will become: a level N CAREER"
	ScreenCharSheet              // the BR character sheet
)

// Screen is the typed screen up: Kind says which pointer is set.
type Screen struct {
	Kind       ScreenKind
	SpaceFight *SpaceFight // Kind ScreenSpaceFight; with Kind ScreenShipSheet: the own panel and compact weapons beside the sheet
	ShipSheet  *ShipSheet  `json:",omitempty"` // Kind ScreenShipSheet: the sheet in place of the fight's left half; with Kind ScreenSpaceFight: drawn over it (Sensors, View Ship)
}

// SpaceFight is the ship fight's screen. gobox holds the layout: where the
// view, the system icons and the weapons piece go (the SHIPS pictures).
type SpaceFight struct {
	Foe        Picture      // the foe's picture in the picture window (File "": none)
	Status     string       // "Closing    " / "Retreating " / "Maneuvering" as the game's text holds it
	Range      int          // shown as "Range: N" (below 0: 0)
	Own, Enemy ShipSystems  // the six systems of each ship
	SensorsOut bool         // own Sensors are 0: the enemy panel shows "Sensors Destroyed" only
	Weapons    []WeaponSlot // own slots 0..2 for the compact panel
	Blink      Blink        // what blinks in this snapshot (zero: nothing)
	Blast      *Blast       // a blast frame drawn over the view (nil: none)
	Log        []Message    // the message window's lines since the game last cleared it
	Pause      string       `json:",omitempty"` // the overflow / key pause's text on row 24 ("": "Press button or enter to continue.")
	StatusLine string       `json:",omitempty"` // a status line on row 24 col 0 in ToneText in place of the bar (pO0F6_0A18 / string_print01: the dying and coma notices); the game shows it as an Event for its pause
}

// ShipSheet is one ship's record (BR1 pO0F6_3346, BR2 pO099_3472) on
// cells (1,1)-(24,17). Ask Continue on it for the key.
type ShipSheet struct {
	Own         bool         `json:",omitempty"` // the own ship: labels ToneOwn, medical supplies on row 11 (else the foe: ToneWarn, the range)
	Known       bool         `json:",omitempty"` // the foe's AC, crew, speed, react, ammo and reloads are shown (false: '?' in ToneBad); the game sets it from the sensor roll and own Sensors > 0
	Name        string       `json:",omitempty"` // the ship's name
	Type        string       `json:",omitempty"` // the ship type's text
	Race        string       `json:",omitempty"` // the race's text
	Tonnage     int          `json:",omitempty"` // shown with " tons"
	Length      int          `json:",omitempty"` // shown with " feet"
	Width       int          `json:",omitempty"` // shown with " feet"
	Cargo       int          `json:",omitempty"` // shown with " tons"
	AC          int          `json:",omitempty"` // armour class
	Crew        int          `json:",omitempty"` // crew figure
	Speed       int          `json:",omitempty"` // the effective speed (pO0F6_321B / pO099_3348)
	React       int          `json:",omitempty"` // reaction figure
	Range       RangeBand    `json:",omitempty"` // the foe only: the range band
	MedSupplies int          `json:",omitempty"` // the own ship only: the party's medical supplies
	Weapons     []WeaponSlot `json:",omitempty"` // mounts 0..3; the count shown is Working only
}

// RangeBand is the foe's range as the sheet names it.
type RangeBand int

const (
	RangeRamming RangeBand = iota // range < 1
	RangeShort                    // < 3
	RangeMedium                   // < 5
	RangeLong                     // < 7
	RangeExtreme                  // else
)

// ShipSystems are a ship's Hull Sens Ctrl Life Fuel Engn figures.
type ShipSystems struct {
	Now, Max [6]int
}

// WeaponSlot is one weapon mount line.
type WeaponSlot struct {
	Short           string // the 8-character short name
	Mounts, Working int    // a slot with Mounts 0 is not shown
	Ammo, Reloads   int    // the ship sheet's columns
}

// Blink is what blinks in a SpaceFight snapshot (the original's 70 ms
// half-steps, ten of them, ending on the normal colour).
type Blink struct {
	Own    bool // the system is the own ship's (else the foe's)
	System int  // 1..6: Hull..Engn's figure blinks (0: none)
	Slot   int  // 1..3: that own weapon line blinks (0: none)
}

// Blast is one blast frame: the SHIPS picture table's entry K (1..15:
// three sizes by five frames) at the cell the game rolled.
type Blast struct {
	K  int
	At image.Point
}

// Message is a line of a screen's message window.
type Message struct {
	Tone  Tone
	Text  string
	Clear bool // the window is cleared first, after a key pause when it holds text (the original's pause-and-clear)
}

// MenuKey is a menu function by gobox's label, with the hotkey the series
// gives it.
type MenuKey struct {
	Label string
	Key   rune
}

// MenuItem is a main menu line of the game's: gobox plays Sequence (a
// Data.Sequences name) full screen and comes back to the menu; no game code runs.
type MenuItem struct {
	Label    string
	Key      rune
	Sequence string
}
