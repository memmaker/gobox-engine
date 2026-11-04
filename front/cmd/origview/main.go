// origview [-scale N] < jobs: draws the original DOS 3D view (the 11x11
// symbol viewport, 88x88 pixels) a party sees at a square and facing, as the
// Gold Box games draw it (ovr031 Draw3dWorld: coab's ovr031.cs, cotab-re
// O116, sotsb-re O17E, buckrogers2-re O0C2 share the code and the tables).
// Each stdin line is a job:
//
//	DATADIR DISK:GEO:B1,B2,B3 X Y DIR OUT.png
//
// DIR is N, E, S or W; B2/B3 127 or 255 = none (then B1 holds all slots, or
// B1 and B3 in Buck Rogers' two-block form). A dev tool: no game imports it.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"gobox/dax"
	"gobox/ega"
	"gobox/geo"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"strings"
)

// The view tables (seg600:0ADA, 0AE4, 0AEE; ovr031 Column_*/Row_*).
var (
	partBase = [10]int{0, 2, 6, 10, 22, 38, 54, 110, 132, 154}
	partW    = [10]int{1, 1, 1, 3, 2, 2, 7, 2, 2, 1}
	partH    = [10]int{2, 4, 4, 4, 8, 8, 8, 11, 11, 2}
	partCol  = [10]int{5, 4, 6, 4, 2, 7, 2, 0, 9, 5}
	partRow  = [10]int{4, 3, 3, 3, 1, 1, 1, 0, 0, 4}
	dx       = [8]int{0, 1, 1, 1, 0, -1, -1, -1}
	dy       = [8]int{-1, -1, 0, 1, 1, 1, 0, -1}
	symFix   = [5]int{0x001, 0x02E, 0x074, 0x0BA, 0x100}
)

const V = 88 // viewport pixels

type mapKey struct {
	disk, geo int
	b         [3]int
}

func none(b int) bool { return b == 127 || b == 255 || b < 0 }

// blockOf: the WALLDEF block and chunk (1-based) wall value n comes from
// (the wall-compare tool's rule, the same as the ports' wall indexes).
func blockOf(m mapKey, n int) (int, int) {
	s := (n - 1) / 5
	if none(m.b[1]) && none(m.b[2]) {
		return m.b[0], s + 1
	}
	if none(m.b[1]) {
		if s == 2 {
			return m.b[2], 3
		}
		return m.b[0], s + 1
	}
	return m.b[s], 1
}

// view holds what Draw3dWorld reads.
type view struct {
	m     *geo.Map
	walls [16][]byte        // wall value -> its 0x9C symbol ids, relocated as LoadWalldef does
	sets  [16][4]*dax.Image // wall value -> symbol sets 0..3 (VGA: one image, raw ids)
	raw   [16][]byte        // unrelocated ids (VGA lookups)
	px    *image.NRGBA
}

func pic(l *dax.Loader, name string, id int) *dax.Image {
	b := l.BlockShared(name, id)
	if b == nil {
		return nil
	}
	if im := dax.ParseVGA(b, true, 0xD); im != nil {
		return im
	}
	return dax.ParseImage(b, true, 0xD)
}

func load(l *dax.Loader, m mapKey) (*view, error) {
	gb := l.BlockShared(fmt.Sprintf("GEO%d", m.disk), m.geo)
	gm, err := geo.ParseMap(gb)
	if err != nil {
		return nil, err
	}
	v := &view{m: gm}
	wname, sname := fmt.Sprintf("WALLDEF%d", m.disk), fmt.Sprintf("8X8D%d", m.disk)
	base := pic(l, "8X8D1", 203)
	for n := 1; n <= 15; n++ {
		blk, chunk := blockOf(m, n)
		if none(blk) {
			continue
		}
		wb := l.BlockShared(wname, blk)
		wd, err := geo.ParseWalldef(wb)
		if err != nil || chunk > len(wd.Sets) {
			continue
		}
		raw := append([]byte(nil), wd.Wall(chunk-1, (n-1)%5)...)
		slot := (n-1)/5 + 1
		delta := symFix[slot-(chunk-1)] - symFix[1] // the block went into slot slot-(chunk-1)
		rel := append([]byte(nil), raw...)
		for i, id := range rel {
			if id >= geo.SymWallBase {
				rel[i] = id + byte(delta)
			}
		}
		v.walls[n], v.raw[n] = rel, raw
		if one := pic(l, sname, blk); one != nil && one.Palette != nil {
			v.sets[n] = [4]*dax.Image{one} // VGA (Buck Rogers): one block, indexed by id
			continue
		}
		v.sets[n][0] = base
		for k := 1; k <= len(wd.Sets) && k <= 3; k++ {
			s := slot - (chunk - 1) + k - 1
			if s > 3 {
				break
			}
			v.sets[n][s] = pic(l, sname, geo.SymbolBlock(blk, k, len(wd.Sets)))
		}
	}
	return v, nil
}

func rgb(c byte) color.NRGBA { p := ega.Palette[c&15]; return color.NRGBA{p[0], p[1], p[2], 255} }

func (v *view) fill(y0, h int, c byte) {
	for y := y0; y < y0+h; y++ {
		for x := range V {
			v.px.SetNRGBA(x, y, rgb(c))
		}
	}
}

// blit draws a picture's frame with its top-left at viewport (x, y), clipped.
func (v *view) blit(im *dax.Image, frame, x, y int) {
	if im == nil || frame < 0 || frame >= im.Count {
		return
	}
	w := im.Width * 8
	for j, c := range im.Frame(frame) {
		var col color.NRGBA
		if im.Palette != nil {
			p := im.Palette[c]
			if p.A == 0 {
				continue
			}
			col = color.NRGBA{p.R, p.G, p.B, 255}
		} else if c >= dax.Clear {
			continue
		} else {
			col = rgb(c)
		}
		px, py := x+j%w, y+j/w
		if px >= 0 && px < V && py >= 0 && py < V {
			v.px.SetNRGBA(px, py, col)
		}
	}
}

// part = draw_3D_8x8_titles: view part p of wall n at cell (row, col).
func (v *view) part(p, n, row, col int) {
	def := v.walls[n]
	if def == nil {
		return
	}
	idx := partBase[p]
	for r := row; r < row+partH[p]; r++ {
		for c := col; c < col+partW[p]; c++ {
			id, raw := int(def[idx]), int(v.raw[n][idx])
			idx++
			if id == 0 || r < 0 || r > 10 || c < 0 || c > 10 {
				continue
			}
			sets := v.sets[n]
			if sets[0] != nil && sets[0].Palette != nil {
				v.blit(sets[0], raw, c*8, r*8)
				continue
			}
			s := 0
			for s < 3 && id >= symFix[s+1] {
				s++
			}
			v.blit(sets[s], id-symFix[s], c*8, r*8)
		}
	}
}

func (v *view) wall(dir, x, y int) int { return int(v.m.WallAt(x, y, dir)) }

// draw = Draw3dWorld (ovr031:0820), far to near.
func (v *view) draw(dir, px, py int) {
	left, back, right := (dir+6)%8, (dir+4)%8, (dir+2)%8
	x, y := px+2*dx[dir], py+2*dy[dir]
	for step := 2; step >= 0; step-- {
		switch step {
		case 2:
			// the far fronts, with the edge pieces (part 9) between them
			for _, side := range [2]struct{ walk, other, sgn int }{{left, right, -1}, {right, left, 1}} {
				tx, ty, col, prev := x, y, 0, 0
				for range 4 {
					w := v.wall(dir, tx, ty)
					if !geo.OnMap(tx, ty) && v.wall(side.other, tx, ty) == 0 {
						prev = 0
					}
					if w != 0 {
						if prev > 0 {
							v.part(9, prev, partRow[9], partCol[9]+col-side.sgn)
						}
						prev = w
						v.part(0, w, partRow[0], partCol[0]+col)
					} else {
						if prev > 0 && v.wall(side.walk, tx-dx[side.walk], ty-dy[side.walk]) != 0 {
							v.part(9, prev, partRow[9], partCol[9]+col-side.sgn)
						}
						prev = 0
					}
					col += 2 * side.sgn
					tx, ty = tx+dx[side.walk], ty+dy[side.walk]
				}
			}
			for _, s := range [2]struct{ walk, p, sgn int }{{left, 1, -1}, {right, 2, 1}} {
				tx, ty, col := x, y, 0
				for i := range 3 {
					if w := v.wall(s.walk, tx, ty); w != 0 {
						c := partCol[s.p] + col
						if i > 0 {
							c += s.sgn
						}
						v.part(s.p, w, partRow[s.p], c)
					}
					col += 2 * s.sgn
					tx, ty = tx+dx[s.walk], ty+dy[s.walk]
				}
			}
		case 1:
			for _, s := range [2]struct{ walk, back, p, col int }{{left, right, 4, -6}, {right, left, 5, 6}} {
				tx, ty, col := x+2*dx[s.walk], y+2*dy[s.walk], s.col
				for range 3 {
					if w := v.wall(dir, tx, ty); w != 0 {
						v.part(3, w, partRow[3], partCol[3]+col)
					}
					if w := v.wall(s.walk, tx, ty); w != 0 {
						v.part(s.p, w, partRow[s.p], partCol[s.p]+col)
					}
					col -= s.col / 2
					tx, ty = tx+dx[s.back], ty+dy[s.back]
				}
			}
		case 0:
			for _, s := range [2]struct{ walk, back, p, col int }{{left, right, 7, -7}, {right, left, 8, 7}} {
				tx, ty, col := x+dx[s.walk], y+dy[s.walk], s.col
				for range 2 {
					if w := v.wall(dir, tx, ty); w != 0 {
						v.part(6, w, partRow[6], partCol[6]+col)
					}
					if w := v.wall(s.walk, tx, ty); w != 0 {
						v.part(s.p, w, partRow[s.p], partCol[s.p]+col)
					}
					col -= s.col
					tx, ty = tx+dx[s.back], ty+dy[s.back]
				}
			}
		}
		x, y = x+dx[back], y+dy[back]
	}
}

// backdrop = Draw3dWorldBackground (VGA path): sky rows 0..43, band 44..45,
// ground 46..87, then the game's own backdrop picture when it has one.
// ponytail: the sky colour comes from area words the ECL sets at run time;
// here it is the usual cyan (11) under the open sky and black under a roof
// (cell >= 0x80), band 7, ground 6 (ovr30:0155's defaults).
func (v *view) backdrop(l *dax.Loader, m mapKey, dir, x, y int) {
	roof := v.m.MapCell(x, y) >= 0x80
	if back := fmt.Sprintf("BACK%d", m.disk); l.Exists(back+".DAX") && !l.Exists("SKYGRND.DAX") { // (Pools of Darkness has a BACK1 of its own use)
		// Buck Rogers: LoadWalldef loads BACKn block b (under a roof) and b+128 (open sky)
		// for the wall block b it loads (O0C2:0E15, O13E:0E1C), the last one loaded wins;
		// Draw3dWorldBackground (O0C2:0175) covers the whole viewport with the one
		// mapWallRoof (cell >= 0x80) picks. The colour fills under it never show.
		b := -1
		for _, k := range m.b { // the slots load in order
			if !none(k) {
				b = k
			}
		}
		if !roof && l.BlockShared(back, b+128) != nil {
			b += 128
		}
		// ponytail: under the open sky with no b+128 picture the original shows only the
		// fills (sky colour from area words); here the roofed picture instead.
		v.blit(pic(l, back, b), 0, 0, 0)
		return
	}
	sky := byte(11)
	if roof {
		sky = 0
	}
	v.fill(0, 44, sky)
	v.fill(44, 2, 7)
	v.fill(46, 42, 6)
	if roof {
		return
	}
	if l.Exists("SKYGRND.DAX") { // Pools of Darkness: a whole picture, block 1 by day (O12A:019D)
		v.blit(pic(l, "SKYGRND", 1), 0, 0, 0)
		return
	}
	if l.Exists("SKY.DAX") { // the sky over the horizon facing north, the horizon at row 7
		if dir == 0 {
			v.blit(pic(l, "SKY", 250), 0, 0, 0)
		}
		v.blit(pic(l, "SKY", 252), 0, 0, 5*8)
	}
}

func Render(l *dax.Loader, m mapKey, x, y, dir int) (*image.NRGBA, error) {
	v, err := load(l, m)
	if err != nil {
		return nil, err
	}
	v.px = image.NewNRGBA(image.Rect(0, 0, V, V))
	v.backdrop(l, m, dir, x, y)
	v.draw(dir, x, y)
	return v.px, nil
}

func scaled(im *image.NRGBA, s int) *image.NRGBA {
	out := image.NewNRGBA(image.Rect(0, 0, V*s, V*s))
	for y := range V * s {
		for x := range V * s {
			out.SetNRGBA(x, y, im.NRGBAAt(x/s, y/s))
		}
	}
	return out
}

func main() {
	scale := flag.Int("scale", 4, "pixel scale")
	flag.Parse()
	loaders := map[string]*dax.Loader{}
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) != 6 {
			continue
		}
		var m mapKey
		var x, y int
		if _, err := fmt.Sscanf(f[1], "%d:%d:%d,%d,%d", &m.disk, &m.geo, &m.b[0], &m.b[1], &m.b[2]); err != nil {
			log.Fatal(f[1], err)
		}
		fmt.Sscan(f[2], &x)
		fmt.Sscan(f[3], &y)
		dir := 2 * strings.Index("NESW", f[4])
		l := loaders[f[0]]
		if l == nil {
			var err error
			if l, err = dax.NewLoader(f[0]); err != nil {
				log.Fatal(err)
			}
			loaders[f[0]] = l
		}
		im, err := Render(l, m, x, y, dir)
		if err != nil {
			log.Fatal(f, err)
		}
		out, err := os.Create(f[5])
		if err != nil {
			log.Fatal(err)
		}
		png.Encode(out, scaled(im, *scale))
		out.Close()
		fmt.Println(f[5])
	}
}
