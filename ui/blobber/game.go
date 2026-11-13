// Blobber: a first-person grid crawler ("gridder") in pure 2D, Deepnight style.
// No 3D pipeline: every surface is a flat quad, transformed by hand, painted far-to-near.
// The texture shader adds Deepnight's look (look.go): stepped darkness into fog, light maps, glows.

package blobber

import (
	"image"
	"image/color"
	"maps"
	"math"
	"math/rand"
	"reflect"

	"gobox/gpu"
)

var (
	W, H  = 640, 480 // set by Layout to the real window size in device pixels
	focal = float64(W) * focalW
)

// The projection, measured from the original view (88 px square; a wall right ahead
// 56 px wide, one square further 24 px): the eye 0.75 squares from the wall ahead, a focal length of
// 42/88 of the square view (in the 4:3 window that square is 3/4 of the width), walls 64/56 as tall
// as wide, standing on the floor 0.5 below the eye.
const (
	focalW  = 42.0 / 88 * 0.75
	wallTop = 0.5 - 64.0/56
)

const (
	near    = 0.05
	eyeBack = 0.25 // camera offset behind the tile center, in tiles (<0.5 keeps it inside the tile)
	radius  = 8    // max ray depth in cells
	tex     = 128  // texture size
)

var level []string // set by SetMap

// Object is a static billboard standing in a cell: map props, levers, items on the floor.
// Size is its longest side in tiles. Objects outside the map are hidden.
type Object struct {
	X, Y int
	Img  *gpu.Image
	Size float64
	At   *[2]float64 // an exact floor position (a tree on a square's edge or corner) instead of the cell's centre
	// Front > 0: while the party stands on its square it stands this far in front of the square's
	// centre, whichever way they face, as the original paints a thing in its own square (Kuto's well)
	Front float64
	Sink  float64     // its foot this far below the floor (a ladder down a well)
	Ring  *[3]float64 // a hole in the floor: half width, half depth, rim width; Img is the rim's stones
}

// texBlack  fills the opening of a Ring.
var texBlack *gpu.Image

// Objects is the host-maintained list of objects on the current map.
var Objects []*Object

// Map cells: '.' is floor, anything else is a wall whose character picks the material (see mats).
func wall(x, y int) bool {
	ch := cell(x, y)
	return ch != '.' && ch != Door
}

// Door is the map character of an open doorway: a passable gap in a wall.
// Locked is a door that is still a wall; scripts open it with `altercell X Y D`.
const Door, Locked = 'D', 'L'

var Faces = map[[3]int][]FacePart{}

// Edges  are thin walls: key {x, y, side} blocks stepping out of cell x, y towards
// dirs[side]. With Faces they give a map edge walls of no thickness, one-sided like the original's.
var Edges = map[[3]int]bool{}

// NoMap  are faces the minimap leaves out: objects the party walks through, like graves.
var NoMap = map[[3]int]bool{}

// blocked: a wall cell at x+dx, y+dy, or a thin wall on the way there.
func blocked(x, y, dx, dy int) bool {
	for i, d := range dirs {
		if d == [2]int{dx, dy} && Edges[[3]int{x, y, i}] {
			return true
		}
	}
	return wall(x+dx, y+dy)
}

// Water  are the cells drawn as water instead of floor; so is everything off the map.
var Water = map[[2]int]bool{}

// Sea : water off the map even without water squares on it (a river along the map's edge).
var Sea bool

// Pond  are cells drawn as water that are not sea: no water off the map for them.
var Pond = map[[2]int]bool{}

// A pond square's pool is a rounded square (a superellipse) of half width View.PondHalf, joined to a
// neighbouring pond square by a neck of half width View.PondNeck; a damp bank View.PondMud wide around it.

// pondEdge is the point of the pool's outline at angle t, from the square's centre.
func pondEdge(t float64) (float64, float64) {
	c, s := math.Cos(t), math.Sin(t)
	return cur.PondHalf * math.Copysign(math.Sqrt(math.Abs(c)), c), cur.PondHalf * math.Copysign(math.Sqrt(math.Abs(s)), s)
}

var texPondWater *gpu.Image // the sea's water, lighter: ponds may be a bright blue

var mudFor map[[2]int]bool         // the Pond map mudTiles were made for
var mudTiles map[[2]int]*gpu.Image // per cell, its share of the bank around the ponds

// pondMud is cell's share of the mud bank around the map's ponds: alpha from the distance to the whole
// outline (discs on the pond cells, joined by strips as wide), so a pond over several cells gets one bank.
func pondMud(cell [2]int) *gpu.Image {
	if len(Pond) == 0 {
		return nil
	}
	if mudTiles == nil || reflect.ValueOf(mudFor).Pointer() != reflect.ValueOf(Pond).Pointer() {
		mudFor, mudTiles = Pond, map[[2]int]*gpu.Image{}
		dist := func(x, z float64) float64 { // to the outline, <0 inside
			d := math.Inf(1)
			for p := range Pond {
				cx, cz := float64(p[0])+0.5, float64(p[1])+0.5
				d = math.Min(d, math.Pow(math.Pow(x-cx, 4)+math.Pow(z-cz, 4), 0.25)-cur.PondHalf)
				for _, s := range [][2]float64{{1, 0}, {0, 1}} {
					if Pond[[2]int{p[0] + int(s[0]), p[1] + int(s[1])}] { // the neck: a box from centre to centre
						a, b := (x-cx)*s[0]+(z-cz)*s[1], math.Abs((x-cx)*s[1]+(z-cz)*s[0])
						d = math.Min(d, math.Hypot(math.Max(0, math.Max(-a, a-1)), math.Max(0, b-cur.PondNeck))+math.Min(0, math.Max(math.Max(-a, a-1), b-cur.PondNeck)))
					}
				}
			}
			return d
		}
		const px = 32
		for p := range Pond {
			for dz := -1; dz <= 1; dz++ {
				for dx := -1; dx <= 1; dx++ {
					c := [2]int{p[0] + dx, p[1] + dz}
					if _, done := mudTiles[c]; done {
						continue
					}
					img := image.NewRGBA(image.Rect(0, 0, px, px))
					for j := 0; j < px; j++ {
						for i := 0; i < px; i++ {
							d := dist(float64(c[0])+(float64(i)+0.5)/px, float64(c[1])+(float64(j)+0.5)/px)
							t := math.Max(0, math.Min(1, d/cur.PondMud))
							a := (1 - t*t*(3-2*t)) * 0.45 // damp earth at the water, a soft fade out (smoothstep)
							img.Set(i, j, color.RGBA{uint8(46 * a), uint8(30 * a), uint8(16 * a), uint8(255 * a)})
						}
					}
					mudTiles[c] = gpu.NewImageFromImage(img)
				}
			}
		}
	}
	return mudTiles[cell]
}

func inMap(x, y int) bool { return y >= 0 && y < len(level) && x >= 0 && x < len(level[y]) }

// texWater, texClouds : floor_water and ceiling_clouds, loaded on first use.
var texWater, texClouds *gpu.Image

type FacePart struct {
	Mat            byte
	T0, T1, U0, U1 float64
	Tex            string // a painted texture by name (View.Textures[TEX]), instead of Mat
}

var texMats = map[string]*surface{} // FacePart.Tex materials, loaded on first use

func texMat(name string) *surface {
	m, ok := texMats[name]
	if !ok {
		if t := wallTexture(name); t != nil {
			m = &surface{img: t, lit: wallTexture(name + "_light")}
		}
		texMats[name] = m
	}
	return m
}

// material returns the texture of the wall at (x, y); unknown chars use stone.
func material(x, y int) *surface {
	if y >= 0 && y < len(level) && x >= 0 && x < len(level[y]) {
		if m := mats[level[y][x]]; m != nil {
			return m
		}
	}
	return mats['#']
}

// ---------- startup texture generation (the "fake 3D" trick) ----------

var (
	rng  = rand.New(rand.NewSource(7))
	bump = 1.0 // heightmap strength baked into the textures
)

// genStone bakes a heightmap of blocks into a lit texture. lx,ly is the baked light direction,
// so a side-facing wall gets its grooves shaded as if seen from the corridor.
func genStone(bw, bh int, offset bool, base, mortar [3]float64, lx, ly float64) *gpu.Image {
	h := make([]float64, tex*tex)
	noise := make([]float64, tex*tex)
	for y := 0; y < tex; y++ {
		for x := 0; x < tex; x++ {
			ox := 0
			if offset && (y/bh)%2 == 1 {
				ox = bw / 2
			}
			bx, by := (x+ox)%bw, y%bh
			d := math.Min(math.Min(float64(bx), float64(bw-1-bx)), math.Min(float64(by), float64(bh-1-by)))
			h[y*tex+x] = math.Max(0, math.Min(1, (d-3)/4)) + rng.Float64()*0.08
			noise[y*tex+x] = 0.85 + 0.3*rng.Float64()
		}
	}
	px := image.NewRGBA(image.Rect(0, 0, tex, tex))
	at := func(x, y int) float64 { return h[((y+tex)%tex)*tex+(x+tex)%tex] }
	for y := 0; y < tex; y++ {
		for x := 0; x < tex; x++ {
			gx, gy := at(x+1, y)-at(x-1, y), at(x, y+1)-at(x, y-1)
			shade := math.Max(0.3, math.Min(1.3, 0.85+0.7*bump*(gx*lx+gy*ly)))
			c := base
			if at(x, y) < 0.3 {
				c = mortar
			}
			n := noise[y*tex+x]
			px.Set(x, y, color.RGBA{u8(c[0] * n * shade), u8(c[1] * n * shade), u8(c[2] * n * shade), 255})
		}
	}
	img := gpu.NewImageFromImage(px)
	return img
}

func u8(v float64) uint8 { return uint8(math.Max(0, math.Min(255, v))) }

var (
	stone             = [3]float64{150, 130, 110}
	mortarC           = [3]float64{80, 72, 70}
	texFloor, texCeil *gpu.Image
	defFloor, defCeil *gpu.Image // before a map picks its own
	texGlow           *gpu.Image // 64x64 white radial falloff, premultiplied
)

// MatColors lists the wall materials by map character: base and mortar colour.
var MatColors = map[byte][2][3]float64{
	'#': {stone, mortarC},
	'B': {{150, 70, 55}, {90, 60, 50}},  // brick
	'M': {{90, 120, 70}, {50, 65, 45}},  // mossy stone
	'W': {{130, 95, 55}, {70, 50, 30}},  // wood
	'I': {{95, 100, 115}, {50, 52, 60}}, // iron
}

var mats = map[byte]*surface{}

var genMats map[byte]*surface    // mats as generated, before the painted ones
var genFloor, genCeil *gpu.Image // the generated floor and ceiling

// paintMats lays the painted textures over the generated materials, floor and ceiling.
func paintMats() {
	clear(mats)
	maps.Copy(mats, genMats)
	painted := map[byte]string{'#': "stone"} // painted textures, each with an optional NAME_light.png
	for ch, name := range paintedWalls() {
		painted[ch] = name
	}
	for ch, name := range painted {
		if t := wallTexture(name); t != nil {
			mats[ch] = &surface{img: t, lit: wallTexture(name + "_light")}
		}
	}
	texFloor, texCeil = genFloor, genCeil
	if t := wallTexture("floor"); t != nil { // painted floor and ceiling, when Textures has them
		texFloor = t
	}
	if t := wallTexture("ceiling"); t != nil {
		texCeil = t
	}
	defFloor, defCeil = texFloor, texCeil
	SetFlats(flatNames[0], flatNames[1])
}

func genTextures() {
	rng.Seed(7)
	worldChanged()
	stoneMat := func(bw, bh int, offset bool, c [2][3]float64) *surface {
		return &surface{
			img:   genStone(bw, bh, offset, c[0], c[1], 0.3, -1),
			right: genStone(bw, bh, offset, c[0], c[1], 1, -0.4),  // wall on our left, face points right
			left:  genStone(bw, bh, offset, c[0], c[1], -1, -0.4), // wall on our right, face points left
		}
	}
	for ch, c := range MatColors {
		mats[ch] = stoneMat(64, 32, true, c)
	}
	door := stoneMat(16, 128, false, MatColors['W']) // vertical planks
	mats[Door], mats[Locked] = door, door
	genMats = maps.Clone(mats)
	genFloor = genStone(64, 64, false, [3]float64{115, 110, 105}, [3]float64{60, 58, 55}, 0.4, -1)
	genCeil = genStone(64, 64, false, [3]float64{70, 66, 72}, [3]float64{40, 38, 42}, 0.4, 1)
	paintMats()
	glow := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			d := math.Hypot(float64(x)-31.5, float64(y)-31.5) / 32
			k := uint8(255 * math.Pow(math.Max(0, 1-d), 2))
			glow.Set(x, y, color.RGBA{k, k, k, k})
		}
	}
	texGlow = gpu.NewImageFromImage(glow)
}

// ---------- game state ----------

type Game struct {
	x, y, dir int
	// smooth movement tween
	fx, fz, fa, tx, tz, ta float64 // from/to pos & angle
	t, sway                float64 // tween progress, roll direction
	moving                 bool
	next                   []Camera // follow mode: the host's moves still to animate, in order
	hold, asked            bool     // follow mode: the host's move key is held; its next move was asked for
	// a step (not a turn) moves at speed v (steps per second) instead of easing, and flows into the
	// next one while the host holds its move key, so walking doesn't slow down at every cell
	walk       bool
	v          float64
	time       float64
	scene      *gpu.Image
	shader     *gpu.Shader      // the dither pass
	meshShader *gpu.MeshShader  // the 3D faces (mesh.go)
	world      *world           // the map's standing geometry
	dyn        []gpu.MeshVertex // the frame's moving faces, and their buffer
	wallRanges []wallRange      // the walls in view, by group
	dyn2       *gpu.Mesh
	bills      []bill
	glows      []glowSrc
	uni        []float32 // the frame's uniforms
	free       free      // No Grid's camera (free.go)
}

var dirs = [4][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}

func (g *Game) startMove(nx, ny, ndir int, sway float64) {
	g.fx, g.fz, g.fa = g.tx, g.tz, g.ta
	g.walk = nx != g.x || ny != g.y
	g.x, g.y, g.dir = nx, ny, ndir
	g.tx, g.tz = float64(nx)+0.5, float64(ny)+0.5
	g.ta = float64(ndir) * math.Pi / 2
	for g.ta-g.fa > math.Pi {
		g.ta -= 2 * math.Pi
	}
	for g.ta-g.fa < -math.Pi {
		g.ta += 2 * math.Pi
	}
	g.t, g.moving, g.sway = 0, true, sway
}

// walking and turning: top speeds in steps / quarter turns per second, and acceleration / braking
const walkSpeed, turnSpeed, walkAccel = 5.0, 7.0, 70.0 // a turn (13 frames) a little quicker than a step (16)

// tween advances a step or turn by one frame: full speed while its key is held (and, for a step, the
// next cell is free), flowing on into the next one; otherwise braking to stop exactly on the cell or
// facing. Without smooth movement a move is done at once.
func (g *Game) tween() {
	const dt = 1.0 / 60
	if !g.moving {
		return
	}
	if !opt.Smooth { // snap: done at once
		g.t, g.v, g.moving = 1, 0, false
		return
	}
	sp := math.Max(MinSpeed, math.Min(MaxSpeed, opt.Speed))
	top, acc := turnSpeed*sp, walkAccel*sp*sp // acceleration scales with speed², so the ease takes the same share of a step
	if g.walk {
		top = walkSpeed * sp
	}
	d := [2]int{int(math.Round(g.tx - g.fx)), int(math.Round(g.tz - g.fz))}
	// the host's next move is known (or, fluid, on its way): flow into it, faster while they pile up
	on := len(g.next) > 0 || opt.Fluid && g.hold && (!g.walk || !blocked(g.x, g.y, d[0], d[1]))
	top *= float64(1 + len(g.next)/2)
	if !on && g.v*g.v >= 2*acc*(1-g.t) { // brake: v²/2a is the distance braking takes
		g.v = math.Max(0.5, g.v-acc*dt)
	} else {
		g.v = math.Min(top, g.v+acc*dt)
	}
	g.t += g.v * dt
	if g.t < 1 {
		return
	}
	g.moving = false
	if on { // straight on, carrying the overshoot and speed into the next step
		over, v := g.t-1, g.v
		if len(g.next) == 0 { // held: wait at speed for the host's next move
			g.t, g.moving = 1, true
			return
		}
		c := g.next[0]
		g.next = g.next[1:]
		g.followStep(c)
		if g.moving {
			g.t, g.v = over, v
			return
		}
	}
	g.t, g.v = 1, 0
}

// ---------- rendering ----------

// camera for this frame
type cam struct{ x, z, ang, roll, fwx, fwz, rtx, rtz, pitch float64 } // pitch: the horizon's shift up, a share of H

func (g *Game) camera() cam {
	if f := g.free; opt.NoGrid && f.on { // No Grid: the eye where the party walks (free.go)
		c := cam{x: f.x, z: f.z, ang: f.yaw, pitch: f.pitch}
		c.fwx, c.fwz = math.Sin(c.ang), -math.Cos(c.ang)
		c.rtx, c.rtz = math.Cos(c.ang), math.Sin(c.ang)
		return c
	}
	e := g.t // tween eases by speed
	if !opt.Smooth {
		e = 1
	}
	c := cam{x: g.fx + (g.tx-g.fx)*e, z: g.fz + (g.tz-g.fz)*e, ang: g.fa + (g.ta-g.fa)*e}
	if g.moving && opt.Smooth {
		s := g.v / (walkSpeed * opt.Speed) // a strafe leans with its speed: in as it starts, steady, out as it stops
		c.ang += 0.025 * s * g.sway
		c.roll = 0.012 * s * g.sway
	}
	c.fwx, c.fwz = math.Sin(c.ang), -math.Cos(c.ang)
	c.rtx, c.rtz = math.Cos(c.ang), math.Sin(c.ang)
	c.x, c.z = c.x-c.fwx*eyeBack, c.z-c.fwz*eyeBack // eye sits behind tile center: side exits visible, turns orbit the center
	return c
}

// visibleCells casts one DDA ray per 2 screen columns across the frustum (plus a margin for sway/roll);
// every open cell a ray crosses before hitting a wall is visible. Frustum + occlusion culling in one loop.
// ponytail: cached per camera; rays can skip a far cell seen through a 1-cell gap, cast more if that shows.
var visCache struct {
	c     cam
	cells map[[2]int]bool
}

func visibleCells(c cam) map[[2]int]bool {
	if visCache.cells != nil && visCache.c == c {
		return visCache.cells
	}
	cells := map[[2]int]bool{{int(math.Floor(c.x)), int(math.Floor(c.z))}: true}
	for sx := -W/2 - 40; sx <= W/2+40; sx += 2 {
		k := float64(sx) / focal
		dx, dz := c.fwx+c.rtx*k, c.fwz+c.rtz*k
		// DDA from (c.x,c.z) along (dx,dz)
		cx, cz := int(math.Floor(c.x)), int(math.Floor(c.z))
		stepX, stepZ := 1, 1
		tdx, tdz := math.Abs(1/dx), math.Abs(1/dz)
		tx, tz := (float64(cx+1)-c.x)*tdx, (float64(cz+1)-c.z)*tdz
		if dx < 0 {
			stepX, tx = -1, (c.x-float64(cx))*tdx
		}
		if dz < 0 {
			stepZ, tz = -1, (c.z-float64(cz))*tdz
		}
		depth := 2 * radius
		if openSky { // daylight hides nothing: rays run until they hit a wall or the map edge
			depth = len(level) + len(level[0])
		}
		for i := 0; i < depth; i++ {
			if tx < tz {
				cx, tx = cx+stepX, tx+tdx
			} else {
				cz, tz = cz+stepZ, tz+tdz
			}
			if wall(cx, cz) && inMap(cx, cz) { // off the map the view runs on over water
				break
			}
			cells[[2]int{cx, cz}] = true
		}
	}
	visCache.c, visCache.cells = c, cells
	return cells
}

// drawSky fills the upper half with the ceiling texture as a panorama: the picture and its mirror
// image side by side make one full turn, so there is no seam and the sky need not tile.
// ponytail: roll is ignored, the horizon stays at H/2.
// night: View.Night of the current View.
func night() bool { return cur != nil && cur.Night }

var nightTint = []float32{0.42, 0.48, 0.72}

// texNightSky, texStars : ceiling_sky_night and ceiling_night_stars, loaded with texClouds.
var texNightSky, texStars *gpu.Image

func (g *Game) drawSky(dst *gpu.Image, c cam) {
	hy := horizon(c)
	if texClouds != nil && night() {
		b := texNightSky.Bounds()
		op := &gpu.DrawImageOptions{}
		op.GeoM.Scale(float64(W)/float64(b.Dx()), hy/float64(b.Dy()))
		dst.DrawImage(texNightSky, op)
		return
	}
	if texClouds != nil { // the ceiling texture is a gradient, stretched; the cloud layer moves
		b := texCeil.Bounds()
		op := &gpu.DrawImageOptions{}
		op.GeoM.Scale(float64(W)/float64(b.Dx()), hy/float64(b.Dy()))
		dst.DrawImage(texCeil, op)
		drawCelestial(dst, c, hy)
		return
	}
	b := texCeil.Bounds()
	s := hy / float64(b.Dy())
	tw := float64(b.Dx()) * s
	turn := math.Mod(c.ang/(2*math.Pi)+1, 1) // 0..1 around the compass
	x := -turn * 2 * tw
	for i := 0; x < float64(W); i, x = i+1, x+tw {
		op := &gpu.DrawImageOptions{}
		if i%2 == 1 {
			op.GeoM.Scale(-s, s)
			op.GeoM.Translate(tw, 0)
		} else {
			op.GeoM.Scale(s, s)
		}
		op.GeoM.Translate(x, 0)
		dst.DrawImage(texCeil, op)
	}
	drawCelestial(dst, c, hy)
}

// horizon is the screen row of the horizon: the middle, shifted by the pitch.
func horizon(c cam) float64 { return float64(H) * (0.5 + c.pitch) }

// resize makes the view w by h pixels.
func (g *Game) resize(w, h int) {
	if w != W || h != H {
		W, H, focal = w, h, float64(w)*focalW
		g.scene = gpu.NewImage(W, H)
	}
}

// Ordered 4x4 Bayer dither with 8 brightness levels: retro darkness without smooth gradients.
const ditherSrc = `
fn bayer2(x: f32, y: f32) -> f32 { return 2.0 * x + 3.0 * y - 4.0 * x * y; }

fn Fragment(dstPos: vec4f, srcPos: vec2f, color: vec4f, custom: vec4f) -> vec4f {
	let c = imageSrc0At(srcPos);
	let x = floor(dstPos.x);
	let y = floor(dstPos.y);
	let b = 4.0 * bayer2(x % 2.0, y % 2.0) + bayer2(floor(x / 2.0) % 2.0, floor(y / 2.0) % 2.0);
	let t = (b + 0.5) / 16.0;
	let levels = 8.0;
	return vec4f(floor(c.rgb * levels + t) / levels, 1.0);
}
`

// theGame is the one Game, the View's.
var theGame *Game

// SetMap swaps the level and places the party at (x, y) without a tween, keeping the facing.
func SetMap(lvl []string, x, y int) {
	level = lvl
	visCache.cells = nil
	worldChanged()
	if theGame == nil {
		theGame = &Game{t: 1}
	}
	g := theGame
	g.x, g.y = x, y
	g.tx, g.tz = float64(x)+0.5, float64(y)+0.5
	g.fx, g.fz, g.fa = g.tx, g.tz, g.ta
	g.t, g.moving = 1, false
}

// cell returns the map character at x,y, or 0 outside the map.
func cell(x, y int) byte {
	if y < 0 || y >= len(level) || x < 0 || x >= len(level[y]) {
		return 0
	}
	return level[y][x]
}

// Face turns the party to dir without a tween.
func Face(dir int) {
	g := theGame
	g.dir = dir
	g.ta = float64(dir) * math.Pi / 2
	g.fa = g.ta
}

// Options are the view's settings (the host's Esc menu sets them).
type Options struct {
	Smooth, Fluid, Dither bool // Fluid: a held key walks or turns on without slowing at each cell
	Torch                 float64
	Speed                 float64 // walk and turn speed, MinSpeed to MaxSpeed times the normal
	NoGrid                bool    // free walking: W A S D and mouselook, the host told each square entered (free.go)
}

// MinSpeed and MaxSpeed bound ViewOptions.Speed.
const MinSpeed, MaxSpeed = 0.5, 2.0

// opt are the options in use: the host's, handed to NewView (the defaults until then).
var opt = &Options{Smooth: true, Fluid: true, Dither: true, Torch: 1, Speed: 1}

// newGame makes the textures and shaders and the one Game, facing dir.
func newGame(dir int) (*Game, error) {
	genTextures()
	tsh, err := gpu.NewMeshShader(meshSrc)
	if err != nil {
		return nil, err
	}
	sh, err := gpu.NewShader([]byte(ditherSrc))
	if err != nil {
		return nil, err
	}
	g := &Game{scene: gpu.NewImage(W, H), shader: sh, meshShader: tsh}
	theGame = g
	g.ta = float64(dir) * math.Pi / 2
	g.dir = dir
	return g, nil
}
