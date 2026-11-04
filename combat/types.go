package combat

import (
	"image"
	"image/color"
)

// Board is the combat map as the tabletop shows it: the game's snapshot,
// translated by the game's own adapter.
type Board struct {
	Ground    [25][50]Square
	Tiles     []image.Image // the tile set by tile number
	WallTiles int           // tiles 0..WallTiles-1 are stone wall pieces (else outdoor things: trees, water ...)
	StoneTile int           // the wall tile whose face is the stone texture (a plain one)
	Floor     color.RGBA    // the ground's colour
	Dungeon   bool
	Units     []Unit
	Cursor    image.Point
	CursorOn  bool
	CursorAt  []image.Point
	Marks     []image.Point
	Reach     []image.Point
	Paths     map[image.Point][]Step
	Targets   []int // unit IDs an Attack takes; aiming: the foes Tab cycles through
	Turn      *Turn // nil: no one to act for
	Panel     int
	Log       []string
	Area      bool
	AreaAt    []image.Point
	Skipped   []image.Point
	AimWhy    string
	Effects   []Effect
	View      image.Point // the original's scrolled window, the camera's framing hint
	Strike    Strike
	Confirm   string // the question a Yes or No answers ("" none)
}

// Turn is what the acting unit may do.
type Turn struct {
	Unit       int
	Can        []Kind // the buttons offered, in the game's menu order
	Aim        bool
	Casting    string
	Moves      int
	Labels     map[Kind]string // optional button words
	Keys       map[Kind]int    // optional hotkey letters by button (0 or missing: derived from the word; -1: none)
	CastDirect bool            // Cast acts at once instead of opening the spell list (no spells to list)
}

// Effect is a sprite over the map for a moment.
type Effect struct {
	X, Y float64
	Pic  image.Image
}

// Step is one square of a walk; Free: leaving it lets an enemy take a free attack.
type Step struct {
	image.Point
	Free bool
}

// Strike is one weapon attack: Seq counts attacks, so a new one is a new Seq.
type Strike struct {
	Seq      int
	From, To int
	Hit      bool
	Damage   int
}

// Square is one map square.
type Square struct {
	Tile   int
	Wall   bool
	Height int
}

// Unit is one combatant on the map.
type Unit struct {
	ID         int
	X, Y, Size int
	Squares    []image.Point
	Party      bool
	Facing     int // 0..7
	Icon       image.Image
	IconSize   image.Point // pixels, 24 or 48
	Active     bool
	Name       string
	HP, AC     int
	MaxHP      int
	Hit        int
	Dmg        string
	Status     string
}

// Kind is what a Decision does.
type Kind int

const (
	ActNone Kind = iota
	ActWalk
	ActStep // one keypad step, N = direction 1..9
	ActAttack
	ActAim
	ActPoint
	ActClose
	ActYes
	ActNo
	ActWait
	ActEnd
	ActQuick
	ActGuard
	ActBandage
	ActFlee
	ActTurn
	ActCast
	ActItems
	ActRanged
	ActManeuver
	ActView
)

// Decision is what the player chose.
type Decision struct {
	Kind    Kind
	N, X, Y int
	Pre     image.Point // a maneuver's pre-aimed square (TODO, not emitted yet)
}

// Theme is the host's colours for what combat draws besides the board's own.
// DefaultTheme is por's current look; a game copies and changes it.
type Theme struct {
	Backdrop    color.RGBA // behind the table
	Wood        color.RGBA // the table's plank colour (its grain darkens it)
	Panel       color.RGBA // a button's face behind its icon
	Button      color.RGBA // the button's frame
	Shadow      color.RGBA // darkens what text sits on
	Letter      color.RGBA // the hotkey and the big letter
	Accent      color.RGBA // the selection frame and the hovered word
	Caption     color.RGBA // the casting caption
	Hurt, Heal  color.RGBA // damage numbers
	TextOutline color.RGBA // round the damage numbers
	Series      string     // the asset series whose bar icons show (icons/<series>/)
}

// DefaultTheme is por's current look (EGA colours).
var DefaultTheme = Theme{
	Backdrop:    color.RGBA{0x0C, 0x0B, 0x0A, 0xFF},
	Wood:        color.RGBA{92, 60, 36, 255},
	Panel:       color.RGBA{0x18, 0x20, 0x38, 0xFF},
	Button:      color.RGBA{82, 82, 82, 255},
	Shadow:      color.RGBA{0, 0, 0, 0xB0},
	Letter:      color.RGBA{255, 255, 255, 255},
	Accent:      color.RGBA{82, 255, 82, 255},
	Caption:     color.RGBA{255, 255, 82, 255},
	Hurt:        color.RGBA{255, 255, 85, 255},
	Heal:        color.RGBA{85, 255, 85, 255},
	TextOutline: color.RGBA{0, 0, 0, 255},
	Series:      "pool",
}

// selectOutline fills a frame of colour c, t pixels thick inside r.
func selectOutline(r image.Rectangle, t int, c color.RGBA, fill func(image.Rectangle, color.RGBA)) {
	fill(image.Rect(r.Min.X, r.Min.Y, r.Max.X, r.Min.Y+t), c)
	fill(image.Rect(r.Min.X, r.Max.Y-t, r.Max.X, r.Max.Y), c)
	fill(image.Rect(r.Min.X, r.Min.Y, r.Min.X+t, r.Max.Y), c)
	fill(image.Rect(r.Max.X-t, r.Min.Y, r.Max.X, r.Max.Y), c)
}
