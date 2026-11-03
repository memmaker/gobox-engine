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
	Unit    int
	Can     []Kind // the buttons offered, in the game's menu order
	Aim     bool
	Casting string
	Moves   int
	Labels  map[Kind]string      // optional button words
	Icons   map[Kind]image.Image // optional button icons
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

// The EGA colours the bar uses.
var egaPalette = [16][3]uint8{
	{0, 0, 0}, {0, 0, 173}, {0, 173, 0}, {0, 173, 173},
	{173, 0, 0}, {173, 0, 173}, {173, 82, 0}, {173, 173, 173},
	{82, 82, 82}, {82, 82, 255}, {82, 255, 82}, {82, 255, 255},
	{255, 82, 82}, {255, 82, 255}, {255, 255, 82}, {255, 255, 255},
}

func ega(i int) color.RGBA { c := egaPalette[i&15]; return color.RGBA{c[0], c[1], c[2], 255} }

// selectOutline fills a green frame t pixels thick inside r.
func selectOutline(r image.Rectangle, t int, fill func(image.Rectangle, color.RGBA)) {
	c := ega(10)
	fill(image.Rect(r.Min.X, r.Min.Y, r.Max.X, r.Min.Y+t), c)
	fill(image.Rect(r.Min.X, r.Max.Y-t, r.Max.X, r.Max.Y), c)
	fill(image.Rect(r.Min.X, r.Min.Y, r.Min.X+t, r.Max.Y), c)
	fill(image.Rect(r.Max.X-t, r.Min.Y, r.Max.X, r.Max.Y), c)
}
