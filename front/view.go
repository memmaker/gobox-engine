package front

import (
	"gobox/game"
	"image"
	"image/color"
	"strings"
	"time"
)

// View and its parts: the game as the player sees it, published by the game
// each time it waits; the presenter reads only this. Nothing in it points
// back into the game's state.

// Pace is what goes on with time while the engine waits (View.Pace): the
// presenter paces it with the call it names.
type Pace int

const (
	PaceNone  Pace = iota
	PaceFight      // CombatStep (a player's turn waits on TurnChoice instead: Battle.Turn)
	PaceRest       // RestStep; a key asks StopRest
	PaceStory      // StoryStep
	PaceWalk       // WalkStep: the auto-walk; a key or a click stops it
	PaceAnim       // Animate: the picture's frames
)

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

// Battle is the combat map as the remaster's 3D view needs it.
type Battle struct {
	Ground    [25][50]Square
	Tiles     []image.Image // the tile set by tile number, in combat's colours
	FloorImg  image.Image   // the plain floor's picture, tiled on every floor square (nil: Floor's colour)
	TileNames []string      // Tiles' sprite names ("": none); a tile named so may have a 3D model
	TilePix   []Picture     // Tiles as the game's pixels in View.Ink, or DAX refs (File, Block, Frame), the presenter's to make into Tiles
	IconPix   []Picture     // Units' icons so, by unit (combat.Unit mirrors Unit, so they are here); used where Icon is nil
	EffectPix []Picture     // Effects' pictures so, by effect; used where Pic is nil
	Floor     color.RGBA    // the ground's colour
	Dungeon   bool          // tiles 0..DungeonTiles-1 are stone wall pieces (else outdoor things: trees, water ...)
	Units     []Unit
	Cursor    image.Point            // the square the cursor was last drawn on
	CursorOn  bool                   // the cursor icon is up (aiming, looking)
	CursorAt  []image.Point          // the squares the cursor covers (a large combatant's footprint)
	Marks     []image.Point          // squares framed for a choice (reach of a move, an area of effect)
	Reach     []image.Point          // squares a click walks to from the main menu (the active unit's move)
	Paths     map[image.Point][]Step // the walk to each Reach square, from the unit's own square
	Targets   []int                  // units (by ID) ActAttack takes on the turn; aiming: the foes Tab cycles through
	Turn      *Turn                  // the player's turn waits for an action (nil: no one to act for). The game fills it on every player's turn, even one whose only offer is Can: the presenter builds a Turn itself only for an AskSquare or AskMove question (askedBattle), since pace, the bar, the Cast list and the status row all read this one field
	Panel     int                    // the unit the status row shows (who acts, or is being hit)
	Log       []string               // this battle's messages, oldest first
	Area      bool                   // Marks are a spell's area (else a move's reach or a choice)
	AreaAt    []image.Point          // the squares a spell's area covers round the aimed square
	AimArea   int                    // aiming a spell with no AreaAt: the area's radius round the hovered square, washed as a preview (0 none; radius+1)
	Skipped   []image.Point          // squares of allies in the area the spell passes over (Bless: an enemy next to them)
	AimWhy    string                 // aiming: why the aimed square cannot be taken ("" when it can); views draw the aim red
	Effects   []Effect               // sprites over the map: a missile, a hit, a fall
	View      image.Point            // top left square of the scrolled view window the original showed (what the camera follows)
	Strike    Strike                 // the last weapon attack
}

// Turn is what the acting member may do (the original's combat menu).
type Turn struct {
	Unit      int        // whose turn
	Can       []ActKind  // the actions offered, in the menu's order
	Spells    []Spell    // what ActCast casts
	Maneuvers []Maneuver // extra: what ActManeuver does (the rule set's; empty under the original's)
	Aim       bool       // aiming: ActAim takes a square, ActPoint previews one, ActClose gives up
	Casting   string     // extra: aiming a spell, its name (a delayed one goes off turns after it was chosen)
	Weapon    *InvItem   // extra: the weapon in hand (its icon on the Attack button), or nil
	Missile   *InvItem   // extra: the ranged weapon readied (its icon on the Ranged button), or nil
	Moves     int        // squares of move left
}

// Effect is a sprite the original draws over the map for a moment.
type Effect struct {
	X, Y float64     // its top left corner, in squares (missiles move in thirds)
	Pic  image.Image // in combat's colours
}

// Step is one square of a walk; Free: leaving it lets an enemy take a free attack.
type Step struct {
	image.Point
	Free bool
}

// Strike is one weapon attack, for views that show it (the original only
// prints it): Seq counts attacks, so a new one is a new Seq.
type Strike struct {
	Seq      int
	From, To int // unit IDs (0: off the map)
	Hit      bool
	Damage   int
}

// Square is one map square.
type Square struct {
	Tile   int
	Wall   bool // move cost 0xFF
	Height int  // blocking height
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

// EventKind is what a CombatEvent reports.
type EventKind int

const (
	EvFrame   EventKind = iota // a moment of the fight to show for Ms (a missile's step, a blow drawn)
	EvMessage                  // Text went to the Log: shown for Ms to be read
	EvMoved                    // Unit stepped From -> To
	EvAttack                   // Unit attacks Target
	EvHit                      // Unit hits Target for Amount
	EvMiss                     // Unit misses Target
	EvDamage                   // Unit takes Amount damage
	EvHeal                     // Unit regains Amount hit points
	EvSpell                    // Unit casts spell Amount (Text: its name)
	EvMissile                  // a missile flies From -> To
	EvDied                     // Unit fell (off the map)
	EvFled                     // Unit got away
	EvSound                    // sound Amount (Text: the sound pack's event) goes off here (extra: with what it belongs to, PlayEventSound)
	EvSong                     // the Amiga song Amount starts (0 stops), for a front that plays its own Sound
	EvSaves                    // the save screen goes up (game.EvSaves)
)

// CombatEvent is one thing that happened in a fight, in order (View.Events,
// handed out by Apply once). An event held for Ms carries the map and the
// party as they stood (Battle, Party): the presenter shows that while it
// plays the event, the live View when it has played them all.
type CombatEvent struct {
	Kind         EventKind
	Unit, Target int // combatant IDs (Unit.ID; 0: none)
	Amount       int
	From, To     image.Point
	Text         string
	Ms           int
	Key          bool // held until a key ("press any key"), not for Ms
	Battle       *Battle
	Party        []Member
	Scene        *View // outside a fight: the whole page as it stood, shown while the event plays
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

// Lore is a note in the journal: ID is its number for ActNoteRemove (-1: not
// removable), Place where the player read it ("" a journal text).
type Lore struct {
	Title, Place string
	Text         []string
	ID           int
}

// Money is an amount the front end shows (as "12g 50c").
type Money struct{ Gold, Copper int }

// Pending is what the game waits on (View.Pending): a shop or temple a
// suspended ECL script opened, a fight, the finale, the treasure after a
// fight, a locked door met while walking. A rule call answers it (Leave,
// LootLeave, TryDoor) and the game goes on.
type Pending int

const (
	PendNone   Pending = iota
	PendShop           // the shop: Browse, Identify, Appraise, KeepAppraised; Leave
	PendTemple         // the temple: Cure, ViewMember, Appraise, KeepAppraised; Leave
	PendFight          // a fight: CombatStep, TurnChoice
	PendStory          // the finale's pictures: StoryStep
	PendLoot           // the treasure after a fight (View.Loot): LootView, LootTake, LootDetect, Lead; LootLeave
	PendDoor           // a locked door (View.Door): TryDoor
)

// Inventory is the inventory screen (ItemsMenu) as the remastered one
// draws it. Its items' Index is what an Action's N names.
type Inventory struct {
	Member   int        // index into View.Party
	Doll     []*InvItem // the readied items by doll slot (the game's slots, nil: empty)
	Pool     []InvItem  // the party inventory
	Cap      int        // the party inventory's slots
	Use      bool       // items can be used here (ActUse)
	Sell     bool       // a shop buys them (ActSell)
	Reload   bool       // the shop refills clips (ActReload)
	Switch   bool       // ActMember shows another member (not in combat)
	Powers   []Power    // what the member can do besides items (the View menu's)
	Wares    []InvItem  // the shop's goods (Browse; nil elsewhere): Index is ActBuy's N, Price what it costs
	Loot     bool       // Wares are a fight's treasure (the Loot screen): ActBuy takes one for nothing
	Treasure *Loot      // with Loot: the treasure's money and its other options (Detect, Share, Pool, Status)
	Gold     int        // the party's money, with Wares
	Copper   int
	Trade    bool // the menu offers ActTrade
	Halve    bool // ... ActHalve
	Join     bool // ... ActJoin
}

// Power is a member's action of the View menu the inventory offers too.
type Power struct {
	Name string
	Kind ActKind
}

// InvItem is one item of the inventory.
type InvItem struct {
	Index   int      // its place in the member's items (Action.N)
	Name    string   // as the list shows it
	Type    byte     // base item type
	Words   []string // the item's name words, known or not (for its icon)
	Count   int
	Readied bool     // it is readied (shown with present.Data.ReadyMark, for a game without a doll)
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

// Quest is a journal entry: Text is what its giver said, paragraph by paragraph.
type Quest struct {
	Name, Giver, Place string
	Text               []string
	Done               bool
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
	Pane          *Pane // the text pane under the 3D view: out adventuring (Adventure), or it holds text (nil otherwise)
	Window        Window
	AreaMap       *AreaMap // with WinAreaMap: the whole map, for a screen of its own
	Area          AreaID
	X, Y          int
	Facing        Dir
	Day           bool        // 6:00 to 20:00, when the scripts show the day sky
	Inside        bool        // the party stands on an inside square (the original's inside sky colour)
	Quests        []Quest     // the quest journal
	Lore          []Lore      // the journal's Notes
	Exploring     bool        // the party explores: the adventuring decisions (ActForward .. ActCastPick) are taken
	Sweeping      bool        // ... at the slow, searching pace (ActSweep): the HUD says so
	Open          [4]bool     // ... the party's square's sides (N E S W) a step may go through: open, or a door to try (No Grid's walls)
	Interact      bool        // ... on a square whose event waits for ActInteract: the HUD says so
	CanCast       bool        // ... and the selected member has a spell to cast (ActCastPick)
	Camp          *Camp       // the party is in camp: the camp decisions (ActRest .. ActIconEdit) are taken; ActClose leaves
	Pending       Pending     // what a suspended script waits on
	Temple        *Temple     // a temple (PendTemple): Cure, ViewMember, Appraise, KeepAppraised, ActLead; Leave
	Shop          *Shop       // a shop (PendShop): Browse (the goods, ActBuy, ActSell), Identify, Appraise, KeepAppraised, ActLead; Leave
	Loot          *Loot       // the treasure after a fight (PendLoot): LootView, LootTake, LootDetect, Lead; LootLeave
	Door          *DoorOpts   // a locked door (PendDoor): TryDoor
	Training      *Training   // the training hall for the selected member (TrainingView, the presenter's): Train, Learn; CloseTraining
	Powers        *PowerPick  // a member chooses a power: PickPower
	Sheet         *Sheet      // the selected member's sheet (the page): ActInventory, ActSpells .. ActCure, ActMember (another member); ActClose leaves
	IconEditor    *IconEditor // the combat icon editor is up (the page; IconDrop): the presenter lays it out and takes its clicks
	SheetPage     *SheetPage  // the character sheet's page is up (the sheet, the creation's, the spell lists over it): the presenter draws it
	Creation      *Creation   // the creation (or Modify) sheet (the page): ActRoll, ActSetField, ActName, ActStat, ActPicker, ActIconEdit, ActKeep; ActClose abandons (the player agreed to Abandon)
	Picker        *Picker     // a grid picker (the page): ActPicked, ActPage; ActClose leaves
	Slots         *Slots      // the save slots: never the engine's; the presenter puts its F10 screen here (SlotsView)
	Ask           *Ask        // a question waits for its answer: menu, list, typed line, member, key, pause (the presenter's to pace) or spell
	Party         []Member
	Border        [8]image.Image // the window border's pieces by place (ui.TopLeft ..) in EGA colours; all nil: the presenter decodes Theme.Border
	BorderPix     [8]Picture     // Border as the game's pixels in the EGA's own colours (the presenter colours them)
	BorderCaps    [4]image.Image // optional end caps (ui.LeftTopCap ..: left top, left bottom, right top, right bottom), over the side pieces; nil: the side piece
	BorderCapsPix [4]Picture     // BorderCaps as the game's pixels, as BorderPix
	Battle        *Battle        // the combat map (nil outside combat)
	Walls         *Walls         // the 3D view built from the rules' own map (nil: the remastered map Data.AreaMaps names)
	Events        []CombatEvent  // what happened in the fight since the last Apply, in order (Apply's View only)
	Note          Note           // the status line's toast, shown on the bottom row a while after At
	Inventory     *Inventory     // the inventory screen is up (nil otherwise)
	Map           *MapSquares    // the current map, drawn by the presenter at MapCell pixels a square (nil where the game has no map)
	NoMap         bool           // the game marks this area unmappable (game.Automap.NoMap)
	Pictures      []Picture      // the screen's pictures, symbols and fills, bottom first
	Adventure     bool           // the party is out in the world: the 3D view (or the overland map) is the screen
	Prompt        []PaneRun      // the prompt row's text (the last row), X the page's column
	ViewText      [][]PaneRun    // text over the view window (WinText: opProtection's runes), rows from 1, X the page's column
	Results       *Results       // the after-combat page (in a box round the window)
	CreditPage    int            // which credits page (Credit.Page) while Credits
	Credits       bool           // the title's credits page (Credits, RemasterCredit)
	Sequence      *Sequence      // a full-screen animation plays (SequenceStep, SequenceDone)
	Intro         *Intro         // the title waits to be played (TitlePicture, TitleCredits, TitleFlash, TitleDone)
	Full          bool           // Pictures are one whole-screen picture (the title's): fitted to the window
	Cells         image.Point    // the page's size in text cells (40x25 at least; a larger window gives more, ActResize)
	Ink           [16]color.RGBA // the text colours (EGA 0..15) in the page's palette now (all zero: the EGA's own)
	Portrait      CellRect       // the sheet's portrait box (zero elsewhere): Shift-click takes the next head (ActHead), a dropped PNG becomes the portrait (ActPortrait)
	IconDrop      bool           // the combat icon editor is up: a dropped PNG becomes the member's icon (ActIcon)
	IconBox       CellRect       // the creation sheet's combat icon (zero elsewhere): a PNG dropped on it becomes the icon (ActIcon)
	Read          *Read          // journal text the game waits to have read: the presenter pages it (JournalPage), then ActOK
	Purse         *Purse         // the party's money and party items (nil without a party)
	Ordering      bool           // the camp's Party Order bar is up: a card dragged onto another's place moves the member (ActOrder)
	MainMenu      []ActKind      // the party menu is up: the functions open now (ActCreate .. ActQuit, ActMember, ActDropMember); ActLead selects a member
	Logo          *Picture       // with MainMenu: the title picture (Pix) the presenter cuts the lettering from (present.Data.Logo)
	Backdrop      *Picture       // the Amiga painted title under MainMenu, File a PNG path the presenter decodes (nil: black)
	Hall          bool           // MainMenu is the training hall's: over the 3D view, Begin leaves the hall
	Screen        *game.Screen   // a game's typed screen is up (nil: none): the presenter draws it from its data, the Ask on its prompt rows
	Seq           uint64         // counts the Views published: a presenter redraws the page on a new one
	Pace          Pace           // what goes on with time: the call the presenter paces it with
	Resting       bool           // camp's rest goes on (RestStep): a key asks to stop it (StopRest)
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
	AC, HP             int // AC as displayed (10 down to negative)
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

// Temple is a temple's services for the selected member.
type Temple struct {
	Cures    []Cure // Cure X is the index
	Gold     int    // the party's gold
	Appraise Appraise
}

// Cure is a temple service.
type Cure struct {
	Name   string
	Price  int    // gold
	Needed bool   // false: Not says why, and the player is asked to pay anyway
	Not    string // "is not blind."
}

// Appraise is the gems and jewelry to appraise (Appraise, false a gem,
// 1 a jewel).
type Appraise struct {
	Gems, Jewelry int
	Valued        int  // the gold piece value of the one appraised, to sell or keep (ActKeep N 0 sells, 1 keeps); 0: none
	Jewel         bool // ... it is a jewel
	CanKeep       bool // ... there is room to keep it
}

// Shop is a shop's services besides its goods (Inventory.Wares).
type Shop struct {
	Unknown     []Unknown // the party's items Id tells something about
	IDPrice     int       // gold an Id costs
	Gold        int
	Pooled      int // gold pooled with the shop (0: none)
	Appraise    Appraise
	Identified  *InvItem // (extra) the item Id last revealed, its Look lines what it does
	IdentifiedN int      // counts the reveals: the presenter shows a new number once
}

// Unknown is an item to identify: Identify(Member, Item).
type Unknown struct {
	Member, Item int // the member and the index in its items (member 0: with the party inventory)
	Owner, Name  string
	Icon         *InvItem // its type and words: the icon
}

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

// Camp is what camp shows of the selected member and the party.
type Camp struct {
	CanCast  bool        // the selected member has a spell to cast (ActCastPick)
	Book     []BookSpell // the selected member's spell book: how often each is memorized on a rest (ActPlan)
	Scrolls  []BookSpell // the spells on their scrolls: Count 1 while being scribed (ActScribe)
	Healing  bool        // someone resting can heal (RestFix)
	Studying bool        // someone resting would memorize or scribe a spell (RestFix)
	Bestiary []Beast     // the monster kinds met
	Settings []Setting   // the game's numeric options, a row each that steps the value on (ActAlter)
	Icon     bool        // camp offers the selected member's combat icon editor (ActIconEdit)
}

// Setting is a numeric option the game publishes in camp: Value 0..Max,
// shown as Names[Value] when Names has it, else as the number.
type Setting struct {
	Name       string
	Value, Max int
	Names      []string
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
