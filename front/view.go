package front

import (
	"gobox/core"
	"gobox/game"
	"image"
	"image/color"
	"strings"
	"time"
)

// View and its parts: the game as the player sees it, published by the game
// each time it waits; the presenter reads only this. Nothing in it points
// back into the game's state.

// AreaMap (extra) is the whole area map for a screen of its own: every
// square's picture (Squares[y][x], nil: none drawn), Tile pixels square,
// and the party's square and marker. The presenter builds the map's picture
// and keeps it while no square's picture changes.
type AreaMap struct {
	Squares [][]image.Image
	Tile    int
	Party   image.Point // the party's map square
	Marker  image.Image // the party's icon, facing its way
	Seen    [][]bool    // Seen[y][x]: the party has seen that square (nil: all seen); unseen squares are drawn as none

	SquarePix [][]Picture // Squares as the game's pixels (Picture.Pix), the presenter's to colour into Squares
	MarkerPix Picture     // Marker so, or a DAX ref
}

// Aim is what a maneuver aims at.
type Aim int

const (
	AimWeapon Aim = iota // a foe, with c's next weapon attack (the bar's Attack)
	AimSelf              // no aim
	AimFoe               // a foe within Range
	AimAlly              // a member of c's side within Range (c too)
	AimArea              // a square within Range: who stands within Area squares of it (Aimed, Burst)
)

// Maneuver is one of the bar's Maneuver list.
type Maneuver struct {
	ID    int
	Name  string
	Tip   string // what it does, one line a sentence
	Aim   Aim
	Range int // squares, AimFoe, AimAlly and AimArea
	Area  int // AimArea: the burst's squares round the square aimed at
	Move  int // AimWeapon: squares the member may move first; the player aims at the square (Aimed; its own: none)
}

// Read is journal text the game shows for the player to read (View.Read):
// the presenter pages it (JournalPage) and answers ActOK, N the page it was
// left on; the script waits until then. It is the game's own (State.Read),
// the same pointer every View while the same text is up.
type Read = game.Read

// Money is core.Money: an amount the front end shows (as "12g 50c").
type Money = core.Money

// Power is a member's action of the View menu the inventory offers too.
type Power struct {
	Name string
	Kind ActKind
}

// InvItem is one item of the inventory.
type InvItem struct {
	Item    core.ItemHandle // names it (Action.Item)
	Name    string          // as the list shows it
	Type    byte            // base item type
	Words   []string        // the item's name words, known or not (for its icon)
	Count   int
	Readied bool     // it is readied (shown with present.Data.ReadyMark)
	Slot    bool     // it has a doll slot (it can be readied)
	Wrong   bool     // the member's class cannot ready it
	Use     bool     // it has a use (a scroll, a spell of its own)
	Look    []string // what it does (the shop's Look text)
	Price   Money    // what selling it pays (a ware: what it costs; loot: none)
	Scribe  bool     // a scroll someone was going to scribe from (the front end asks before it goes)
	Kind    ItemKind // its party inventory tab
	Rules   []string // extra: how it attacks in combat (the bar's tooltip), set on the Turn's Weapon and Missile
}

// ItemKind sorts the party inventory into tabs (extra).
type ItemKind int

const (
	KindMisc ItemKind = iota
	KindWeapon
	KindArmor
)

// PowerPick is the pick screen: Name chooses What, one of Options (PickPower).
type PowerPick struct {
	Name, What string
	Options    []Maneuver
}

// Policy (extra) is a member's standing answer to one of the rule set's
// reactions, chosen beforehand on the sheet: nothing asks mid-turn.
type Policy uint8

const (
	Always Policy = iota // the zero value: an older save's, a foe's
	Never
	OnHit // only when it turns a hit (the rule set says which way)
)

func (p Policy) String() string { return [...]string{"Always", "Never", "On hit"}[p%3] }

// Reaction is one of the rule set's reactions c has, with c's policy.
type Reaction struct {
	ID     int
	Name   string
	Policy Policy
}

// Walls is a 3D view the presenter builds from the rules' own map
// (View.Walls), seen from View.X, Y, Facing: a wall face on every side
// Side[y][x][d] (N E S W) that holds a wall value, textured with that
// value's picture.
type Walls struct {
	Key                 string              // what the scene is built from (map, wall sets): a new Key builds it again
	Side                [][][4]int          // a square's wall value by side, rows from the top (0: open)
	Fronts              map[int]image.Image // a wall value's picture: the original's front view of it
	FrontPix            map[int]Picture     // Fronts as the game's pixels (Picture.Pix, in View.Ink) or DAX refs (File, Block, Frame), the presenter's to make into Fronts
	Pool                map[int]string      // a wall value's remastered texture by gobox/walls pool name, in place of Fronts
	Floor, Sky          string              // the flats' pool names (floor_NAME, ceiling_NAME); outside (not View.Inside) the sky shows
	FloorRGBA, CeilRGBA *color.RGBA         // plain flat colours (nil: none), used where Floor or Sky is "": the pool's floor and ceiling are not
	Sun                 bool                // by day outside: the sun up, at SunAz
	SunAz, SunEl        float64             // the sun's bearing (compass radians, 0 north, pi/2 east) and height above the horizon
	Mountains           bool                // by day outside: the mountain range on the horizon
	Backdrop            *Picture            // over the whole view behind the walls, opaque and stretched, in place of the flats (nil: none): Img, or the game's pixels or DAX ref the presenter decodes into Img
}

// Dir is a facing: N, E, S, W.
type Dir int

const (
	North Dir = iota
	East
	South
	West
)

// AreaID is the GEO block a 3D area is drawn from (the same area in every
// disk's GEO file that has it).
type AreaID int

// Window is what the view window (cells ViewWindow) shows.
type Window int

const (
	WinThreeD  Window = iota // the 3D view of Area at X, Y, Facing
	WinAreaMap               // the wilderness or city map
	WinPicture               // a picture, portrait or approaching sprite
	WinText                  // text in the whole view window (opClearBox)
)

// Member is one party member as the party list shows them.
type Member struct {
	Name     string
	HP       int
	MaxHP    int
	AC       int         // as displayed: 10 down to negative
	Status   string      // as the sheet names it ("Okay", "Unconscious" ...)
	Class    string      // as the sheet names it ("Fighter/Thief" ...)
	Selected bool        // the selected member, or the candidate while a member is picked (SelectCharacter)
	Colour   int         // the roster's EGA colour for the name when not Selected: red not okay, yellow a monster, magenta ready for a level, else cyan
	Head     image.Image // the paper-doll head the player chose, in EGA colours (nil if missing)
	HeadPix  Picture     // Head as the game's pixels in the EGA's own colours, not View.Ink (the presenter colours it)
	Portrait image.Image // extra: the PNG the player dropped on the sheet, in place of Head (nil: none)
	Powers   []Power     // extra: what the member can do now out of combat (a card's context menu): View, Items, then powers'; Kind with the member's index acts
	Exp      int         // extra: experience, what the level took (LevelExp) and what the next takes (NextExp; 0: it can rise no further)
	LevelExp int
	NextExp  int
	Spells   string // extra: a spellcaster's spells ready over its slots ("0/3", the overview); "" otherwise
	Memo     string // extra: in camp a spellcaster's spells ready over its slots ("0/3"); "" otherwise
}

// Pane is the text under the 3D view (TextPane): what DisplayTextBlock,
// the messages and the debug values put there, a line per row.
type Pane struct {
	Lines [][]PaneRun
	Rows  int // >0: the presenter shows Rows lines at a time and a key brings the next lot (0: all)
}

// PaneRun is text in one colour on a pane line from column X of the pane
// (0: its left edge); Upcase picks the glyphs as DisplayChar does.
type PaneRun struct {
	X      int
	Text   string
	Fg, Bg color.RGBA
	Upcase bool
	Blink  bool // the presenter alternates Fg with the highlight colour a few times (blink.go)
}

// MapSquares is a map as the maps show it (minimap, View.Map, the full
// map): squares [y][x], the party's square (off the map: not shown) and
// the step towards its facing.
type MapSquares struct {
	Squares [16][16]MapSquare
	Party   image.Point
	Ahead   image.Point
}

// MapSquare is a square of a map: how far the party has seen it, its
// walls (N E S W) and its label (Letter 0: none).
type MapSquare struct {
	Seen   Seen
	Walls  [4]MapWall
	Kind   int  // the label's kind: its colour is Data.LabelInks[Kind]
	Letter byte // the label's legend letter
}

// Seen is how much of a square the party has seen.
type Seen uint8

const (
	Unseen  Seen = iota
	SeenFar      // seen from afar
	Visited      // walked on
)

// MapWall is a wall on a map.
type MapWall uint8

const (
	NoWall MapWall = iota
	Wall
	Door
	LockedDoor
)

// Shown reports whether the pane holds visible text (runs of blanks, as a
// cleared pane leaves, show nothing).
func (p *Pane) Shown() bool {
	for _, l := range p.Lines {
		for _, r := range l {
			if strings.TrimSpace(r.Text) != "" {
				return true
			}
		}
	}
	return false
}

// Note is a toast for the status line (View.Note): Text, put up at At
// (zero: none).
type Note struct {
	Text string
	At   time.Time
}

// View is one published snapshot.
type View struct {
	Pane       *Pane // the text pane under the 3D view: out adventuring (Adventure), or it holds text (nil otherwise)
	Window     Window
	AreaMap    *AreaMap // with WinAreaMap: the whole map, for a screen of its own
	Area       AreaID
	X, Y       int
	Facing     Dir
	Day        bool       // 6:00 to 20:00, when the scripts show the day sky
	Inside     bool       // the party stands on an inside square (the original's inside sky colour)
	Exploring  bool       // the party explores: the adventuring decisions (ActForward .. ActCastPick) are taken
	Sweeping   bool       // ... at the slow, searching pace (ActSweep): the HUD says so
	Interact   bool       // ... on a square whose event waits for ActInteract: the HUD says so
	CanCast    bool       // ... and the selected member has a spell to cast (ActCastPick)
	Camp       *Camp      // the party is in camp: the camp decisions (ActRest .. ActIconEdit) are taken; ActClose leaves
	Temple     *Temple    // a temple (PendTemple): Cure, ViewMember, Appraise, KeepAppraised, ActLead; Leave
	Shop       *Shop      // a shop (PendShop): Browse (the goods, ActBuy, ActSell), Identify, Appraise, KeepAppraised, ActLead; Leave
	Loot       *Loot      // the treasure after a fight (PendLoot): LootView, LootTake, LootDetect, Lead; LootLeave
	Door       *DoorOpts  // a locked door (PendDoor): TryDoor
	Training   *Training  // the training hall for the selected member (TrainingView, the presenter's): Train, Learn; CloseTraining
	Powers     *PowerPick // a member chooses a power: PickPower
	Ask        *Ask       // a question waits for its answer: menu, list, typed line, member, key, pause (the presenter's to pace) or spell
	Party      []Member
	Walls      *Walls         // the 3D view built from the rules' own map (nil: the remastered map Data.AreaMaps names)
	Note       Note           // the status line's toast, shown on the bottom row a while after At
	Pictures   []Picture      // the screen's pictures, symbols and fills, bottom first
	Adventure  bool           // the party is out in the world: the 3D view (or the overland map) is the screen
	Prompt     []PaneRun      // the prompt row's text (the last row), X the page's column
	ViewText   [][]PaneRun    // text over the view window (WinText: opProtection's runes), rows from 1, X the page's column
	Results    *Results       // the after-combat page (in a box round the window)
	CreditPage int            // which credits page (Credit.Page) while Credits
	Credits    bool           // the title's credits page (Credits)
	CreditsPic bool           // the title step is the game's credits as a picture: the presenter adds its remaster credit line
	Sequence   *Sequence      // a full-screen animation plays (SequenceStep, SequenceDone)
	Intro      *Intro         // the title waits to be played (TitlePicture, TitleCredits, TitleDone)
	Full       bool           // Pictures are one whole-screen picture (the title's): fitted to the window
	Cells      image.Point    // the page's size in text cells (40x25 at least; a larger window gives more, ActResize)
	Ink        [16]color.RGBA // the text colours (EGA 0..15) in the page's palette now (all zero: the EGA's own)
	Read       *Read          // journal text the game waits to have read: the presenter pages it (JournalPage), then ActOK
	Purse      *Purse         // the party's money and party items (nil without a party)
	Ordering   bool           // the camp's Party Order bar is up: a card dragged onto another's place moves the member (ActOrder)
	MainMenu   []ActKind      // the party menu is up: the functions open now (ActCreate .. ActQuit, ActMember, ActDropMember); ActLead selects a member
	Hall       bool           // MainMenu is the training hall's: over the 3D view, Begin leaves the hall
	Screen     *game.Screen   // a game's typed screen is up (nil: none): the presenter draws it from its data, the Ask on its prompt rows
}

// Results is the after-combat page: the hired NPCs hiding their
// shares (hiredShare), else the outcome (showResult), or the party
// destroyed; the presenter lays it out in the window's box.
type Results struct {
	Shares  []string // hiredShare: "X takes and hides his share.", each wrapped from column 5 to 34 on rows 5, 7 ...
	Outcome string   // "The party has won." ... on row 3
	Exp     string   // "Each character receives 120" on row 5, "experience points." on row 7
	Taken   string   // "The party takes ...", wrapped from row 9 ("" none)
	Harmed  []Harm   // the members out of the fight dead or worse, or with harmful effects (extra), from row 12
	End     string   // the party is destroyed: the text under "The END!" (the box's title), wrapped in EndAt
	EndAt   CellRect
}

// Harm is a member's line on the after-combat page: its name and what
// harms it.
type Harm struct{ Name, What string }

// SheetPage is what the character sheet's page shows (CharacterSheet),
// laid out by the presenter, in the colours of
// View.Ink. The DAX portrait is among View.Pictures.
type SheetPage struct {
	Name               string
	NameColour         int    // red when not okay, yellow a monster, else cyan (displayName)
	NPC                bool   // "(NPC)" after the name
	Spells             string // a spell list's title over the name row ("Spells in Memory"); "" none
	Sex, Race          string
	Age                int
	Alignment          string
	Class              string
	Stats              [6]int   // STR .. CHA
	StatRows           []Stat   // the game's own abilities in its order, in place of Stats and Str00 (nil: STR .. CHA)
	Derived            []Stat   // the only figures shown, in order, in place of Age, AC, HP, THAC0, Damage and Move (nil: those)
	FieldRows          []string // the creation's fields one per row from row 3, in place of Sex, Race, Age, Alignment and Class (nil: those)
	Str00              string   // the exceptional strength after STR: "(78)", ""
	Creating           bool     // the creation's sheet: no level, experience, party items, items or status
	Levels             string   // "1/2"
	Exp                int
	Next               int // the experience the next level takes (0: none)
	AC, HP             int // AC as displayed (10 down to negative): the front one
	RearAC             int // the rear armour class, as displayed
	MaxHP              int // HP below it shows yellow
	THAC0              int
	Damage             string
	Move               int
	Saves              [5]int // extra: the saving throw targets, paralysis, petrification, rod, breath, spell (0: not shown)
	Mods               [6]int // extra: the rule set's ability modifiers, STR .. CHA (shown with Rules)
	Rules              []Stat // extra: the rule set's own figures (CombatRules.Sheet; nil: none, and no Mods)
	Status             string
	Weapon, Armour     string         // extra: the readied weapon and armour ("": none shown; POR shows them on its inventory)
	Encumbrance, Coins string         // extra: the load carried and the coins held, as the game writes them ("": not shown)
	Effects            string         // the spell effects, as many as fit the row
	PortraitAt         CellRect       // the portrait's box (zero: no portrait)
	Portrait           image.Image    // a dropped PNG in the box, in place of the DAX portrait (nil: none)
	Icons              [2]image.Image // the creation's combat icon from a PNG, ready and action poses, in View.IconBox (nil: the DAX icon among the pictures)
}

// Stat is one of a rule set's figures on the sheet: "Fortitude", "13",
// and the tooltip telling what it is and how it is worked out.
type Stat struct {
	Name, Value string
	Help        []string
}

// Purse is the party's money and party inventory, on the overview.
type Purse struct {
	Gold, Copper, Gems, Jewelry int
	Items, Slots                int // the party items held and their room ("12/60")
}

// Sheet is what the character sheet offers besides the page.
type Sheet struct {
	Powers    []Power    // Spells, Magic, Heal, Cure
	Reactions []Reaction // extra: the rule set's reactions and the member's policy for each (SetPolicy)
}

// Temple is a temple being up (presence only: the services are play.Trading's).
type Temple struct{}

// Shop is a shop being up (presence only: its goods and services are play.Trading's).
type Shop struct{}

// Training is the training hall's offer to the selected member.
type Training struct {
	Name    string
	Why     string      // the hall refuses: why
	Becomes []string    // "a level 3 Fighter" (Train trains)
	Price   int         // gold
	Will    string      // the offer's "will become" lines as the game words them ("": Name will become Becomes)
	Cost    string      // the price as the game writes it ("": Price in gold); red while Gold < Price
	Gold    int         // the party's gold
	Learn   []BookSpell // a mage level was gained: the spell for the book (Learn), chosen before leaving
}

// Camp is camp being up; Bestiary is the monster kinds met (the rest of
// camp is play.Camping's).
type Camp struct {
	Bestiary []Beast
}

// BookSpell is a spell with a count, and the slots of its level.
type BookSpell struct {
	ID      int // the spell (ActPlan, ActScribe X)
	Name    string
	Level   string // its group, as "Cleric Level 1"
	Count   int
	Planned int // spells planned of its level
	Slots   int // slots of its level
}

// Beast is a bestiary entry.
type Beast struct {
	Name            string
	Danger          int // I..VIII
	Text            string
	Icon            image.Image // the combat icon (nil: none)
	Pic             image.Image // the picture window's when first met (nil: none)
	IconPix, PicPix Picture     // Icon and Pic as the game's pixels in View.Ink (the presenter colours them)
}

// Pages is help's or the journal's page (View.Pages), in a box round the
// whole page: ActPage -1 and +1 page back and on (the journal: scroll a
// tall picture first), ActClose closes, any other key the next page.
type Pages struct {
	Title       PaneRun     // the box's title, on row 1
	Top         int         // the row of Lines[0]
	Lines       [][]PaneRun // the page's text, a line a row, X the page's column
	Scans       []Scan      // the journal's pictures on the page
	Hint        PaneRun     // on the message row: the page count or the keys
	Page, Pages int
}

// Scan is a journal picture (the file Name in journal/): the part Src of
// Img scaled into the page rectangle At, on paper of colour Paper.
type Scan struct {
	Name    string
	Img     image.Image
	Src, At image.Rectangle
	Paper   color.RGBA
}

// FullMap is the full-screen map (FullMapView), the presenter's: its
// cursor cell's context menu has Walk (WalkTo), Mark and Remove.
type FullMap struct {
	Name     string      // the map's name, top left
	Map      *MapSquares // the squares, the legend's letter on each labelled square
	At       image.Point // the map's top left on the page, in pixels (At, Cell, LegendAt: the presenter lays them out)
	Cell     int         // pixels a square
	Cursor   image.Point // the cursor's square
	Legend   []MapLabel  // the legend's lines, from cell LegendAt down (the game gives them all, the presenter keeps those that fit)
	LegendAt image.Point
	Info     string  // the cursor's square and its label, on the message row
	Label    *string // the label being typed (nil: none), on the prompt row
	Labelled bool    // the cursor's square has a label (Mark is not offered)
	Mine     bool    // ... the player's (Remove is offered)
}

// MapLabel is a line of the full map's legend: the letter on its squares
// (in EGA colour Colour, by kind), its text and its first square.
type MapLabel struct {
	Letter byte
	Text   string
	Kind   int // its colour is Data.LabelInks[Kind]
	First  image.Point
}

// Sequence is the animation State.Sequence names: Anim its files' base name in Dir.
type Sequence struct {
	Name, Anim, Dir string
}
