package blobber

import (
	"image"

	"gobox/gpu"
)

// View is the renderer for a host that owns the game loop and the rules
// (a remaster): it draws a Scene from a Camera the host
// moves, into the host's image, and animates each step and turn. Nothing
// here reads the keyboard.
//
// ponytail: one View per process; the scene and camera live in package
// state. Make them per-View when two views must
// draw at once.
type View struct {
	g *Game

	// Textures are the wall, floor and ceiling images by name (FacePart.Tex,
	// "floor_NAME", "ceiling_NAME"); the host fills them, and calls
	// ResetTextures after changing them. Nil: only the generated materials.
	Textures map[string]image.Image
	// Load is the fallback for a name Textures lacks (nil: none), called
	// once a name on first draw, so a large pool decodes only what is shown.
	Load func(name string) image.Image

	// Painted maps a wall character to the Textures name drawn for it (a
	// name may have a name+"_light" glow); '#' is "stone" unless listed.
	Painted map[byte]string

	// The open sky. Night: the night sky and its stars, the world lit by
	// moonlight. SunUp: the sun at bearing SunAz (compass radians, 0 north,
	// pi/2 east) and SunEl above the horizon. Mountains: a range on the
	// northern horizon (in daylight).
	Night, SunUp, Mountains bool
	SunAz, SunEl            float64

	// The ponds' shape (see pondEdge): NewView sets pool half width 0.46,
	// neck half width 0.36 and bank width 0.12 (por's values).
	PondHalf, PondNeck, PondMud float64
}

// cur is the View drawing (one per process, see View).
var cur *View

// Camera is a cell and a facing (0 N, 1 E, 2 S, 3 W).
type Camera struct{ X, Y, Dir int }

// NewView makes the textures and shaders. Call it once the window runs.
// textures is View.Textures (nil: only the generated materials); o are the
// host's options, read as they change (the host owns them).
// The host sets View.Night to start at night.
func NewView(textures map[string]image.Image, o *Options) (*View, error) {
	opt = o
	v := &View{Textures: textures, PondHalf: 0.46, PondNeck: 0.36, PondMud: 0.12}
	cur = v
	g, err := newGame(0)
	if err != nil {
		return nil, err
	}
	v.g = g
	return v, nil
}

// SetFlats makes floor and sky (or ceiling) the flats under and over the
// camera's square (names as in the package's SetFlats).
func (v *View) SetFlats(floor, sky string) { SetFlats(floor, sky) }

// ResetTextures forgets the GPU copies of Textures and of what Load gave (the
// host changed them); each decodes again when next drawn. Scenes parsed
// before hold their props' old images: parse them again.
func (v *View) ResetTextures() {
	for _, t := range wallCache {
		if t != nil {
			t.Deallocate()
		}
	}
	clear(wallCache)
	clear(texMats)
	if texPondWater != nil {
		texPondWater.Deallocate()
	}
	texWater, texClouds, texNightSky, texStars, texPondWater = nil, nil, nil, nil, nil
	if genMats != nil {
		paintMats() // the world is rebuilt (SetFlats)
	}
}

// Show makes s the scene and puts the camera at c at once.
func (v *View) Show(s *Scene, c Camera) {
	s.Show()
	v.FreeOff()
	SetMap(s.Rows, c.X, c.Y)
	Face(c.Dir)
}

// Follow moves the camera to c: a step to a neighbouring cell or a turn is
// animated like the player's own (moves that come while one runs follow it
// without stopping), a turn and a step at once (the auto-walk's corner) as
// the turn, then the step; anything further is a jump.
func (v *View) Follow(c Camera) {
	g := v.g
	last := Camera{g.x, g.y, g.dir}
	if n := len(g.next); n > 0 {
		last = g.next[n-1]
	}
	if c == last {
		return
	}
	if turn := (Camera{last.X, last.Y, c.Dir}); !opt.NoGrid && !oneMove(last, c) && oneMove(turn, c) { // both at once: never a cut, the way walked stays in view
		v.Follow(turn)
		v.Follow(c)
		return
	}
	g.asked = false // the move asked for came
	if opt.NoGrid { // the free camera walks (free.go): no steps queue up behind it
		g.next = g.next[:0]
		SetMap(level, c.X, c.Y)
		Face(c.Dir)
		return
	}
	if g.moving && opt.Smooth && len(g.next) < 8 && oneMove(last, c) {
		g.next = append(g.next, c)
		return
	}
	g.next = g.next[:0]
	g.followStep(c)
}

// oneMove: c is one step or one turn from l.
func oneMove(l, c Camera) bool {
	dx, dy := c.X-l.X, c.Y-l.Y
	return dx*dx+dy*dy == 1 && c.Dir == l.Dir || dx == 0 && dy == 0
}

// followStep animates the move to c, or jumps there when it is not one.
func (g *Game) followStep(c Camera) {
	if !oneMove(Camera{g.x, g.y, g.dir}, c) {
		SetMap(level, c.X, c.Y)
		Face(c.Dir)
		return
	}
	dx, dy := c.X-g.x, c.Y-g.y
	sway := 0.0 // a sidestep rolls the view, as the player's own does
	if r := dirs[(g.dir+1)%4]; [2]int{dx, dy} == r {
		sway = 1
	} else if [2]int{dx, dy} == [2]int{-r[0], -r[1]} {
		sway = -1
	}
	g.startMove(c.X, c.Y, c.Dir, sway)
}

// Flow tells the view whether the host's move key is held and reports
// whether to send the next move now: with fluid movement, once per move,
// halfway through it, so the move arrives before this one ends and the
// view flows on without slowing (walking and turning alike).
// ready says the host can take the move now.
func (v *View) Flow(held, ready bool) bool {
	g := v.g
	g.hold = held
	if !held {
		g.asked = false
		return false
	}
	if !ready || !opt.Fluid || !opt.Smooth || g.asked || len(g.next) > 0 || g.moving && g.t < 0.5 { // idle (a fast step ended before the key counted as held): ask now
		return false
	}
	g.asked = true
	return true
}

// Moving reports whether the view is still stepping or turning (or has
// moves queued).
func (v *View) Moving() bool { return v != nil && v.g != nil && (v.g.moving || len(v.g.next) > 0) }

// Fluid reports whether held keys are paced by Flow.
func (v *View) Fluid() bool { return opt.Fluid && opt.Smooth }

// Update advances the animation by one tick (1/60 s).
func (v *View) Update() {
	v.g.time += 1.0 / 60
	v.g.tween()
}

// Draw draws the 3D view filling dst.
func (v *View) Draw(dst *gpu.Image) {
	b := dst.Bounds()
	v.g.resize(b.Dx(), b.Dy())
	n := gpu.Stats.Draws
	v.g.draw3D(dst, v.g.camera())
	gpu.Stats.Draws3D += gpu.Stats.Draws - n
}
