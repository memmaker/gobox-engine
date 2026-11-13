package combat

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"slices"

	"gobox/gpu"
)

// The world: one unit per square, X along the map's columns, Z down its
// rows, Y up. Walls are one square tall.

const (
	combatAtlasCols = 16  // tiles per atlas row
	atlasRows = 17  // the game's 256 tiles, then a row of our own
	tilePx    = 24  // a tile's side in pixels
	atlasPx   = 128 // an atlas slot's side: tiles scale up to it by whole texels, block pictures fit it
	wallSlot  = 256
	floorSlot = 257
)

// kind is how a square stands in the world.
type kind int

const (
	plain kind = iota // the floor alone
	block             // a stone wall block
	stand             // a tree, a cactus, a boulder: its tile extruded upright
	flat              // water, a body, grass: its tile lying on the floor
	slab              // a thin wall: a post with arms toward the walls beside it
)

// slabT is a slab wall's half thickness, in squares.
const slabT = 0.125

// slabH is a slab wall's height, in squares: the original's thin walls stand
// about one and a half times a figure's height.
const slabH = 1.5

// kindOf decides from the square and its tile set: dungeon walls are
// blocks (their pieces only draw the block in the original's perspective,
// so other dungeon squares are floor); outdoors, what blocks sight stands
// up, so does a sprite that blocks movement (a log: a tile with see-through
// texels, unlike water), and the rest lies flat. A tile props.json lists
// says which itself: extrude stands, flat lies, block is a cube in its own picture.
func kindOf(s Square, walls int, empty, cut func(int) bool, ps *propSet) kind {
	switch {
	case ps.modeOf(s.Tile) == "block":
		return block
	case ps.modeOf(s.Tile) == "slab":
		return slab
	case walls > 0 && s.Tile < walls:
		if s.Wall {
			return block
		}
		return plain
	case empty(s.Tile):
		return plain
	case ps.modeOf(s.Tile) == "extrude":
		return stand
	case ps.modeOf(s.Tile) == "flat":
		return flat
	case s.Wall && (s.Height > 0 || cut(s.Tile)):
		return stand
	}
	return flat
}

// layout is every square's kind, and the trunk tile under each standing
// crown, +1 (0: none); a trunk's square is plain floor. A crown with listed
// trunks takes whichever is south of it; any other stands on a trunk only
// when it is tall (Height > 0) and a flat sprite lies south.
func layout(g *[25][50]Square, walls int, empty, cut func(int) bool, ps *propSet) (kinds [25][50]kind, foot [25][50]int) {
	for z := range g {
		for x := range g[z] {
			kinds[z][x] = kindOf(g[z][x], walls, empty, cut, ps)
		}
	}
	for z := range 24 {
		for x := range 50 {
			t, tr := g[z+1][x].Tile, ps.trunkOf(g[z][x].Tile)
			if kinds[z][x] == stand && (tr != nil && slices.Contains(tr, t) ||
				tr == nil && g[z][x].Height > 0 && kinds[z+1][x] == flat && cut(t)) {
				foot[z][x], kinds[z+1][x] = t+1, plain
			}
		}
	}
	return
}

// mapMesh builds the whole map as one mesh over the atlas: the floor with
// corner occlusion, the blocks, and the outdoor things. A thing two tiles
// tall in the original's perspective (a tree: crown over trunk) stands as
// one on the blocked square, the crown on top of the trunk.
func mapMesh(g *[25][50]Square, walls int, empty, cut func(int) bool, tiles []image.Image, ps *propSet) []gpu.MeshVertex {
	kinds, foot := layout(g, walls, empty, cut, ps)
	have := func(t int) bool { return t >= 0 && t < len(tiles) && tiles[t] != nil }
	for z := range kinds { // a standing sprite whose tile has no picture (a partial map) is plain floor: nothing the game sends ever panics
		for x := range kinds[z] {
			if kinds[z][x] == stand && (!have(g[z][x].Tile) || foot[z][x] > 0 && !have(foot[z][x]-1)) {
				kinds[z][x], foot[z][x] = plain, 0
			}
		}
	}
	upright := func(vs []gpu.MeshVertex, t int, fx, fy, fz float32) []gpu.MeshVertex {
		uv := tileUV(t)
		u0, v0, du, dv := uv[3][0], uv[3][1], uv[1][0]-uv[3][0], uv[1][1]-uv[3][1]
		for _, v := range extrude(tiles[t], tilePx) {
			v.X, v.Y, v.Z = v.X+fx+.5, v.Y+fy+.5, v.Z+fz+.5
			v.U, v.V = u0+v.U*du, v0+v.V*dv
			vs = append(vs, v)
		}
		return vs
	}
	solid := func(x, z int) bool { return x >= 0 && z >= 0 && x < 50 && z < 25 && kinds[z][x] == block }
	// occlusion of the floor corner at grid point (x, z): blocks among its four squares
	ao := func(x, z int) float32 {
		n := 0
		for _, d := range [4][2]int{{-1, -1}, {0, -1}, {-1, 0}, {0, 0}} {
			if solid(x+d[0], z+d[1]) {
				n++
			}
		}
		return 1 - 0.22*float32(n)
	}
	up := [3]float32{0, 1, 0}
	var vs []gpu.MeshVertex
	for z := range 25 {
		for x := range 50 {
			fx, fz := float32(x), float32(z)
			k := kinds[z][x]
			if k != block { // a slab stands on the floor too
				vs = shade(quad(vs, [4][3]float32{{fx, 0, fz + 1}, {fx + 1, 0, fz + 1}, {fx + 1, 0, fz}, {fx, 0, fz}}, tileUV(floorSlot), up),
					ao(x, z+1), ao(x+1, z+1), ao(x+1, z), ao(x, z))
			}
			switch k {
			case flat:
				const y = 0.005 // just over the floor
				vs = quad(vs, [4][3]float32{{fx, y, fz + 1}, {fx + 1, y, fz + 1}, {fx + 1, y, fz}, {fx, y, fz}}, tileUV(g[z][x].Tile), up)
			case stand:
				// sunk by the clear rows under the lowest opaque texel, so it stands on the floor
				if f := foot[z][x]; f > 0 {
					vs = append(vs, stacked(tiles[g[z][x].Tile], tiles[f-1], tileUV(g[z][x].Tile), tileUV(f-1), fx, fz)...)
				} else {
					vs = upright(vs, g[z][x].Tile, fx, -gap(tiles[g[z][x].Tile]), fz)
				}
			case slab:
				t := g[z][x].Tile
				side, top := tileUV(t), tileUV(t)
				if ps.top[t] > 0 {
					top = tileUV(ps.top[t])
				}
				wall := func(x, z int) bool {
					return x >= 0 && z >= 0 && x < 50 && z < 25 && (kinds[z][x] == block || kinds[z][x] == slab)
				}
				e, w, s, n := wall(x+1, z), wall(x-1, z), wall(x, z+1), wall(x, z-1)
				if !e && !w && !s && !n {
					e, w = true, true // alone: a short wall along x
				}
				const a, b = .5 - slabT, .5 + slabT
				// the post, open toward its arms; each arm open at both ends
				vs = slabBox(vs, fx, fz, a, a, b, b, side, top, !n, !s, !w, !e)
				if e {
					vs = slabBox(vs, fx, fz, b, a, 1, b, side, top, true, true, false, false)
				}
				if w {
					vs = slabBox(vs, fx, fz, 0, a, a, b, side, top, true, true, false, false)
				}
				if s {
					vs = slabBox(vs, fx, fz, a, b, b, 1, side, top, false, false, true, true)
				}
				if n {
					vs = slabBox(vs, fx, fz, a, 0, b, a, side, top, false, false, true, true)
				}
			case block:
				side := tileUV(wallSlot)
				top := side
				if t := g[z][x].Tile; ps.modeOf(t) == "block" { // a voxel in its own tile's pictures
					side, top = tileUV(t), tileUV(t)
					if ps.top[t] > 0 {
						top = tileUV(ps.top[t])
					}
				}
				vs = quad(vs, [4][3]float32{{fx, 1, fz + 1}, {fx + 1, 1, fz + 1}, {fx + 1, 1, fz}, {fx, 1, fz}}, top, up)
				// sides: darker toward the floor (occlusion where wall meets ground)
				if !solid(x, z+1) {
					vs = shade(quad(vs, [4][3]float32{{fx, 0, fz + 1}, {fx + 1, 0, fz + 1}, {fx + 1, 1, fz + 1}, {fx, 1, fz + 1}}, side, [3]float32{0, 0, 1}), .65, .65, 1, 1)
				}
				if !solid(x, z-1) {
					vs = shade(quad(vs, [4][3]float32{{fx + 1, 0, fz}, {fx, 0, fz}, {fx, 1, fz}, {fx + 1, 1, fz}}, side, [3]float32{0, 0, -1}), .65, .65, 1, 1)
				}
				if !solid(x+1, z) {
					vs = shade(quad(vs, [4][3]float32{{fx + 1, 0, fz + 1}, {fx + 1, 0, fz}, {fx + 1, 1, fz}, {fx + 1, 1, fz + 1}}, side, [3]float32{1, 0, 0}), .65, .65, 1, 1)
				}
				if !solid(x-1, z) {
					vs = shade(quad(vs, [4][3]float32{{fx, 0, fz}, {fx, 0, fz + 1}, {fx, 1, fz + 1}, {fx, 1, fz}}, side, [3]float32{-1, 0, 0}), .65, .65, 1, 1)
				}
			}
		}
	}
	return vs
}

// slabBox is the box (x0, z0)-(x1, z1) of the square at (fx, fz), one
// slabH high, with the faces facing -z, +z, -x, +x as asked; each side
// face takes the part of the side picture its span covers, so a wall
// reads as one picture per two squares whatever pieces build it.
func slabBox(vs []gpu.MeshVertex, fx, fz, x0, z0, x1, z1 float32, side, top [4][2]float32, nz, pz, nx, px bool) []gpu.MeshVertex {
	part := func(a, b float32) [4][2]float32 { // u from a to b of the picture
		u0, du := side[0][0], side[1][0]-side[0][0]
		return [4][2]float32{{u0 + a*du, side[0][1]}, {u0 + b*du, side[1][1]}, {u0 + b*du, side[2][1]}, {u0 + a*du, side[3][1]}}
	}
	X0, X1, Z0, Z1 := fx+x0, fx+x1, fz+z0, fz+z1
	// one picture spans two squares, so its panes come out wider than tall
	h := func(sq, a float32) float32 { return (float32(int(sq)%2) + a) / 2 }
	vs = quad(vs, [4][3]float32{{X0, slabH, Z1}, {X1, slabH, Z1}, {X1, slabH, Z0}, {X0, slabH, Z0}}, top, [3]float32{0, 1, 0})
	if pz {
		vs = shade(quad(vs, [4][3]float32{{X0, 0, Z1}, {X1, 0, Z1}, {X1, slabH, Z1}, {X0, slabH, Z1}}, part(h(fx, x0), h(fx, x1)), [3]float32{0, 0, 1}), .65, .65, 1, 1)
	}
	if nz {
		vs = shade(quad(vs, [4][3]float32{{X1, 0, Z0}, {X0, 0, Z0}, {X0, slabH, Z0}, {X1, slabH, Z0}}, part(1-h(fx, x1), 1-h(fx, x0)), [3]float32{0, 0, -1}), .65, .65, 1, 1)
	}
	if px {
		vs = shade(quad(vs, [4][3]float32{{X1, 0, Z1}, {X1, 0, Z0}, {X1, slabH, Z0}, {X1, slabH, Z1}}, part(1-h(fz, z1), 1-h(fz, z0)), [3]float32{1, 0, 0}), .65, .65, 1, 1)
	}
	if nx {
		vs = shade(quad(vs, [4][3]float32{{X0, 0, Z0}, {X0, 0, Z1}, {X0, slabH, Z1}, {X0, slabH, Z0}}, part(h(fz, z0), h(fz, z1)), [3]float32{-1, 0, 0}), .65, .65, 1, 1)
	}
	return vs
}

// stacked is a crown over its trunk as one picture (24x48), extruded once
// and standing on the square at (fx, fz), sunk by the trunk's clear rows.
// The picture is not in the atlas as one, so each half takes its own tile's
// texels (crown and trunk uv: tileUV): the front and back quads are cut at
// the seam (Y = 0 of the slab); the side faces sample one texel, in a half.
func stacked(crown, trunk image.Image, cuv, tuv [4][2]float32, fx, fz float32) []gpu.MeshVertex {
	pic := image.NewRGBA(image.Rect(0, 0, tilePx, 2*tilePx))
	draw.Draw(pic, image.Rect(0, 0, tilePx, tilePx), crown, crown.Bounds().Min, draw.Src)
	draw.Draw(pic, image.Rect(0, tilePx, tilePx, 2*tilePx), trunk, trunk.Bounds().Min, draw.Src)
	vs := extrude(pic, 2*tilePx) // twice as thick a texel's worth: scaled by 2 below it is the single sprite's
	d := -gap(trunk)
	put := func(v gpu.MeshVertex, top bool) gpu.MeshVertex { // into its half's tile, onto the square
		uv, w := tuv, v.V*2-1
		if top {
			uv, w = cuv, v.V*2
		}
		v.U, v.V = uv[3][0]+v.U*(uv[1][0]-uv[3][0]), uv[3][1]+w*(uv[1][1]-uv[3][1])
		v.X, v.Y, v.Z = 2*v.X+fx+.5, 2*v.Y+1+d, 2*v.Z+fz+.5
		return v
	}
	var out []gpu.MeshVertex
	for _, q := range [2][]gpu.MeshVertex{vs[:6], vs[6:12]} { // front, back: cut at the seam
		a, b, c, e := q[0], q[1], q[2], q[5]
		m, n := a, b
		m.Y, m.V, n.Y, n.V = 0, .5, 0, .5
		for k, h := range [2][4]gpu.MeshVertex{{a, b, n, m}, {m, n, c, e}} {
			for _, i := range [6]int{0, 1, 2, 0, 2, 3} {
				out = append(out, put(h[i], k == 1))
			}
		}
	}
	for _, v := range vs[12:] {
		out = append(out, put(v, v.V < .5))
	}
	return out
}

// gap is the height of the clear rows under a sprite's lowest opaque
// texel, as a fraction of its side.
func gap(img image.Image) float32 {
	b := img.Bounds()
	for y := b.Max.Y - 1; y >= b.Min.Y; y-- {
		for x := b.Min.X; x < b.Max.X; x++ {
			if _, _, _, a := img.At(x, y).RGBA(); a>>8 >= alphaCut {
				return float32(b.Max.Y-1-y) / float32(max(b.Dx(), b.Dy()))
			}
		}
	}
	return 0
}

// shade darkens the last quad's corners (in quad order) by a, b, c, d.
func shade(vs []gpu.MeshVertex, a, b, c, d float32) []gpu.MeshVertex {
	q := vs[len(vs)-6:]
	for i, k := range [6]float32{a, b, c, a, c, d} {
		q[i].R, q[i].G, q[i].B = k, k, k
	}
	return vs
}

// tileUV is atlas slot t's corners in quad order: bottom left, bottom
// right, top right, top left (a hair inside, so nearest sampling never
// reaches the neighbour).
func tileUV(t int) [4][2]float32 {
	const w, h = combatAtlasCols * atlasPx, atlasRows * atlasPx
	const in = 0.01
	u0, v0 := (float32(t%combatAtlasCols*atlasPx)+in)/w, (float32(t/combatAtlasCols*atlasPx)+in)/h
	u1, v1 := u0+(atlasPx-2*in)/w, v0+(atlasPx-2*in)/h
	return [4][2]float32{{u0, v1}, {u1, v1}, {u1, v0}, {u0, v0}}
}

// floorTile is the ground's texture: its colour with a little grain, and
// a faint edge so the squares read (a pixel darker all round).
func floorTile(c color.RGBA) *image.RGBA {
	im := image.NewRGBA(image.Rect(0, 0, tilePx, tilePx))
	for y := range tilePx {
		for x := range tilePx {
			h := uint32(x/2*73856093^y/2*19349663) % 97 // 2x2 grain, the same every square
			k := 0.94 + float64(h)/97*0.1
			if x == 0 || y == 0 || x == tilePx-1 || y == tilePx-1 {
				k *= 0.82
			}
			f := func(v uint8) uint8 { return uint8(min(255, float64(v)*k)) }
			im.SetRGBA(x, y, color.RGBA{f(c.R), f(c.G), f(c.B), 255})
		}
	}
	return im
}

// box is a unit box: X and Z -0.5..0.5, Y 0..1.
func box() []gpu.MeshVertex {
	uv := [4][2]float32{}
	var vs []gpu.MeshVertex
	vs = quad(vs, [4][3]float32{{-.5, 1, .5}, {.5, 1, .5}, {.5, 1, -.5}, {-.5, 1, -.5}}, uv, [3]float32{0, 1, 0})
	vs = quad(vs, [4][3]float32{{-.5, 0, .5}, {.5, 0, .5}, {.5, 1, .5}, {-.5, 1, .5}}, uv, [3]float32{0, 0, 1})
	vs = quad(vs, [4][3]float32{{.5, 0, -.5}, {-.5, 0, -.5}, {-.5, 1, -.5}, {.5, 1, -.5}}, uv, [3]float32{0, 0, -1})
	vs = quad(vs, [4][3]float32{{.5, 0, .5}, {.5, 0, -.5}, {.5, 1, -.5}, {.5, 1, .5}}, uv, [3]float32{1, 0, 0})
	vs = quad(vs, [4][3]float32{{-.5, 0, -.5}, {-.5, 0, .5}, {-.5, 1, .5}, {-.5, 1, -.5}}, uv, [3]float32{-1, 0, 0})
	return vs
}

// place is translate(x, y, z) * rotateY(yaw) * scale(sx, sy, sz).
func place(x, y, z, yaw, sx, sy, sz float32) gpu.Mat4 {
	c, s := float32(math.Cos(float64(yaw))), float32(math.Sin(float64(yaw)))
	return gpu.Mat4{c * sx, 0, -s * sx, 0, 0, sy, 0, 0, s * sz, 0, c * sz, 0, x, y, z, 1}
}

// perspective is a right-handed projection with z 0..1.
func perspective(fovy, aspect, near, far float32) gpu.Mat4 {
	f := 1 / float32(math.Tan(float64(fovy)/2))
	return gpu.Mat4{f / aspect, 0, 0, 0, 0, f, 0, 0, 0, 0, far / (near - far), -1, 0, 0, near * far / (near - far), 0}
}

type vec [3]float64

func (a vec) sub(b vec) vec     { return vec{a[0] - b[0], a[1] - b[1], a[2] - b[2]} }
func (a vec) add(b vec) vec     { return vec{a[0] + b[0], a[1] + b[1], a[2] + b[2]} }
func (a vec) mul(k float64) vec { return vec{a[0] * k, a[1] * k, a[2] * k} }
func (a vec) dot(b vec) float64 { return a[0]*b[0] + a[1]*b[1] + a[2]*b[2] }
func (a vec) norm() vec         { return a.mul(1 / math.Sqrt(a.dot(a))) }
func (a vec) cross(b vec) vec {
	return vec{a[1]*b[2] - a[2]*b[1], a[2]*b[0] - a[0]*b[2], a[0]*b[1] - a[1]*b[0]}
}

// camera is a view of the map: where from, toward what, how wide.
type camera struct {
	eye, at vec
	fovy    float64
	aspect  float64
}

// basis is the camera's right, up and forward directions.
func (c camera) basis() (r, u, f vec) {
	f = c.at.sub(c.eye).norm()
	r = f.cross(vec{0, 1, 0}).norm()
	return r, r.cross(f), f
}

func (c camera) viewProj() gpu.Mat4 {
	r, u, f := c.basis()
	z := f.mul(-1)
	e := c.eye
	view := gpu.Mat4{float32(r[0]), float32(u[0]), float32(z[0]), 0, float32(r[1]), float32(u[1]), float32(z[1]), 0,
		float32(r[2]), float32(u[2]), float32(z[2]), 0, float32(-r.dot(e)), float32(-u.dot(e)), float32(-z.dot(e)), 1}
	return perspective(float32(c.fovy), float32(c.aspect), 0.1, 200).Mul(view)
}

// ndc is where world point w lands on the screen, in -1..1 (y up); ok is
// false behind the eye (pick's inverse).
func (c camera) ndc(w vec) (nx, ny float64, ok bool) {
	r, u, f := c.basis()
	d := w.sub(c.eye)
	z := d.dot(f)
	if !(z > 0.1) { // behind, or NaN from a zero camera (the first frame)
		return 0, 0, false
	}
	t := math.Tan(c.fovy / 2)
	return d.dot(r) / (z * t * c.aspect), d.dot(u) / (z * t), true
}

// pick is the map square under screen point (nx, ny) in -1..1 (y up), on
// the floor; ok is false off the map.
func (c camera) pick(nx, ny float64) (x, z int, ok bool) {
	x, z, hit := c.plane(nx, ny)
	return x, z, hit && onBoard(image.Pt(x, z))
}

// plane is the square of the ground's plane at (nx, ny), on the board or
// past it (hit false: the ray misses the plane).
func (c camera) plane(nx, ny float64) (x, z int, hit bool) {
	r, u, f := c.basis()
	t := math.Tan(c.fovy / 2)
	d := f.add(r.mul(nx * t * c.aspect)).add(u.mul(ny * t))
	if d[1] >= 0 {
		return 0, 0, false
	}
	p := c.eye.add(d.mul(-c.eye[1] / d[1]))
	return int(math.Floor(p[0])), int(math.Floor(p[2])), true
}
