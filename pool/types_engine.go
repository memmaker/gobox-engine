package pool

import (
	"gobox/core"
	"gobox/game"
	"image"
	"image/color"
	"strings"
)

// CellRect is a rectangle of 8x8 text cells, both corners included. For a
// box it is the inside, as DrawBorder takes it: the border is the ring of
// cells around it (Framed).
type CellRect struct{ X1, Y1, X2, Y2 int }

// PaneRun is text in one colour on a pane line from column X of the pane
// (0: its left edge); Upcase picks the glyphs as DisplayChar does.
type PaneRun struct {
	X      int
	Text   string
	Fg, Bg color.RGBA
	Upcase bool
}

// Pane is the text under the 3D view (TextPane): what DisplayTextBlock,
// the messages and the debug values put there, a line per row.
type Pane struct {
	Lines [][]PaneRun
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
	Stats              [6]int // STR .. CHA
	Str00              string // the exceptional strength after STR: "(78)", ""
	Creating           bool   // the creation's sheet: no level, experience, party items, items or status
	Levels             string // "1/2"
	Exp                int
	Next               int // the experience the next level takes (0: none)
	AC, HP             int // AC as displayed (10 down to negative)
	MaxHP              int // HP below it shows yellow
	THAC0              int
	Damage             string
	Move               int
	Saves              [5]int        // extra: the saving throw targets, paralysis, petrification, rod, breath, spell (0: not shown)
	Mods               [6]int        // extra: the rule set's ability modifiers, STR .. CHA (shown with Rules)
	Rules              []core.Figure // extra: the rule set's own figures (CombatRules.Sheet; nil: none, and no Mods)
	Status             string
	Weapon, Armour     string         // extra: the readied weapon and armour ("": none shown; POR shows them on its inventory)
	Encumbrance, Coins string         // extra: the load carried and the coins held, as the game writes them ("": not shown)
	Effects            string         // the spell effects, as many as fit the row
	PortraitAt         CellRect       // the portrait's box (zero: no portrait)
	Portrait           image.Image    // a dropped PNG in the box, in place of the DAX portrait (nil: none)
	Icons              [2]image.Image // the creation's combat icon from a PNG, ready and action poses, in View.IconBox (nil: the DAX icon among the pictures)
}

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

// Beast is a bestiary entry.
type Beast struct {
	Name            string
	Danger          int // I..VIII
	Text            string
	Icon            image.Image // the combat icon (nil: none)
	Pic             image.Image // the picture window's when first met (nil: none)
	IconPix, PicPix Picture     // Icon and Pic as the game's pixels in View.Ink (the presenter colours them)
}

// Sheet is what the character sheet offers besides the page.
type Sheet struct {
	Powers    []Power         // Spells, Magic, Heal, Cure
	Reactions []core.Reaction // extra: the rule set's reactions and the member's policy for each (SetPolicy)
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

// Window is what the view window shows (game.Window).
type Window = game.Window

const (
	WinThreeD  = game.WindowScene
	WinAreaMap = game.WindowMap
	WinPicture = game.WindowPicture
	WinText    = game.WindowText
)

// Harm is a member's line on the after-combat page: its name and what
// harms it.
type Harm = game.Harm

// Pixels is the rectangle in screen pixels.
func (r CellRect) Pixels() image.Rectangle {
	return image.Rect(r.X1*8, r.Y1*8, r.X2*8+8, r.Y2*8+8)
}

// Framed is the box with its border.
func (r CellRect) Framed() CellRect { return CellRect{r.X1 - 1, r.Y1 - 1, r.X2 + 1, r.Y2 + 1} }

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

// Picture is one thing drawn on the page, bottom first: frame Frame of
// block Block of the DAX file File (a picture, an animation's frame, a
// symbol, a tile) decoded in the page's palette with its mask pixels
// transparent, its top left at At.Min and cut to At; or, Fill set (the
// presenter's own layouts; the engine draws none), a fill of Colour over At.
type Picture struct {
	File         string // "" for a fill (the presenter adds its own pictures, the minimap, so)
	Block, Frame int
	Name         string      // a pool portrait name ("head_003"); with File, drawn from the pool when it resolves
	Img          image.Image // nil for a fill, or for a picture front/present decodes from File, Block, Frame
	Mask         *uint8      // with File: this colour index is transparent too, as the mask (nil: only asset.Clear)
	Pix          []uint8     // the game's own pixels, PixW wide, Img nil: front/present converts them in View.Ink (IndexedRGBA)
	PixW         int
	Palette      []color.RGBA // with Pix: Pix are indices into it (a VGA picture; alpha 0 transparent), not EGA; nil: EGA
	At           image.Rectangle
	Fill         bool
	Colour       color.RGBA
}

// PowerKind is what a member's power is (Power.Kind, play.UsePower.Kind).
type PowerKind int

const (
	powerNone PowerKind = iota
	PowerSpells
	PowerMagic
	PowerHeal
	PowerCure
	PowerView
	PowerItems
	PowerPlan
	PowerScribe
	PowerIcon
	PowerDrop
)

// Power is a member's action of the View menu the inventory offers too.
type Power struct {
	Name string
	Kind PowerKind
}

// Spell is one memorised spell verbCast can cast.
type Spell struct {
	ID   int
	Name string
}

// Pending is what the game waits on (View.Pending): a shop or temple a
// suspended ECL script opened, a fight, the finale, the treasure after a
// fight, a locked door met while walking. A rule call answers it (Leave,
// LootLeave, TryDoor) and the game goes on.
type Pending int

const (
	PendNone     Pending = iota
	PendShop             // the shop: Browse, Identify, Appraise, KeepAppraised; Leave
	PendTemple           // the temple: Cure, ViewMember, Appraise, KeepAppraised; Leave
	PendFight            // a fight: CombatStep, TurnChoice
	PendStory            // (unused: POR's finale is the manifest's End; the front's numbering keeps the place)
	PendLoot             // the treasure after a fight (State.Loot, gobox's loot screen): TakeLoot, LootDetect; LootLeave
	PendDoor             // a locked door (View.Door): TryDoor
	PendSequence         // a full-screen animation (State.Sequence): Continue
)

// Quest is a journal entry: Text is what its giver said, paragraph by paragraph.
type Quest struct {
	Name, Giver, Place string
	Text               []string
	Done               bool
}

// Turn is what the acting member may do (the original's combat menu).
type Turn struct {
	Unit      int                 // whose turn
	Can       []TurnVerb          // the actions offered, in the menu's order
	Spells    []Spell             // what verbCast casts
	Maneuvers []core.ManeuverCard // extra: what verbManeuver does (the rule set's; empty under the original's)
	Aim       bool                // aiming: verbAim takes a square, verbPoint previews one, verbClose gives up
	Casting   string              // extra: aiming a spell, its name (a delayed one goes off turns after it was chosen)
	Weapon    *core.ItemCard      // extra: the weapon in hand (its icon on the Attack button), or nil
	Moves     int                 // squares of move left
}

// PowerPick is the pick screen: Name chooses What, one of Options (PickPower).
type PowerPick struct {
	Name, What string
	Options    core.ManeuverCards
}

// Strike is one weapon attack, for views that show it (the original only
// prints it): Seq counts attacks, so a new one is a new Seq.
type Strike struct {
	Seq      int
	From, To int // unit IDs (0: off the map)
	Hit      bool
	Damage   int
}

// DoorOpts is a locked door's options besides Exit (View.Door).
type DoorOpts struct{ Bash, Pick, Knock bool }

// DoorWay is what TryDoor does with a locked door.
type DoorWay int

const (
	DoorExit  DoorWay = iota // leave it
	DoorBash                 // force it
	DoorPick                 // pick the lock
	DoorKnock                // cast Knock
)

// Battle is the combat map as the remaster's 3D view needs it.
type Battle struct {
	Ground    [25][50]core.Square
	Tiles     []image.Image // the tile set by tile number, in combat's colours
	TilePix   []Picture     // Tiles as the game's pixels in View.Ink, or DAX refs (File, Block, Frame), the presenter's to make into Tiles
	IconPix   []Picture     // Units' icons so, by unit (combat.Unit mirrors Unit, so they are here); used where Icon is nil
	EffectPix []Picture     // Effects' pictures so, by effect; used where Pic is nil
	Floor     color.RGBA    // the ground's colour
	Dungeon   bool          // tiles 0..DungeonTiles-1 are stone wall pieces (else outdoor things: trees, water ...)
	Units     []Unit
	Cursor    image.Point                 // the square the cursor was last drawn on
	CursorOn  bool                        // the cursor icon is up (aiming, looking)
	CursorAt  []image.Point               // the squares the cursor covers (a large combatant's footprint)
	Marks     []image.Point               // squares framed for a choice (reach of a move, an area of effect)
	Reach     []image.Point               // squares a click walks to from the main menu (the active unit's move)
	Paths     map[image.Point][]core.Step // the walk to each Reach square, from the unit's own square
	Targets   []int                       // units (by ID) verbAttack takes on the turn; aiming: the foes Tab cycles through
	Turn      *Turn                       // the player's turn waits for an action (nil: no one to act for). The game fills it on every player's turn, even one whose only offer is Can: the presenter builds a Turn itself only for an AskSquare or AskMove question (askedBattle), since pace, the bar, the Cast list and the status row all read this one field
	Panel     int                         // the unit the status row shows (who acts, or is being hit)
	Log       []string                    // this battle's messages, oldest first
	Area      bool                        // Marks are a spell's area (else a move's reach or a choice)
	AreaAt    []image.Point               // the squares a spell's area covers round the aimed square
	Skipped   []image.Point               // squares of allies in the area the spell passes over (Bless: an enemy next to them)
	AimWhy    string                      // aiming: why the aimed square cannot be taken ("" when it can); views draw the aim red
	Effects   []Effect                    // sprites over the map: a missile, a hit, a fall
	View      image.Point                 // top left square of the scrolled view window the original showed (what the camera follows)
	Strike    Strike                      // the last weapon attack
}

// Unit is one combatant on the map.
type Unit struct {
	ID         int // its place on the combat list: the same unit from snapshot to snapshot
	X, Y, Size int
	Squares    []image.Point // the squares it stands on (a large combatant's footprint)
	Party      bool
	Facing     int         // 0..7
	Icon       image.Image // the combat icon, facing and pose (striking) as the original draws it
	IconSize   image.Point // the built icon's size in pixels, 24 a square: what the figure's size and footprint follow (a PNG icon keeps it at its own resolution)
	Active     bool        // whose turn it is
	Name       string
	HP, AC     int
	MaxHP      int
	Hit        int    // the active member's chance in percent to hit it with one blow (0: not a target)
	Dmg        string // the readied attack's damage ("1d8+1")
	Status     string // "" when it can act, else why not ("Helpless", "Unconscious" ...)
}

// Effect is a sprite the original draws over the map for a moment.
type Effect struct {
	X, Y float64     // its top left corner, in squares (missiles move in thirds)
	Pic  image.Image // in combat's colours
}

// Inventory is the inventory screen (ItemsMenu) as the remastered one
// draws it. Its items' Index is what play's Item names.
type Inventory struct {
	Member int              // index into View.Party
	Doll   []*core.ItemCard // the readied items by doll slot (the game's slots, nil: empty)
	Pool   []core.ItemCard  // the party inventory
	Cap    int              // the party inventory's slots
	Use    bool             // items can be used here (UseItem)
	Sell   bool             // a shop buys them (Sell)
	Switch bool             // partyView shows another member (not in combat)
	Powers []Power          // what the member can do besides items (the View menu's)
	Wares  []core.ItemCard  // the shop's goods (Browse; nil elsewhere): Index is what Buy takes, Price what it costs
	Loot   bool             // Wares are a fight's treasure (the Loot screen): Buy takes one for nothing
	Gold   int              // the party's money, with Wares
	Copper int
	Trade  bool // the menu offers Trade
	Halve  bool // ... Halve
	Join   bool // ... Join
}

// Loot is the treasure screen's options besides View and Exit (View.Loot).
type Loot struct {
	Take   bool // there is treasure to take (LootTake)
	Detect bool // the selected member has Detect Magic memorized (LootDetect)
	Share  bool // there is money to divide among the party (LootShare)
}

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

// PickHead .. PickForget are the pickers OpenPicker readies.
const (
	PickHead     = iota // the portrait's head (creation sheet)
	PickBody            // its body
	PickIconHead        // the combat icon's head (icon editor)
	PickWeapon          // its weapon: the icon's body (CBODY)
	PickAdd             // the saved characters to add to the party (party menu)
	PickForget          // the saved characters to forget (party menu)
	PickAddFrom         // another game's saved characters to add to the party (party menu; Engine.addFrom)
)

// Lore and Read are the game's.
type (
	Lore = game.Lore
	Read = game.Read
)

// Sound is the sound driver the rules call synchronously (gobox/audio.Driver;
// a fake in tests). Cards: 0 Tandy, 1 speaker, 2 none.
type Sound interface {
	Effect(n int, quiet bool) // PlaySound n; quiet: hook only, start nothing
	CutAfter(frames int)      // stop and unhook after frames output frames
	Wait(n int)               // 0 hooks, n > 0 the cut mode
	Reset()
	SetCard(c int)
	Song(n, next int, again bool) bool // soundtrack song n then next; false: no soundtrack
	AmigaSong(n int)
	SetSoundtrack(on bool, n, music int)
	Track() (on bool, song int) // the soundtrack plays, its song
	Sounding() bool             // the Amiga players sound
}

// TurnVerb is a decision of a member's turn in a fight (Turn.Can, turnAct).
type TurnVerb int

const (
	VerbNone TurnVerb = iota
	VerbWalk
	VerbAttack
	VerbManeuver
	VerbCast
	VerbItems
	VerbTurn
	VerbQuick
	VerbGuard
	VerbBandage
	VerbWait
	VerbEnd
	VerbFlee
	VerbPoint // the aim: preview a square
	VerbAim   // the aim: take a square
	VerbClose // the aim: give up
)

// TurnAct is one decision of the turn: Verb with its unit, spell or
// maneuver N, or its square X, Y. It stays data so a question can make it again.
type TurnAct struct {
	Verb TurnVerb
	N    int
	X, Y int
}
