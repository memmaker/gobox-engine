// Package ui draws the combat map as a tabletop: the map is a board
// on a wooden table, dungeon walls are stone blocks, outdoor things stand
// or lie on it, and the combatants are figurines extruded from their combat
// icons on round painted bases (blue for the party, red for its enemies).
// The camera turns in eighths like XCOM's. It only reads Board:
// what it shows is what the original shows, in 3D.
package ui

import (
	"image"
	"image/color"
	"image/draw"

	xdraw "golang.org/x/image/draw"
	"io/fs"
	"log"
	"math"
	"math/rand/v2"
	"slices"
	"time"

	"gobox/gpu"
)

// CombatPresenter keeps the GPU side of one combat map.
type CombatPresenter struct {
	Theme                      CombatTheme          // the host's colours; set before the first Draw
	Alt                        bool                 // Alt is held (the host sets it each frame): the game's alt+ chords
	Shift                      bool                 // Shift is held (the host sets it each frame): the health bars show
	Icons                      map[Kind]image.Image // a button's icon by action (the game's art; without one the hotkey letter shows); the weapon in hand on Attack, what Ranged shoots
	IconKey                    string               // names them, so the bar redraws when they change
	Meshes                     fs.FS                // <name>.gltf and its PNG for the tiles that have a 3D model (nil: none)
	Props                      fs.FS                // the series' combat directory: props.json says how its tiles draw (nil: all inferred)
	props                      *propSet             // from Props, for the names
	blocks                     map[int]image.Image  // block tiles' own cube pictures by atlas slot (readBlocks)
	white, atlas, shadow, wood *gpu.Image
	tiles                      [256]image.Image // the tiles the atlas holds
	empty                      [256]bool        // tiles with nothing on the ground
	cut                        [256]bool        // tiles with see-through texels: sprites, not ground cover
	floor                      color.RGBA
	floorTile                  image.Image
	names                      []string             // the tile names the meshes were looked up for
	meshOf                     [256]*tileMesh       // the tiles that draw a model
	meshes                     map[string]*tileMesh // by name; nil: none (looked up once)
	ground                     [25][50]Square       // the squares the meshes show
	yawOf                      [25][50]float32      // a model's turn on its square (backYaw)
	walls                      int                  // wall tiles when a dungeon, else 0
	board, shadows, table      *gpu.Mesh
	box, base, flatQuad        *gpu.Mesh
	figs                       map[image.Image]*figure // ponytail: never evicted; a few dozen pictures a game
	bar                        combatBar

	b             *Board       // the last frame's
	out           []Decision   // queued for Actions
	Picking       bool         // the spell list is up (Cast): the bar is off
	PickKind      Kind         // the list Picking shows: ActCast's spells or ActManeuver's maneuvers
	Asking        bool         // a yes/no question (View.Ask, AskConfirm) waits: the bar shows Yes and No
	K             int          // the UI's glyph scale (gobox ui.K), set by the host before Draw: the damage numbers' size
	Pointer       image.Point  // the mouse on the screen, as the host last saw it (the bar's hover; captures script it)
	pointed       image.Point  // the square the aim was last pointed at
	aimAt         *image.Point // the aim cursor (front state, aimTo; nil: the game's Cursor)
	moving        int          // Move's unit ID: its arrows send each step, Reach or not (an off-map step is the game's to answer: flee)
	free          *image.Point // the keys' and pad's square cursor while not aiming (nil: the mouse points)
	mouseAt       image.Point  // the pointer as Point last saw it: a moved mouse takes over from free
	Stick         [2]float64   // the pad's right stick (x right, y down, -1..1): pans the camera
	cursor        image.Point  // the aim cursor's square in the last frame
	tabFrom       *image.Point // Tab was pressed with the aim here: the camera jumps to where it goes
	tabTo         *vec         // the target Tab went to: watched while aiming
	focus, offset vec          // what the camera watches; W A S D's and the right stick's pan from it
	centring      bool         // a new focus: the eye eases all the way onto it, not only into the comfort box
	yaw, yawTo    float64      // camera heading: 0 looks north, as the original's screen
	dist, distTo  float64      // camera distance in squares
	at            vec          // what the camera looks at (eased)
	cam           camera       // the last frame's camera, for picking
	screen        image.Rectangle
	units         map[int]*walker        // each unit's figure as shown (moveUnits)
	paths         map[image.Point][]Step // the last board's walks (a move walks the one it offered)
	started       bool                   // a board was shown: units new after it rise onto the board
	StepTime      float64                // seconds a figure takes for a step (the host sets it from the game's pace; 0: DefaultStepTime)
	sparks        []spark
	sparkMesh     *gpu.Mesh        // all of this frame's sparks, one draw
	sparkVs       []gpu.MeshVertex // its vertices, reused
	boxVs         []gpu.MeshVertex // one spark's cube
	strike        Strike           // the last attack shown
	strikeT       float64          // seconds since it began
	burst         bool             // its sparks are out
	last          time.Time
	clock         float64 // seconds shown, for pulses

	Glyph    func(byte) [8]byte // the game's font, for the damage numbers
	tops     map[int]vec        // over each unit's head, as last drawn (where its numbers start); kept a frame for one gone
	drawnIDs []int              // the units drawn last frame
	floats   []floater
}

// tileMesh is a tile's model on the GPU.
type tileMesh struct {
	mesh *gpu.Mesh
	tex  *gpu.Image
}

type figure struct {
	mesh *gpu.Mesh
	tex  *gpu.Image
	glow [3]float32 // its average colour, for an effect's light and sparks
}

// spark is a particle rising from a spell or a hit.
type spark struct {
	p, v      vec
	life, max float64
	col       [3]float32
}

// Framing and look.
const (
	pitch     = 0.82 // camera radians above the board
	minDist   = 5.0
	maxDist   = 24.0
	pedestal  = 0.09 // a base's height
	lunge     = 0.35 // seconds an attack's lunge lasts
	boardEdge = 0.25 // how far the board stands above the table
)

var (
	partyColour = [4]float32{0.22, 0.42, 0.95, 1}
	enemyColour = [4]float32{0.88, 0.2, 0.14, 1}
	cursorGlow  = [4]float32{1, 0.78, 0.2, 1}
	markGlow    = [4]float32{0.25, 0.8, 0.85, 1} // the original frames choices in light cyan
	hoverGlow   = [4]float32{0.45, 0.45, 0.45, 1}
	areaGlow    = [4]float32{1, 0.5, 0.1, 1}
	areaFill    = [4]float32{0.22, 0.09, 0.01, 1}
	targetGlow  = [4]float32{0.95, 0.12, 0.08, 1}
	blockedGlow = [4]float32{1, 0.08, 0.05, 1} // an aim that cannot be taken: area, frames and cursor in red
	blockedFill = [4]float32{0.26, 0.015, 0.01, 1}
	reachGlow   = [4]float32{0.015, 0.05, 0.13, 1}
	pathGlow    = [4]float32{0.1, 0.35, 0.8, 1}
	skipGlow    = [4]float32{0.55, 0.55, 0.75, 1} // an ally in a spell's area the cast passes over (Bless: an enemy next to it)
	freeGlow    = [4]float32{0.9, 0.1, 0.05, 1}
	sunDir      = [3]float32{0.45, 1, 0.6}
)

// Draw shows b on dst.
func (p *CombatPresenter) Draw(dst *gpu.Image, b *Board) {
	p.initCamera()
	b = p.aimBoard(b)
	p.b = b
	if b.Turn == nil || b.Turn.Aim || !slices.Contains(b.Turn.Can, p.PickKind) {
		p.Picking = false
	}
	p.sync(b)
	dt := p.tick()
	if b.Strike.Seq != p.strike.Seq {
		p.strike, p.strikeT, p.burst = b.Strike, 0, false
	}
	p.strikeT += dt
	figs := p.moveUnits(b, dt)

	// where the eye goes: an effect, else the target Tab went to, else whose
	// turn it is (the aim cursor otherwise is the player's to follow: the camera pans by hand)
	if !b.CursorOn {
		p.tabFrom, p.tabTo = nil, nil
	}
	if f := p.tabFrom; f != nil && b.Cursor != *f { // Tab moved the aim: look there at once
		t := vec{float64(b.Cursor.X) + .5, 0, float64(b.Cursor.Y) + .5}
		p.tabFrom, p.tabTo, p.at = nil, &t, t
	}
	p.cursor = b.Cursor
	if b.Turn == nil || b.Turn.Aim { // the free cursor lives for one turn's choice
		p.free = nil
	}
	focus := restFocus(p.focus, b)
	for _, u := range b.Units {
		if u.Active {
			c := unitCentre(u)
			focus = vec{c[0], 0, c[1]}
		}
	}
	if p.tabTo != nil {
		focus = *p.tabTo
	}
	if c := p.free; c != nil { // the keys' cursor: the eye follows it out of the comfort box
		focus = vec{float64(c.X) + .5, 0, float64(c.Y) + .5}
	}
	if n := len(b.Effects); n > 0 {
		f := b.Effects[n-1]
		focus = vec{f.X + .5, 0, f.Y + .5}
	}
	if focus != p.focus { // a new thing to watch: let go of the pan
		p.focus, p.offset, p.centring = focus, vec{}, true
	}
	panned := p.pan(dt, focus)
	focus = focus.add(p.offset)
	k := 1 - math.Exp(-dt*8)
	p.yaw += (p.yawTo - p.yaw) * k
	p.dist += (p.distTo - p.dist) * k
	if panned { // the camera is in the player's hand: it goes where the keys say, at once
		p.at = focus
	} else {
		c := p.chase(focus)
		if p.centring {
			c = 1
		}
		d := focus.sub(p.at)
		p.at = p.at.add(d.mul(k * c))
		if d = focus.sub(p.at); d[0]*d[0]+d[2]*d[2] < .01 {
			p.centring = false
		}
	}

	r := dst.Bounds()
	h := p.dist * math.Cos(pitch)
	p.cam = camera{
		eye:  vec{p.at[0] + h*math.Sin(p.yaw), p.dist * math.Sin(pitch), p.at[2] + h*math.Cos(p.yaw)},
		at:   p.at,
		fovy: 0.8, aspect: float64(r.Dx()) / float64(r.Dy()),
	}
	p.screen = r
	vp := p.cam.viewProj()
	eye := [3]float32{float32(p.cam.eye[0]), float32(p.cam.eye[1]), float32(p.cam.eye[2])}
	light := gpu.MeshOptions{Light: sunDir, Ambient: 0.5, Eye: eye}
	// an effect lights its surroundings in its own colour
	var glows []*figure
	var effects []Effect // a pictureless effect (no art resolved) has nothing to draw
	for _, f := range b.Effects {
		if f.Pic == nil {
			continue
		}
		effects = append(effects, f)
		g := p.figure(f.Pic, 0)
		glows = append(glows, g)
		light.Point = [4]float32{float32(f.X) + .5, 1.1, float32(f.Y) + .5, 3.5}
		light.PointColor = [4]float32{g.glow[0], g.glow[1], g.glow[2], 1.6}
	}
	opts := func(model gpu.Mat4, tint [4]float32, spec float32) *gpu.MeshOptions {
		o := light
		o.Model, o.Tint, o.Spec = model, tint, spec
		return &o
	}

	dst.Fill(p.Theme.Backdrop)
	dst.ClearDepth()
	if !gpu.Skip("table") {
		dst.DrawMesh(p.table, p.wood, vp, opts(gpu.Ident, [4]float32{1, 1, 1, 1}, 0.1))
	}
	if !gpu.Skip("board") {
		dst.DrawMesh(p.board, p.atlas, vp, opts(gpu.Ident, [4]float32{1, 1, 1, 1}, 0.04))
	}

	// the tiles that have a model, on their squares (the floor shows under them)
	if !gpu.Skip("board") {
		for z := range p.ground {
			for x, s := range p.ground[z] {
				if m := p.modelAt(s); m != nil {
					o := opts(place(float32(x)+.5, 0, float32(z)+.5, p.yawOf[z][x], 1.0/meshUnits, 1.0/meshUnits, 1.0/meshUnits), [4]float32{1, 1, 1, 1}, 0.12)
					o.Ambient = 0.85                   // lit as the figurines are
					dst.DrawMesh(m.mesh, m.tex, vp, o) // ponytail: a draw per square; batch if a map shows hundreds
				}
			}
		}
	}

	// the ground's overlays (a cloud, a body): flat over their squares, under what stands
	for y, row := range b.Ground {
		for x, s := range row {
			if s.Overlay <= 0 || s.Overlay > len(b.Tiles) || b.Tiles[s.Overlay-1] == nil {
				continue
			}
			o := gpu.MeshOptions{Blend: gpu.MeshAlpha, Tint: [4]float32{1, 1, 1, 1}, Light: sunDir, Ambient: 1}
			o.Model = place(float32(x)+.5, 0.01, float32(y)+.5, 0, 1, 1, 1) // over the floor's top (at 0.002 the floor hid it)
			dst.DrawMesh(p.flatQuad, p.figure(b.Tiles[s.Overlay-1], 0).tex, vp, &o)
		}
	}
	// shadows under what stands: the map's things, then the figurines
	shade := gpu.MeshOptions{Blend: gpu.MeshAlpha, Tint: [4]float32{0, 0, 0, 0.7}}
	if !gpu.Skip("shadows") {
		dst.DrawMesh(p.shadows, p.shadow, vp, &shade)
	}
	for _, f := range figs {
		if gpu.Skip("shadows") {
			break
		}
		fw, fd := footprint(f.u)
		o := shade
		o.Model = place(f.x+.12, 0.004, f.z+.1, 0, (fw+.05)*f.scale, 1, (fd+.05)*f.scale)
		dst.DrawMesh(p.flatQuad, p.shadow, vp, &o)
	}

	// frames on the board: choices, the cursor, the square under the mouse
	frame := func(x, z int, tint [4]float32, lift float32) {
		if gpu.Skip("frames") {
			return
		}
		o := gpu.MeshOptions{Blend: gpu.MeshAdd, Tint: tint}
		fx, fz := float32(x), float32(z)
		for _, bar := range [4][4]float32{{fx + .5, fz + .06, .9, .06}, {fx + .5, fz + .94, .9, .06}, {fx + .06, fz + .5, .06, .9}, {fx + .94, fz + .5, .06, .9}} {
			o.Model = place(bar[0], 0.006+lift, bar[1], 0, bar[2], 0.01, bar[3])
			dst.DrawMesh(p.box, p.white, vp, &o)
		}
	}
	marks, fill, cursor := markGlow, areaFill, cursorGlow
	if b.Area { // a spell's area: XCOM's orange
		marks = areaGlow
	}
	if b.AimWhy != "" { // the aimed square cannot be taken: say so in red
		marks, fill, cursor = blockedGlow, blockedFill, blockedGlow
	}
	for _, m := range b.Marks {
		if !targetAt(b, m) { // a target's red frame stands alone
			frame(m.X, m.Y, marks, 0)
		}
	}
	for _, a := range b.AreaAt { // the ground a spell's area covers: an orange wash
		if slices.Contains(b.Skipped, a) { // an ally the cast passes over: framed apart below
			continue
		}
		o := gpu.MeshOptions{Blend: gpu.MeshAdd, Tint: fill, Model: place(float32(a.X)+.5, 0.0035, float32(a.Y)+.5, 0, .96, .005, .96)}
		dst.DrawMesh(p.box, p.white, vp, &o)
	}
	if r := b.AimArea - 1; r >= 0 && len(b.AreaAt) == 0 { // the area previewed round the hovered square
		if hx, hy, ok := p.hoverSquare(); ok {
			for _, a := range areaRound(image.Pt(hx, hy), r) {
				o := gpu.MeshOptions{Blend: gpu.MeshAdd, Tint: areaFill, Model: place(float32(a.X)+.5, 0.0035, float32(a.Y)+.5, 0, .96, .005, .96)}
				dst.DrawMesh(p.box, p.white, vp, &o)
			}
		}
	}
	for _, s := range b.Skipped {
		frame(s.X, s.Y, skipGlow, 0.001)
	}
	for _, u := range b.Units { // the foes a click attacks: red
		if slices.Contains(b.Targets, u.ID) {
			for _, s := range u.Squares {
				frame(s.X, s.Y, targetGlow, 0.001)
			}
		}
	}
	for _, r := range b.Reach { // where a click walks: XCOM's blue
		o := gpu.MeshOptions{Blend: gpu.MeshAdd, Tint: reachGlow, Model: place(float32(r.X)+.5, 0.003, float32(r.Y)+.5, 0, .96, .005, .96)}
		dst.DrawMesh(p.box, p.white, vp, &o)
	}
	if mx, my, ok := p.hoverSquare(); ok {
		frame(mx, my, hoverGlow, 0)
		for i, s := range b.Paths[image.Pt(mx, my)] { // the walk a click takes; red where leaving lets a foe strike
			if i == 0 { // the mover's own square
				continue
			}
			tint := pathGlow
			if s.Free {
				tint = freeGlow
			}
			o := gpu.MeshOptions{Blend: gpu.MeshAdd, Tint: tint, Model: place(float32(s.X)+.5, 0.004, float32(s.Y)+.5, 0, .5, .005, .5)}
			dst.DrawMesh(p.box, p.white, vp, &o)
		}
	}
	pulse := float32(0.75 + 0.25*math.Sin(p.clock*5))
	if b.CursorOn {
		for _, c := range b.CursorAt {
			frame(c.X, c.Y, [4]float32{cursor[0] * pulse, cursor[1] * pulse, cursor[2] * pulse, 1}, 0.002)
		}
	}

	// the figurines
	var heads []head // where each one's health bar goes (Shift)
	for _, f := range figs {
		if gpu.Skip("figures") {
			break
		}
		u, x, z := f.u, f.x, f.z
		ib := iconSize(u)
		sq := float32(max(ib.X, ib.Y)) / tilePx * f.scale
		tint := enemyColour
		if u.Party {
			tint = partyColour
		}
		fw, fd := footprint(u) // the base covers the squares it stands on, not the icon's size
		if u.Down {            // a body: its icon lying on the floor, no base, as the original leaves the fallen
			b := u.Icon.Bounds()
			o := opts(place(x, 0.015, z, 0, sq*.92*float32(b.Dx())/float32(max(b.Dx(), b.Dy())), 1, sq*.92*float32(b.Dy())/float32(max(b.Dx(), b.Dy()))), [4]float32{.75, .75, .75, 1}, 0)
			o.Ambient = 0.9
			dst.DrawMesh(p.flatQuad, p.figure(u.Icon, 0).tex, vp, o)
			continue
		}
		fw, fd = fw*f.scale, fd*f.scale // rising onto the board, sinking away
		dst.DrawMesh(p.base, p.white, vp, opts(place(x, 0, z, 0, fw-.22*f.scale, pedestal*f.scale, fd-.22*f.scale), tint, 0.25))
		if u.Active { // whose turn: a gold ring round the base
			dst.DrawMesh(p.base, p.white, vp, opts(place(x, 0, z, 0, fw-.14*f.scale, pedestal*.5*f.scale, fd-.14*f.scale),
				[4]float32{0.95 * pulse, 0.72 * pulse, 0.2 * pulse, 1}, 0.9))
		}
		fig := p.figure(u.Icon, max(ib.X, ib.Y))
		size := sq * .92
		half := size * float32(ib.Y) / float32(max(ib.X, ib.Y)) / 2
		// the attacker lunges at its target; the target flashes red and sheds sparks when hit, grey ones when missed
		tint = [4]float32{1, 1, 1, 1}
		if t := p.strikeT; t < lunge {
			if to, ok := p.units[p.strike.To]; ok && u.ID == p.strike.From {
				dx, dz := float32(to.at[0])-x, float32(to.at[1])-z
				if n := float32(math.Hypot(float64(dx), float64(dz))); n > 0 {
					d := float32(math.Sin(math.Pi*t/lunge)) * .35 / n
					x, z = x+dx*d, z+dz*d
				}
			}
			if u.ID == p.strike.To && t >= lunge/2 {
				col, n := [3]float32{.6, .6, .6}, 6
				if p.strike.Hit {
					col, n = [3]float32{1, .15, .05}, 16
					w := float32(4 * (1 - t/lunge)) // the flash, fading
					tint = [4]float32{1 + w, 1 - w/4, 1 - w/4, 1}
				}
				for i := 0; !p.burst && i < n; i++ {
					p.sparks = append(p.sparks, spark{
						p:   vec{float64(x), pedestal + float64(half), float64(z)},
						v:   vec{rand.Float64()*2 - 1, rand.Float64() * 1.5, rand.Float64()*2 - 1},
						max: 0.3 + rand.Float64()*.3, col: col,
					})
				}
				p.burst = true
			}
		}
		o := opts(place(x, pedestal+half, z, 0, size, size, size), tint, 0.12)
		o.Ambient = 0.85 // fixed facing, so its back is lit enough to read from any heading
		dst.DrawMesh(fig.mesh, fig.tex, vp, o)
		heads = append(heads, head{u, vec{float64(x), float64(pedestal + 2*half + .1), float64(z)}})
	}

	// effects: the sprite hangs over its square, glowing, shedding sparks
	for i, f := range effects {
		g := glows[i]
		ib := f.Pic.Bounds()
		sq := float32(max(ib.Dx(), ib.Dy())) / tilePx
		x, z := float32(f.X)+float32(ib.Dx())/tilePx/2, float32(f.Y)+float32(ib.Dy())/tilePx/2
		full := gpu.MeshOptions{Tint: [4]float32{1.3, 1.3, 1.3, 1}, Light: sunDir, Ambient: 1}
		full.Model = place(x, 1.05, z, float32(p.yaw), sq*.9, sq*.9, sq*.9) // over the heads, facing the eye: light, not figurines
		dst.DrawMesh(g.mesh, g.tex, vp, &full)
		halo := gpu.MeshOptions{Blend: gpu.MeshAdd, Tint: [4]float32{g.glow[0], g.glow[1], g.glow[2], 0.5}}
		halo.Model = place(x, 0.01, z, 0, 2.2, 1, 2.2)
		dst.DrawMesh(p.flatQuad, p.shadow, vp, &halo) // a pool of light on the board
		for range 2 {
			p.sparks = append(p.sparks, spark{
				p:   vec{float64(x) + rand.Float64()*.6 - .3, 0.8 + rand.Float64()*.4, float64(z) + rand.Float64()*.6 - .3},
				v:   vec{rand.Float64()*.4 - .2, 0.5 + rand.Float64()*.6, rand.Float64()*.4 - .2},
				max: 0.5 + rand.Float64()*.5, col: g.glow,
			})
		}
	}
	if !gpu.Skip("sparks") {
		p.drawSparks(dst, vp, dt)
	}
	if p.Shift {
		p.drawHealth(dst, vp, heads)
	}
	p.tops = nextTops(p.tops, p.drawnIDs, heads)
	p.drawnIDs = p.drawnIDs[:0]
	for _, hd := range heads {
		p.drawnIDs = append(p.drawnIDs, hd.u.ID)
	}
	p.drawDamage(dst, vp, dt)
}

// drawSparks moves, fades and draws the particles, all in one draw call:
// a draw each made missiles and hits stutter.
func (p *CombatPresenter) drawSparks(dst *gpu.Image, vp gpu.Mat4, dt float64) {
	if p.sparkMesh == nil {
		p.sparkMesh, p.boxVs = gpu.NewMesh(nil), box()
	}
	const size = .05
	c, sn := float32(math.Cos(p.yaw))*size, float32(math.Sin(p.yaw))*size
	vs := p.sparkVs[:0]
	live := p.sparks[:0]
	for _, s := range p.sparks {
		s.life += dt
		if s.life >= s.max {
			continue
		}
		s.p = s.p.add(s.v.mul(dt))
		a := float32(1 - s.life/s.max)
		x, y, z := float32(s.p[0]), float32(s.p[1]), float32(s.p[2])
		for _, v := range p.boxVs { // place(x, y, z, yaw, size...) on the CPU
			v.X, v.Y, v.Z = x+c*v.X+sn*v.Z, y+size*v.Y, z-sn*v.X+c*v.Z
			v.NX, v.NZ = c/size*v.NX+sn/size*v.NZ, -sn/size*v.NX+c/size*v.NZ
			v.R, v.G, v.B, v.A = s.col[0], s.col[1], s.col[2], a
			vs = append(vs, v)
		}
		live = append(live, s)
	}
	p.sparks, p.sparkVs = live, vs
	if len(vs) == 0 {
		return
	}
	p.sparkMesh.Set(vs)
	dst.DrawMesh(p.sparkMesh, p.white, vp, &gpu.MeshOptions{Blend: gpu.MeshAdd})
}

// tick is the time since the last frame, at most a tenth of a second.
func (p *CombatPresenter) tick() float64 {
	now := time.Now()
	dt := 0.1
	if !p.last.IsZero() {
		dt = min(now.Sub(p.last).Seconds(), 0.1)
	}
	if p.last.IsZero() || now.Sub(p.last) > unseenGap { // the board was off the screen (a new fight, a page over it): the figures stand where they are
		p.units, p.started = nil, false
	}
	p.last = now
	p.clock += dt
	return dt
}

// sync brings the atlas and the board up to b.
func (p *CombatPresenter) sync(b *Board) {
	if p.white == nil {
		p.white, p.tops = gpu.NewImage(1, 1), map[int]vec{}
		p.white.Fill(color.White)
		p.box = gpu.NewMesh(box())
		p.base = gpu.NewMesh(disc(24))
		p.flatQuad = gpu.NewMesh(quad(nil, [4][3]float32{{-.5, 0, .5}, {.5, 0, .5}, {.5, 0, -.5}, {-.5, 0, -.5}},
			[4][2]float32{{0, 1}, {1, 1}, {1, 0}, {0, 0}}, [3]float32{0, 1, 0}))
		p.atlas = gpu.NewImage(combatAtlasCols*atlasPx, atlasRows*atlasPx)
		p.shadow = gpu.NewImageFromImage(blob(32))
		p.wood = gpu.NewImageFromImage(woodTexture(64, p.Theme.Wood))
		p.table = gpu.NewMesh(tableMesh())
	}
	changed := b.Floor != p.floor || b.FloorTile != p.floorTile
	for i := range p.tiles {
		var t image.Image
		if i < len(b.Tiles) {
			t = b.Tiles[i]
		}
		if t != p.tiles[i] {
			p.tiles[i], changed = t, true
			p.empty[i], p.cut[i] = isEmpty(t), hasHoles(t)
		}
	}
	if !slices.Equal(b.TileNames, p.names) {
		p.names, changed = slices.Clone(b.TileNames), true
		p.meshOf = [256]*tileMesh{}
		p.props = readProps(p.Props, p.names)
		p.blocks = readBlocks(p.Props, p.names, p.props)
		for i, n := range p.names {
			if i < len(p.meshOf) {
				p.meshOf[i] = p.tileMesh(n)
			}
		}
	}
	if changed || b.Ground != p.ground {
		p.floor, p.floorTile = b.Floor, b.FloorTile
		a := image.NewRGBA(p.atlas.Bounds())
		put := func(slot int, t image.Image) {
			o := image.Pt(slot%combatAtlasCols*atlasPx, slot/combatAtlasCols*atlasPx)
			xdraw.NearestNeighbor.Scale(a, image.Rectangle{o, o.Add(image.Pt(atlasPx, atlasPx))}, t, t.Bounds(), draw.Src, nil)
		}
		for i, t := range p.tiles {
			if t != nil {
				put(i, t)
			}
		}
		put(wallSlot, stoneTexture(p.tiles[:b.WallTiles], b.StoneTile, b.Floor))
		if b.FloorTile != nil {
			put(floorSlot, b.FloorTile)
		} else if t := baseFloor(&b.Ground, p.tiles[:]); t >= 0 {
			put(floorSlot, p.tiles[t]) // the area's own ground: no square under a prop is ever a hole
		} else {
			put(floorSlot, floorTile(b.Floor))
		}
		for slot, t := range p.blocks {
			put(slot, t)
		}
		p.atlas.WritePixels(a.Pix)
	}
	walls := 0
	if b.Dungeon {
		walls = b.WallTiles
	}
	if p.board == nil || changed || b.Ground != p.ground || walls != p.walls {
		p.ground, p.walls = b.Ground, walls
		for _, m := range []*gpu.Mesh{p.board, p.shadows} {
			if m != nil {
				m.Dispose()
			}
		}
		empty := func(t int) bool { return t < 0 || t >= len(p.empty) || p.empty[t] }
		cut := func(t int) bool { return t >= 0 && t < len(p.cut) && p.cut[t] }
		// a square with a model gets the floor its neighbours do. Floor tile
		// given: the floor quad shows it. Else the tile most walkable squares
		// show (the floor is a tile on the board's bare ground)
		g, count, under := p.ground, map[int]int{}, -1
		for z := range g {
			for _, s := range g[z] {
				if b.FloorTile == nil && p.modelAt(s) == nil && !s.Wall && kindOf(s, p.walls, empty, cut, p.props) == flat {
					count[s.Tile]++
					if under < 0 || count[s.Tile] > count[under] {
						under = s.Tile
					}
				}
			}
		}
		for z := range g {
			for x, s := range g[z] {
				if p.modelAt(s) != nil {
					g[z][x] = Square{Tile: under}
				}
			}
		}
		p.yawOf = backYaw(&p.ground, p.walls, empty, cut, p.props)
		p.board = gpu.NewMesh(append(mapMesh(&g, p.walls, empty, cut, p.tiles[:], p.props), boardSides()...))
		p.shadows = gpu.NewMesh(standShadows(&g, p.walls, empty, cut, p.props))
	}
}

// modelAt is the model that stands on square s, nil if none. A dungeon
// piece on a walkable square is no wall: nothing stands there.
func (p *CombatPresenter) modelAt(s Square) *tileMesh {
	if s.Tile < 0 || s.Tile >= len(p.meshOf) || p.walls > 0 && s.Tile < p.walls && !s.Wall {
		return nil
	}
	return p.meshOf[s.Tile]
}

// tileMesh is the model of the tile named n (cached; nil: it has none).
func (p *CombatPresenter) tileMesh(n string) *tileMesh {
	if n == "" || p.Meshes == nil {
		return nil
	}
	if m, ok := p.meshes[n]; ok {
		return m
	}
	if p.meshes == nil {
		p.meshes = map[string]*tileMesh{}
	}
	var m *tileMesh
	if _, err := fs.Stat(p.Meshes, n+".gltf"); err == nil {
		vs, img, err := readMesh(p.Meshes, n)
		if err != nil || img == nil {
			log.Printf("combat: mesh %s: %v (texture %v)", n, err, img != nil)
		} else {
			m = &tileMesh{gpu.NewMesh(vs), gpu.NewImageFromImage(img)}
		}
	}
	p.meshes[n] = m
	return m
}

// figure is the extruded mesh for a picture (cached).
// texels: the width of the sprite the picture stands for (0: its own),
// which sets the slab's thickness (extrude).
func (p *CombatPresenter) figure(pic image.Image, texels int) *figure {
	if f, ok := p.figs[pic]; ok {
		return f
	}
	if p.figs == nil {
		p.figs = map[image.Image]*figure{}
	}
	if texels == 0 {
		b := pic.Bounds()
		texels = max(b.Dx(), b.Dy())
	}
	f := &figure{gpu.NewMesh(extrude(pic, texels)), gpu.NewImageFromImage(pic), averageColour(pic)}
	p.figs[pic] = f
	return f
}

// isEmpty: no tile, or nothing on the ground.
func isEmpty(t image.Image) bool {
	if t == nil {
		return true
	}
	b := t.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if _, _, _, a := t.At(x, y).RGBA(); a != 0 {
				return false
			}
		}
	}
	return true
}

// hasHoles is whether a tile has see-through texels (alphaCut).
func hasHoles(t image.Image) bool {
	if t == nil {
		return false
	}
	b := t.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if _, _, _, a := t.At(x, y).RGBA(); a>>8 < alphaCut {
				return true
			}
		}
	}
	return false
}

// averageColour is a picture's mean opaque colour, brightened to a light's.
func averageColour(t image.Image) [3]float32 {
	var r, g, bl, n float64
	b := t.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if cr, cg, cb, a := t.At(x, y).RGBA(); a > 0x8000 {
				r, g, bl, n = r+float64(cr), g+float64(cg), bl+float64(cb), n+1
			}
		}
	}
	if n == 0 {
		return [3]float32{1, 1, 1}
	}
	m := max(r, g, bl) / n / 0xFFFF
	if m == 0 {
		return [3]float32{1, 1, 1}
	}
	return [3]float32{float32(r / n / 0xFFFF / m), float32(g / n / 0xFFFF / m), float32(bl / n / 0xFFFF / m)}
}

// areaRound is the squares within r of c (a square burst, as the Gold Box
// draws an area), clipped to the 50x25 map.
// ponytail: square burst; a game whose areas aren't square states AreaAt instead.
func areaRound(c image.Point, r int) (out []image.Point) {
	for y := c.Y - r; y <= c.Y+r; y++ {
		for x := c.X - r; x <= c.X+r; x++ {
			if x >= 0 && y >= 0 && x < 50 && y < 25 {
				out = append(out, image.Pt(x, y))
			}
		}
	}
	return out
}

// dirs are the sides n, e, s, w as (dx, dz).
var dirs = map[string][2]int{"n": {0, -1}, "e": {1, 0}, "s": {0, 1}, "w": {-1, 0}}

// backYaw turns each model with a backrest (props "back") so the backrest
// faces a wall beside it, the first of n, e, s, w; with none it keeps its
// heading. place's yaw takes (x, z) to (x cos + z sin, -x sin + z cos).
func backYaw(g *[25][50]Square, walls int, empty, cut func(int) bool, ps *propSet) (yaw [25][50]float32) {
	if ps == nil {
		return yaw
	}
	ang := func(d [2]int) float64 { return math.Atan2(float64(-d[1]), float64(d[0])) }
	for z := range g {
		for x, s := range g[z] {
			if s.Tile < 0 || s.Tile >= len(ps.back) || ps.back[s.Tile] == "" {
				continue
			}
			for _, k := range []string{"n", "e", "s", "w"} {
				d := dirs[k]
				nx, nz := x+d[0], z+d[1]
				if nx < 0 || nz < 0 || nx >= 50 || nz >= 25 {
					continue
				}
				if w := kindOf(g[nz][nx], walls, empty, cut, ps); w == block || w == slab {
					yaw[z][x] = float32(ang(d) - ang(dirs[ps.back[s.Tile]]))
					break
				}
			}
		}
	}
	return yaw
}

// mostlySolid: at least half of t's texels are opaque.
func mostlySolid(t image.Image) bool {
	b, n := t.Bounds(), 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if _, _, _, a := t.At(x, y).RGBA(); a>>8 >= alphaCut {
				n++
			}
		}
	}
	return 2*n >= b.Dx()*b.Dy()
}

// baseFloor is the area's base floor: the opaque tile most open squares
// show, -1 if none. It is the ground under every prop and model.
func baseFloor(g *[25][50]Square, tiles []image.Image) int {
	count, best := map[int]int{}, -1
	for z := range g {
		for _, s := range g[z] {
			t := s.Tile
			if s.Wall || t < 0 || t >= len(tiles) || tiles[t] == nil || !mostlySolid(tiles[t]) { // a ground (its dark texels may be keyed clear), not a prop
				continue
			}
			if count[t]++; best < 0 || count[t] > count[best] {
				best = t
			}
		}
	}
	return best
}

// nextTops are the heads drawn now over those drawn before (was: the ids
// drawn last frame): a unit that fell keeps its last place for its number,
// until a new fight (no id drawn last frame drawn now) clears the old ones.
func nextTops(tops map[int]vec, was []int, heads []head) map[int]vec {
	same := len(was) == 0
	for _, hd := range heads {
		same = same || slices.Contains(was, hd.u.ID)
	}
	if !same || tops == nil {
		tops = map[int]vec{}
	}
	for _, hd := range heads {
		tops[hd.u.ID] = hd.top
	}
	return tops
}

// restFocus is where the eye rests with no unit active (a Turn 0 tick): the
// last focus, else the middle of the units, else the View's corner.
func restFocus(last vec, b *Board) vec {
	if last != (vec{}) {
		return last
	}
	var sum vec
	for _, u := range b.Units {
		c := unitCentre(u)
		sum = sum.add(vec{c[0], 0, c[1]})
	}
	if n := float64(len(b.Units)); n > 0 {
		return vec{sum[0] / n, 0, sum[2] / n}
	}
	return vec{float64(b.View.X) + 6, 0, float64(b.View.Y) + 3.5}
}

// aimBoard is b with the aim cursor, front state (aimTo), over the game's:
// it starts on the game's Cursor (which names no squares) and ends with the aim.
func (p *CombatPresenter) aimBoard(b *Board) *Board {
	if b.Turn == nil || !b.Turn.Aim {
		p.aimAt = nil
		return b
	}
	if p.aimAt == nil && len(b.CursorAt) > 0 { // the game draws its own aim
		return b
	}
	c := *b
	p.b = &c
	var at image.Point
	for _, u := range c.Units {
		if u.Active {
			at = image.Pt(u.X, u.Y)
		}
	}
	if len(c.Targets) == 0 { // none published: Tab cycles the foes, the nearest first
		// ponytail: every foe on the map, not only those in the spell's range (the board has no range)
		for _, u := range c.Units {
			if !u.Party {
				c.Targets = append(c.Targets, u.ID)
			}
		}
		dist := func(id int) int {
			for _, u := range c.Units {
				if u.ID == id {
					d := image.Pt(u.X, u.Y).Sub(at)
					return max(d.X, -d.X, d.Y, -d.Y)
				}
			}
			return 1 << 30
		}
		slices.SortStableFunc(c.Targets, func(a, b int) int { return dist(a) - dist(b) })
	}
	if p.aimAt == nil { // the game's Cursor, else the nearest target, else the actor: never a square nobody chose
		switch {
		case c.CursorOn:
			at = c.Cursor
		case len(c.Targets) > 0:
			for _, u := range c.Units {
				if u.ID == c.Targets[0] {
					at = image.Pt(u.X, u.Y)
				}
			}
		}
		p.aimAt = &at
	}
	p.aimTo(*p.aimAt)
	return p.b
}
