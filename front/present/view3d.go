// view3d.go holds the 3D view's overlays: the map pane, the scenes and pictures, the door-open animation and boxes drawn over the view.

package present

import (
	"fmt"
	"gobox/blobber"
	"gobox/front"
	"gobox/gpu"
	"gobox/gpu/vector"
	"gobox/ui"
	"gobox/walls"
	"image"
	"image/color"
	"io/fs"
	"log"
	"strconv"
	"strings"
)

// mapRect is where the map goes: a square in the top right corner, a card
// gap from the edges, mapPx on a side.
func mapRect(dst *gpu.Image, _ int) image.Rectangle {
	b := dst.Bounds()
	gap, side := cardGap*cardBase(dst), mapPx(dst)
	return image.Rect(b.Dx()-gap-side, gap, b.Dx()-gap, gap+side)
}

// mapPx is the map's side: one map pixel to a card pixel at size 1 (Esc,
// Graphics, Minimap size), at most 60 percent of the height and a third of
// the width; the cards make way for it.
func mapPx(dst *gpu.Image) int {
	b := dst.Bounds()
	return int(float64(mapSide) * fitScale(float64(cardBase(dst)), Host.Map, func(k float64) bool {
		return k*float64(mapSide) <= float64(b.Dy())*0.6 && k*float64(mapSide) <= float64(b.Dx())/3
	}))
}

// fitScale is ui.FitScale: unit (K) times a whole size step that fits.
func fitScale(unit, s float64, ok func(k float64) bool) float64 { return ui.FitScale(unit, s, ok) }

// drawMap shows the map as the full map draws it (View.Map) in mapRect.
func (p *Presenter) drawMap(dst *gpu.Image, v front.View, font func(byte, bool) [8]byte) {
	p.mapAt = image.Rectangle{}
	if v.Map == nil || Host.MapOff {
		return
	}
	if p.mapImg == nil {
		p.mapImg = gpu.NewImage(mapSide, mapSide)
	}
	if v.Map != p.mapSrc { // a new View: upload it once
		p.mapImg.WritePixels(mapImage(v.Map, mapCell, v.Ink, font).Pix)
		p.mapSrc = v.Map
	}
	r := mapRect(dst, len(v.Party))
	p.mapAt = r
	a := float32(Host.MapAlpha) // the opacity slider (Esc, Options)
	vector.StrokeRect(dst, float32(r.Min.X)+.5, float32(r.Min.Y)+.5, float32(r.Dx())-1, float32(r.Dy())-1, 1, color.NRGBA{0x80, 0x80, 0x80, uint8(255 * a)})
	vector.DrawFilledRect(dst, float32(r.Min.X+1), float32(r.Min.Y+1), float32(r.Dx()-2), float32(r.Dy()-2), color.NRGBA{A: uint8(255 * a)})
	k := float64(r.Dx()-2) / float64(mapSide) // any scale: the square keeps the aspect
	op := &gpu.DrawImageOptions{}
	op.ColorScale.ScaleAlpha(a)
	op.GeoM.Scale(k, k)
	op.GeoM.Translate(float64(r.Min.X)+(float64(r.Dx())-float64(mapSide)*k)/2, float64(r.Min.Y)+(float64(r.Dy())-float64(mapSide)*k)/2)
	dst.DrawImage(p.mapImg, op)
}

// scene is the remastered map of the area v shows, or nil for an area
// without one (logged once: there is no other view of it).
func (p *Presenter) scene(v front.View) *blobber.Scene {
	name, ok := areaMaps[v.Area]
	if !ok {
		name = fmt.Sprintf("area%d", v.Area)
	}
	return p.sceneNamed(name)
}

// openAt (Data.OpenAt, map:NAME:X:Y:D, D 0..3 for N E S W) shows the 3D
// map NAME from square X, Y alone, whatever the engine shows.

// drawOpen draws the view openAt names.
func (p *Presenter) drawOpen(dst *gpu.Image) {
	f := strings.Split(openAt, ":")
	if len(f) != 5 || f[0] != "map" {
		log.Fatalf("open at %q: want map:NAME:X:Y:D", openAt)
	}
	x, _ := strconv.Atoi(f[2])
	y, _ := strconv.Atoi(f[3])
	d, _ := strconv.Atoi(f[4])
	sc := p.sceneNamed(f[1])
	if sc == nil {
		log.Fatalf("open at %q: no map %s", openAt, f[1])
	}
	if p.shown != f[1] {
		p.view.Show(sc, blobber.Camera{X: x, Y: y, Dir: d})
		p.shown = f[1]
		blobber.SetFlats(sc.Floor, sc.Ceiling)
	}
	p.animate()
	p.view.Draw(dst)
}

// wallImages is View.Textures: the whole shared pool (the maps name their textures; walls.txt is the character table).
func wallImages() map[string]image.Image { return walls.Textures(walls.Names()...) }

// paintedWalls is View.Painted, from assets/walls/walls.txt: map character, then texture name.
func paintedWalls() map[byte]string {
	out := map[byte]string{}
	b, _ := fs.ReadFile(art, "assets/walls/walls.txt")
	for _, l := range strings.Split(string(b), "\n") {
		if f := strings.Fields(strings.SplitN(l, "#", 2)[0]); len(f) == 2 && len(f[0]) == 1 {
			out[f[0][0]] = f[1]
		}
	}
	return out
}

// sceneNamed is the remastered map NAME, or nil (logged once).
func (p *Presenter) sceneNamed(name string) *blobber.Scene {
	if p.view == nil {
		var err error
		if p.view, err = blobber.NewView(wallImages()); err != nil {
			log.Fatalln("remaster:", err)
		}
		p.view.Painted = paintedWalls()
		blobber.Window = func(d *gpu.Image, t string, l []string, k []byte, sel int, b bool) { // its menu and options in the game's box
			p.keyWindow(d, t, l, k, nil, sel, b)
		}
	}
	if sc, ok := p.scenes[name]; ok {
		return sc
	}
	var sc *blobber.Scene
	if f, err := art.Open("maps/" + name + ".map"); err != nil {
		log.Println("remaster:", err)
	} else {
		defer f.Close()
		if sc, _, err = blobber.ParseScene(f); err != nil {
			log.Printf("remaster: %s.map: %v", name, err)
		}
	}
	if p.scenes == nil {
		p.scenes = map[string]*blobber.Scene{}
	}
	p.scenes[name] = sc // nil too: a missing or broken map is logged once
	return sc
}

// Scripted stops animate's clock: a capture script ticks the renderer itself (Tick), so
// every frame of a scripted run is the same run to run.
var Scripted bool

// drawBox copies box r of the classic frame onto dst at scale k, centred
// on cx, cy: its inside, in a plain window of tiles' cells.
func drawBox(dst, inside *gpu.Image, tiles [3]*gpu.Image, r front.CellRect, k float64, cx, cy int) {
	px := r.Pixels()
	var g gpu.GeoM
	g.Scale(k, k)
	g.Translate(float64(cx)-k*float64(px.Dx())/2, float64(cy)-k*float64(px.Dy())/2)
	in := window(dst, tiles, g, r.X2-r.X1+1, r.Y2-r.Y1+1, nil)
	op := &gpu.DrawImageOptions{}
	op.GeoM.Scale(k, k)
	op.GeoM.Translate(float64(in.Min.X), float64(in.Min.Y))
	dst.DrawImage(inside, op)
}
