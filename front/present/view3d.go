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
	if v.Map == nil || Host.MapOff || mapBarred(v) {
		return
	}
	if p.mapImg == nil {
		p.mapImg = gpu.NewImage(mapSide, mapSide)
	}
	cur := -1
	if p.fonts != nil {
		cur = p.fonts.Cur
	}
	if p.mapSrc == nil || *v.Map != *p.mapSrc || v.Ink != p.mapInk || cur != p.mapFont { // the map changed: upload it once (not every View, which builds a new one)
		p.mapImg.WritePixels(mapImage(v.Map, mapCell, v.Ink, font).Pix)
		m := *v.Map
		p.mapSrc, p.mapInk, p.mapFont = &m, v.Ink, cur
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
	if v.Walls != nil {
		return p.wallScene(v.Walls)
	}
	if _, ok := areaMaps[v.Area]; !ok && v.Walls == nil && !v.Exploring { // no 3D view: no map to look up
		return nil
	}
	return p.sceneNamed(sceneName(v))
}

// sceneName names the scene v shows: the rules' map's (View.Walls), or the remastered map's.
func sceneName(v front.View) string {
	if v.Walls != nil {
		return "walls:" + v.Walls.Key
	}
	if name, ok := areaMaps[v.Area]; ok {
		return name
	}
	return fmt.Sprintf("area%d", v.Area)
}

// wallScene is the scene of the rules' own map w: a face and a thin wall on
// every side that holds a wall value, built once a Key.
func (p *Presenter) wallScene(w *front.Walls) *blobber.Scene {
	p.newView()
	name := "walls:" + w.Key
	if sc, ok := p.scenes[name]; ok {
		return sc
	}
	gpu.Note("scene %s built", name)
	rows := make([]string, len(w.Side))
	for y := range rows {
		rows[y] = strings.Repeat(".", len(w.Side[y]))
	}
	sc := &blobber.Scene{Rows: rows, Faces: map[[3]int][]blobber.FacePart{}, Edges: map[[3]int]bool{},
		NoMap: map[[3]int]bool{}, Water: map[[2]int]bool{}, Pond: map[[2]int]bool{}, Floor: w.Floor, Ceiling: w.Sky, FloorRGBA: w.FloorRGBA, CeilRGBA: w.CeilRGBA}
	for y, row := range w.Side {
		for x, sides := range row {
			for d, n := range sides {
				tex := p.wallTexture(w, n)
				if tex == "" {
					continue
				}
				k := [3]int{x, y, d}
				sc.Faces[k] = []blobber.FacePart{{Tex: tex, T1: 1, U1: 1}}
				sc.Edges[k] = true
			}
		}
	}
	p.scenes[name] = sc
	return sc
}

// wallTexture is the texture of wall value n of w: its remastered one from
// the pool, else its picture under a name of w's own ("": none).
func (p *Presenter) wallTexture(w *front.Walls, n int) string {
	if n == 0 {
		return ""
	}
	if t := w.Pool[n]; t != "" && walls.Has(series, t) { // decoded on first draw (View.Load)
		return t
	}
	im := w.Fronts[n]
	if im == nil {
		return ""
	}
	t := w.Key + "#wall" + strconv.Itoa(n) // ponytail: one texture per map and value, kept for the run
	p.view.Textures[t] = im
	return t
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

// wallImage is View.Load: a texture of the game's series of the shared pool,
// decoded when first drawn (the maps name their textures; walls.txt is the character table).
func wallImage(name string) image.Image { return walls.Image(series, name) }

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
	p.newView()
	if sc, ok := p.scenes[name]; ok {
		return sc
	}
	var sc *blobber.Scene
	gpu.Note("scene %s parsed", name)
	if f, err := art.Open("maps/" + name + ".map"); err != nil {
		log.Println("remaster:", err)
	} else {
		defer f.Close()
		if sc, _, err = blobber.ParseScene(f); err != nil {
			log.Printf("remaster: %s.map: %v", name, err)
		}
	}
	p.scenes[name] = sc // nil too: a missing or broken map is logged once
	return sc
}

// wallStyle puts Host.PixelWalls' wall style in use (newView: every scene
// lookup, so a saved style holds from the first frame): the 3D view drops its
// textures and scenes (props hold their images), so the next frame draws the
// other set, each texture decoded again when first drawn.
func (p *Presenter) wallStyle() {
	style := "hd"
	if Host.PixelWalls {
		style = "pixel"
	}
	if walls.Style() == style {
		return
	}
	walls.SetStyle(style)
	if p.view != nil {
		p.view.ResetTextures()
		clear(p.scenes)
		p.shown = ""
	}
}

// newView makes the 3D view and the scenes' cache once.
func (p *Presenter) newView() {
	p.wallStyle()
	if p.scenes == nil {
		p.scenes = map[string]*blobber.Scene{}
	}
	if p.view == nil {
		gpu.Note("3D view made")
		var err error
		if p.view, err = blobber.NewView(map[string]image.Image{}); err != nil {
			log.Fatalln("remaster:", err)
		}
		p.view.Load = wallImage
		p.view.Painted = paintedWalls()
		blobber.Window = func(d *gpu.Image, t string, l []string, k []byte, sel int, b bool) { // its menu and options in the game's box
			p.keyWindow(d, t, l, k, nil, sel, b)
		}
	}
}

// Scripted stops animate's clock: a capture script ticks the renderer itself (Tick), so
// every frame of a scripted run is the same run to run.
var Scripted bool

// drawBox copies box r of the classic frame onto dst at scale k, centred
// on cx, cy: its inside, in a plain window of tiles' cells.
func drawBox(dst, inside *gpu.Image, tiles ui.Ribbon, r front.CellRect, k float64, cx, cy int) {
	px := r.Pixels()
	var g gpu.GeoM
	g.Scale(k, k)
	g.Translate(float64(cx)-k*float64(px.Dx())/2, float64(cy)-k*float64(px.Dy())/2)
	if r == innerWindow.Framed() { // the 3D view's inner box: the theme's "view." edges
		edgePrefix = "view."
		defer func() { edgePrefix = "" }()
	}
	in := window(dst, tiles, g, r.X2-r.X1+1, r.Y2-r.Y1+1, nil)
	op := &gpu.DrawImageOptions{}
	op.GeoM.Scale(k, k)
	op.GeoM.Translate(float64(in.Min.X), float64(in.Min.Y))
	dst.DrawImage(inside, op)
}
