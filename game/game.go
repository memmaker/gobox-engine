// Package game is the UI-free contract between a Gold Box game and gobox.
// A game states domain facts (State), asks domain questions (Ask), accepts
// rule verbs (Decision) and reports Events. It never says how anything is
// drawn: no cells, rects or positions, and colours only as the original's
// text inks by EGA name (Ink) and art's own palettes. The title, ending
// and credits are the adventure folder's scripts (gobox/adventure.Sequence),
// which gobox composes; State names only the step up (Intro). Not yet so: the live page's
// pictures where the original drew them (Drawn). Modelled on pod's
// rules.State / Decision / Apply / Pending / Event. Stdlib only.
package game

import (
	"image"
	"image/color"
	"path/filepath"
	"time"

	"gobox/rules"
)

// Picture is an art ref: a DAX File/Block/Frame (Block -1: none), or a Name,
// plus optionally raw indexed pixels (Pixels are palette indices, W*H long).
type Picture struct {
	Source       PicSource // what File names
	File         string
	Block, Frame int
	Name         string
	W, H         int
	Pixels       []byte       // palette indices, W*H
	Palette      []color.RGBA // with Pixels: the art's own colours by index (an Amiga block's palette, a VGA picture's; alpha 0 clear); nil: EGA and asset.Amiga
	Clear        int          // with Pixels: the index drawn transparent, -1 none (ignored without Pixels)
	Alt          *Picture     // with Pixels: the DOS art's twin (F12) where it differs, the front shows one; nil: the same in both
	Off          image.Point  // in an Alt: its place's offset from the picture it stands in for (a bigger twin centred)
}

// PicSource is what a Picture's File names.
type PicSource uint8

const (
	SourceDAX PicSource = iota // a DAX archive (Block, Frame), or a pool Name
	SourcePNG                  // a PNG file's path (POR's adventure art/TITLE.PNG)
)

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

// Walls are the facts of a 3D view seen from the party's square: a wall
// face on every side. Fronts name the wall art (Data.Art), read live.
type Walls struct {
	Key               string         // what the scene is built from: a new Key builds it again
	Side              [][][4]int     // a square's wall value by side (N E S W), rows from the top (0: open)
	Fronts            map[int]string // a wall value's art name in Data.Art
	Pool              map[int]string // a wall value's remastered texture name, in place of Fronts
	Floor, Sky        string         // the flats' pool names
	FloorInk, CeilInk Ink            // plain flat colours, EGA inks (InkNone: none)
	Sun               bool           // by day outside: the sun up, at SunAz
	SunAz, SunEl      float64        // the sun's bearing (0 north, pi/2 east) and height
	Mountains         bool           // by day outside: the mountain range on the horizon
	// Backdrop, when set, is drawn opaque over the whole 3D view behind the
	// walls, stretched with nearest-neighbour filtering, in place of the sky
	// and floor flats (nil: none, the flats and CeilInk/FloorInk fills show).
	Backdrop *Picture `json:",omitempty"`
	Scenery  Scenery
}

// Intro is the title (or the end) waiting to be played: the steps of the
// adventure's Start script (End with End), gobox's to compose.
type Intro struct {
	Full      bool  // picture 2 and the credits follow picture 1 (when Pictures is nil)
	Pictures  []int // the steps: N > 0 the script's step N (adventure.Sequence.Step), -N credits page N-1 (nil: 1, with Full 1 and 2; with no -N step the credits follow the last picture)
	Lightning bool  // picture 2 flashes
	End       bool  // the End script's steps, not the Start's
	Party     int   // the party's members: a layer with a higher MinParty is not shown (End)
	MsgDelay  int   // the game's message delay: a MsgDelayed layer's frames are paced by it (End)
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
}

// Harm is a member's line on the after-combat page.
type Harm struct{ Name, What string }

// Read is journal text the game shows for the player to read.
type Read struct {
	Title string     // the box's title
	Texts []ReadText // the texts in order; ui lays them out
	Page  int        // the page to start on
	Seq   int        // up by one for each new text: the front shows a new Seq once, owns its paging and closing, and never reports back
}

// ReadText is one titled text of a Read.
type ReadText struct {
	Title string
	Paras []ReadPara
}

// ReadPara is a paragraph of a ReadText, as written: Text "" is a blank line
// (and no Picture), Picture names a picture instead of text, Voice (""
// none) is the paragraph's recording key.
type ReadPara struct{ Text, Picture, Voice string }

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
type Stat = rules.Figure

// SheetPage is what the character sheet's page shows.
type SheetPage struct {
	Name                        string
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
	AC, HP, MaxHP, THAC0        int // AC: the front armour class
	RearAC                      int // the rear one
	Damage                      string
	Move                        int
	Saves                       [5]int
	Mods                        [6]int
	Rules                       []Stat
	Status                      string
	Weapon, Armour              string
	Encumbrance, Coins, Effects string
	HasPortrait                 bool   // the member has a portrait (a head and body, or a dropped PNG): the sheet shows its box
	Portrait                    []byte // a dropped PNG in the box, its bytes (gobox decodes it)
	IconPNG                     []byte // the creation's combat icon from a PNG, its bytes (gobox cuts the ready and action poses)
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
	CreditsPic bool `json:",omitempty"` // the title step shows the game's credits as a picture (adventure.SequenceStep.Credits)
	CreditPage int  `json:",omitempty"` // which one (adventure.Credit.Page) while Credits
	Full       bool // the pictures are one whole-screen picture
	IconDrop   bool // the combat icon designer is up
}

// TextRun is text in one of the original's colours, a run of a line
// (a game font draws its own capitals; the text keeps its case).
type TextRun struct {
	Text   string
	Fg, Bg Ink
	Blink  bool // the front alternates the run's ink with its highlight colour a few times, ending on normal
}

// Ink is a text colour of the original by its EGA name; gobox maps it to
// its palette. InkNone: gobox's own (the body text's ink, no background).
type Ink uint8

const (
	InkNone Ink = iota
	InkBlack
	InkBlue
	InkGreen
	InkCyan
	InkRed
	InkMagenta
	InkBrown
	InkLightGrey
	InkDarkGrey
	InkLightBlue
	InkLightGreen
	InkLightCyan
	InkLightRed
	InkLightMagenta
	InkYellow
	InkWhite
)

// EGAInk is EGA colour n (0..15) as an Ink.
func EGAInk(n int) Ink { return InkBlack + Ink(n&15) }

// EGA is the ink's EGA colour number (0..15); ok false for InkNone.
func (k Ink) EGA() (n int, ok bool) { return int(k) - int(InkBlack), k != InkNone }

// Drawn is a picture where the original drew it on its 320x200 page, At
// in that page's pixels: the live script's page (Common.Layers). gobox
// does not place these yet (REFACTOR N2's remainder).
type Drawn struct {
	Picture
	At     image.Rectangle
	Anim   []Drawn         // the frames it cycles while the game waits (the original's wait-loop animation), the first its own; nil: still
	Delays []time.Duration // with Anim: each frame's time
}

// Picker is a grid picker up (the creation's heads and bodies, the icon's
// parts, the saved characters): every picture it holds, in order; the front
// lays them out in pages.
type Picker struct {
	Cur       int       // the picture it opens on
	Pics      []Picture // one per picture; Pixels nil where it has none (W, H in pixels, the DAX ref as loaded)
	Labels    []string  // the pictures' names (saved characters)
	Classes   []string  // ... their classes ("": the files are gone)
	Portraits [][]byte  // ... their PNG portraits' bytes (nil: none; the DAX head is among Pics)
	Heads     []string  // ... their pool portrait heads ("": none picked)
	Ask       string    // the question before a pick (%s: the label), "" none
}

// IconEditor is the combat icon designer: the icon built so far and the
// colours its six parts take. Colours name palette entries (Looks), in the
// same index space as Picture.Pixels.
type IconEditor struct {
	Large, Custom bool       // Custom: the poses are a dropped PNG (IconPNG), no parts or colours
	Poses         [2]Picture // ready and action poses (Pixels nil: none)
	IconPNG       []byte     // Custom: the PNG's bytes (gobox cuts its poses)
	Colour        [6][2]int  // each part's two colour numbers (0..15)
	Looks         [16]byte   // the palette entry each colour number shows as
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
	Square    *rules.MapPos // the square the 3D view stands on where X, Y is another map's (the overland); nil: X, Y
	Facing    int           // 0 N, 2 E, 4 S, 6 W
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
	HP, MaxHP, AC, THAC0, Move, Age      int            // Age: years (the sheet's)
	Exp                                  uint32
	Items                                []Item // in the rules' order; gobox keeps it
	Spells                               []SpellRef
	Portrait                             Picture
	Dropped                              []byte         // the member's own PNG portrait's bytes (dropped on the sheet), drawn in place of the pool head; nil none
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
	Powers                               []PowerOffer   // the member\'s special powers menu
	LevelExp, NextExp                    int            // the experience the level took and the next takes (0: none)
	SpellsReady, Memo                    string         // spells ready over the slots ("0/3"), in camp too
	Effects                              []Line         // the sheet's list of spell effects on the member, as the game words it
	Page                                 *SheetPage     // the member's sheet as the game's own sheet mode shows it, kept current in every state; gobox's look-only sheet (a fight's or shop's View) draws it, nil: built from the Member (MemberPage)
	Foe                                  bool           // on the monsters' side (gobox colours its name as such); last, for parity's unsetLate
	SheetPowers                          []Entry        // the sheet's power offers (UsePower)
	Reactions                            []Reaction     // the sheet's reactions (SetPolicy)
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
	Handle                      rules.ItemHandle // names it in the item verbs (play.Drop's Item ...); NoItem never
	Name                        string
	Readied, Identified, Cursed bool
	Charges, Price, Weight      int
	Type                        int            // item type id (picks the icon)
	Base                        rules.ItemID   // its base kind (picks the icon; 0 none)
	Kind                        string         // "weapon", "armor", "ring", ...
	Count                       int            // stack size, 0 or 1 for one
	Slot                        string         // readied in: weapon, shield, armor, helm, belt, robe, cloak, boots, gauntlets, ring, ring2, arrows, bolts, missiles, launcher; "" none
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

// ByPosition gives each item the handle of its place in the list (the
// first 1), and Position is the place a handle so given names.
// ponytail: the adapter of the games whose items have no handle of their
// own (cotab, sotsb, pod, buckrogers); it goes when they switch to pool.
func ByPosition(items []Item) []Item {
	for i := range items {
		items[i].Handle = PositionHandle(i)
	}
	return items
}

// PositionHandle is the handle ByPosition gives place i.
func PositionHandle(i int) rules.ItemHandle { return rules.ItemHandle(i + 1) }

// Position is the place PositionHandle gave h (-1 for NoItem).
func Position(h rules.ItemHandle) int { return int(h) - 1 }

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
	Size      rules.Size
	Known     []rules.MapPos // squares seen; empty: nothing is known
	Marked    []rules.MapPos
	Notes     []string       // the player's text for each Marked square
	NoteKinds []string       // a Label.Kind for each Marked square, \"\" none
	Walls     []Wall         // by square
	Doors     []Wall         // State says what each door is
	Walked    []rules.MapPos // squares walked on (Known also holds those only seen)
	Labels    []Label        // the map's own labels; the player's are Marked/Notes
	Marker    string         // the party marker's art prefix on this map (Data.Art "<Marker>/<facing>"); "" is "marker"
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
	At   rules.MapPos
	Text string
	Kind string // "shop", "temple", "inn", "civic", ...
}

// Map is the area the party walks, as the game knows it; gobox moves the
// party on it and tells the game with Arrive and Face. A game gives it as
// play.Game's Map.
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
	At         rules.MapPos
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
	At   rules.BoardPos
	Free bool
}

// Reach is a square the unit can walk to and the steps it costs.
type Reach struct {
	At    rules.BoardPos
	Steps int
	Path  []Step // the walk there from the unit's own square
}

// Ground is one battlefield square.
type Ground struct {
	At      rules.BoardPos
	Art     Picture
	Block   bool
	Height  int     // the square's height (a wall, a tree)
	Tile    int     // the tile number (Combat.Tiles)
	Crown   bool    // the top of a two-square thing: it stands on the square south's picture
	Overlay Picture // drawn on the ground over the tile, under the units (a cloud, a body); none: no File, Name or Pixels
}

// Unit is a combatant.
type Unit struct {
	ID         int
	Name       string
	At         rules.BoardPos
	Facing     int
	Foe        bool
	HP, MaxHP  int
	Art        Picture
	Dead, Gone bool
	Size       int              // footprint edge in squares
	Squares    []rules.BoardPos // squares covered
	AC         int
	Status     string
	Dmg        string // the unit's damage, e.g. "1d8+6"
	Member     int    // party index of a party unit, -1 for a monster
	Active     bool   // it is the unit the player acts for
	Hit        int    // the active unit's chance in percent to hit it with one blow, 0 not a target
	Pose       int    // its icon's pose: 0 ready, 1 action
	IconPNG    []byte // a member's own PNG icon's bytes: gobox cuts the pose and faces it by Facing (nil: composed from its icon picks)
}

// Weapon is a Turn's weapon: its icon (Type, Words) and tooltip.
type Weapon struct {
	Name        string
	Type        int
	Base        rules.ItemID // its base kind (picks the icon)
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

// Shop, Service, Training, Loot, Door, Vault: the place an area script has
// opened (Mode says which); nil when none is open.
// Offer is a gem or jewel of the purse to appraise (Appraise), by the member:
// Value 0, not valued yet; else the one valued, to keep or sell.
type Offer struct {
	Member  int
	Offer   OfferKind // what is offered (Appraise's Offer)
	Kind    string    // Gems or Jewelry
	Value   int
	CanKeep bool // Appraise.Keep is allowed
}

// OfferKind is what an Offer values: the purse's gem or piece of jewelry.
type OfferKind int

const (
	OfferGem OfferKind = iota
	OfferJewel
)

// ItemRef names one member's item.
type ItemRef struct {
	Member int
	Item   rules.ItemHandle
}

type Shop struct {
	Offers        []Offer
	Name          string
	Goods         []Item
	Buys          bool           // the shop buys from the party
	Reloads       bool           // the shop refills clips (Reload)
	Pooled        map[string]int // money pooled with the shop, by coin (nil: none)
	IdentifyPrice int
	Unknown       []ItemRef // the party's unidentified items the shop can identify
	Identified    *ItemRef  // the item last identified here, nil none
	Identifies    int       // identifications done here so far
}

type Service struct {
	Name     string
	Price    int
	Needs    []int          // party indices the service would help; the rest have no need of it
	Not      map[int]string // party index to why the service is of no use ("is not blind.")
	Declines bool           // the temple only says why, and does not offer to cast it anyway
	Only     bool           // it cannot be chosen for one it would not help (the wound cures)
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
	Minimums     map[string]int // the chosen class's lowest ability scores, by Member.Stats key ("str", ...)
	// Stats are the abilities the sheet shows, in the game's order and
	// count (nil: STR INT WIS DEX CON CHA from Draft.Stats, with 18/xx).
	Stats []StatDef `json:",omitempty"`
	// Derived are the figures the sheet shows besides the abilities
	// (age, AC, THAC0, ...), in the game's order. When set they are the only
	// source: a figure not given is not shown. nil: the figures come from
	// Draft.
	Derived []Derived `json:",omitempty"`
}

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
	Overwrite  string // the game's own words for saving over it ("": "Overwrite NAME?")
}

// Confirm is one of gobox's yes/no confirms as the game words it: Asks
// are asked one after another (each Yes goes on, any No stops), each
// line a "\n"-split text with %s the item's or member's name; Yes: they
// start on Yes (else No).
type Confirm struct {
	Asks []string
	Yes  bool
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

// Line is one row of a PickItem list.
type Line struct {
	Text     string
	Header   bool  // a heading, not pickable
	Dim, Off bool  // shown dimmed; not pickable
	Spell    int   // the spell the row names, 0 if none
	Hides    []int // Ask.Options hidden while this row is chosen
}

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
	EvEngaged // a fight began: Combat is its board before the first turn (the step's events play from it)
)

// Data is declarative overrides only; no drawing, no positions (the
// exceptions are the package doc's). A game's
// rules package exports it as func Data() Data, and its cmd hands gobox
// the rules and the data together (ui.Launch).
type Data struct {
	Commands    []PartyCommand // the party menu's extra entries; play.Assembling.Commands[i] does entry i
	SaveDir     string         `json:"-"` // where gobox writes slot files ("": UserConfigDir/gobox/<ID>/saves)
	Series      string
	Coin        string            `json:",omitempty"` // the money Purse, Sale and Pooled count in (Credits; "": Gold)
	ID          string            // game id
	VerbKeys    map[string]rune   `json:",omitempty"` // the series' hotkeys of combat verbs by label ("Aid": 'H'), over gobox's
	VerbChords  map[string]string `json:",omitempty"` // more keys of a combat verb by label, comma-separated chords such as alt+q,alt+a or space or shift+u
	VerbOrder   []string          `json:",omitempty"` // the series' order of the turn's verbs by label (POR: Move View Aim Use Cast Turn Quick Done); those it leaves out follow in gobox's
	PartyMenu   []MenuKey         `json:",omitempty"` // the series' party menu: gobox's function labels in its order with its hotkeys (nil: gobox's)
	Texts       map[string]string
	Confirms    map[string]Confirm `json:",omitempty"` // gobox's own confirms in the original's words, by act: "Drop", "DropScribe", "Sell", "SellMany", "SellFree", "SellFreeMany", "DropMember" (nil: gobox's)
	Credits     []string           `json:"-"`
	SpellRules  map[int]string     // what each spell does, by spell id (the spell tip text)
	Art         map[string]Picture `json:"-"`
	Sounds      map[string]string
	Palette     []byte                        // palette bytes
	MenuItems   []MenuItem                    // the game's own main menu lines, after gobox's
	MusicPlaces []string                      // every Where.Music trigger the game sets (gobox makes their data/music folders)
	OwnMusic    string                        `json:"-"` // the player's own music folder (the host's data/music) for a game with Adventure set ("": beside Files)
	Imports     []string                      // other games' rosters and saves this game reads, by game id
	QuickParty  []QuickPick                   // the party menu's Quick Start party (every game offers it while the party is empty)
	Clues       []Clue                        `json:"-"` // the clue book's chapters (nil: the game has none)
	Scans       func(name string) image.Image `json:"-"` // the picture a Read's line "\x01name" shows (nil: none)
	Border      *Border                       // the window border's pieces in the game's DAX symbols (nil: gobox's default)

	DaxMask      map[string]int `json:",omitempty"` // a DAX file's transparent colour index, as the original's loader takes it (File, e.g. "CHARS", "CPIC3"; no entry: none)
	Files        string         `json:"-"`          // the game's DOS files: the DAX archives Picture File and Block, Border and AdLib name ("": none)
	Adventure    string         `json:"-"`          // the imported adventure folder (gobox import): Picture File and Block from its art/, AmigaMusic's and AdLib's files from its sound/ (SoundDir), the sequences' animations from its anim/ (AnimDir); set, Files is not read ("": none)
	SoundTables  []byte         `json:"-"`          // the sound driver's tables (gobox/audio NewSynth's data; nil: none)
	AdLib        *AdLib         // the AdLib music and effects (nil: none)
	AmigaMusic   *AmigaMusic    // the Amiga's Music-X songs and GLIB effects, played instead of the DOS sound (nil: none)
	DungeonTiles int            // the tiles at the start of the combat set
	Rests        []RestKind     // camp's rests, in the menu's order (nil: none)
	IconParts    []string       // the combat icon designer's parts
	Bindings     []Binding      // the game's own help rows, after the Any time Ctrl-F row
}

// Binding is one row of the help: in context Ctx, Keys (and Mouse, Pad) do Desc.
type Binding struct {
	Ctx, Desc        string
	Keys, Mouse, Pad string
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

// Border is the window border as frames of block Block of DAX file File:
// Pieces by place (top-left, top-right, bottom-left, bottom-right, top,
// bottom, left, right; below 0: none), Caps the 4 end caps, Mask the colour
// index left transparent (nil: none), and Edges an edge's cells one by one
// ("top", "bottom", "rule", "left", "right", "divider"; see gobox's present).
// Sets, when set, are the frame sets' blocks by Common.SymbolSet (pod's
// BORDERS 0..3; Block is set 0's); Amiga is the Amiga release's frame, the
// border under Amiga art (nil: the Amiga shows this one). MarkShift, when
// set, is the selected member's card's border: every piece's and edge's
// frame plus MarkShift (cotab's yellow 3D-view box), not the others' grey.
type Border struct {
	File   string
	Block  int
	Pieces [8]int
	Caps   []int
	Mask   *uint8
	Edges  map[string][]int
	Sets   []int       `json:",omitempty"`
	Amiga  *AmigaFrame `json:",omitempty"`

	MarkShift int `json:",omitempty"`
}

// AmigaFrame is the Amiga release's window frame as its program draws it:
// Screens are the frame screens (each the pieces it blits, in order, on a
// 320x200 screen), and every border cell is the 8x8 cell of a screen at
// the place gobox draws it (Pieces, Caps and Edges as Border's, by cell).
// Sets are the frame sets' blocks by Common.SymbolSet, as Border.Sets: a
// blit of block Sets[0] draws from the set's block.
type AmigaFrame struct {
	File    string
	Screens [][]FrameBlit
	Pieces  [8]FrameCell
	Edges   map[string][]FrameCell
	Mask    uint8 // the colour a Masked blit leaves clear (amigaimport Gfx.Mask)
	Sets    []int `json:",omitempty"`
}

// FrameBlit is one piece the Amiga program draws: frame Frame of block
// Block of the AmigaFrame's File, its top-left pixel at X, Y (the call's
// position less the piece's hot spot), through its mask when Masked.
type FrameBlit struct {
	Block, Frame, X, Y int
	Masked             bool `json:",omitempty"`
}

// FrameCell is the 8x8 cell at column Col, row Row of screen Screen,
// upside down when FlipY (a top edge's cell as a bottom edge's).
type FrameCell struct {
	Screen, Col, Row int
	FlipY            bool `json:",omitempty"`
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
// end, as a Read's Texts.
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

// ScreenKind is a whole screen a game asks gobox to put up
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
	At rules.BoardPos
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

// MenuItem is a main menu line of the game's: gobox plays Sequence (an
// adventure.SeqIntro or SeqEnding animation) full screen and comes back to the menu; no game code runs.
type MenuItem struct {
	Label    string
	Key      rune
	Sequence string
}

// PowerOffer is one of a member's special powers on the inventory.
type PowerOffer struct {
	Label string
	Kind  PowerKind
}

// PowerKind is what a member's power does (UsePower.Kind).
type PowerKind int

const (
	PowerCast    PowerKind = iota + 1 // cast a memorised spell
	PowerDisplay                      // list the spell effects on the member
	PowerHeal                         // a paladin lays on hands
	PowerCure                         // a paladin cures disease
)

// The pickers a game opens (OpenPicker.Kind).
const (
	PickHead     = iota // the portrait's head (creation sheet)
	PickBody            // its body
	PickIconHead        // the combat icon's head (icon editor)
	PickWeapon          // its weapon: the icon's body (CBODY)
	PickAdd             // the saved characters to add to the party (party menu)
	PickForget          // the saved characters to forget (party menu)
)

// PartyCommand is a party menu entry of the game's own.
type PartyCommand struct {
	Label string
	Key   rune
	From  string `json:",omitempty"` // Add from another game: its id (Data.Imports)
}

// SoundDir is where the game's sound files are (AdLib, AmigaMusic, the
// animations' songs): the adventure's sound/, else Files ("": none).
func (d Data) SoundDir() string {
	if d.Adventure != "" {
		return filepath.Join(d.Adventure, "sound")
	}
	return d.Files
}

// AnimDir is where Sequences' animation files are: the adventure's anim/,
// else Files ("": none).
func (d Data) AnimDir() string {
	if d.Adventure != "" {
		return filepath.Join(d.Adventure, "anim")
	}
	return d.Files
}
