package parity

import (
	"encoding/json"
	"gobox/asset"
	"gobox/ui"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"gobox/games/por/porgame"
)

// The maps (the minimap, View.Map, the full map) were pictures the engine
// drew; testdata/map_*.png are those pictures, recorded with their typed
// data (map_*.json) at a few squares, the map's edges among them. The
// presenter's mapImage of the data must give them pixel for pixel.
func TestMapImage(t *testing.T) {
	var ink [16]color.RGBA
	for i := range ink {
		ink[i] = color.RGBA{asset.EGAPalette[i][0], asset.EGAPalette[i][1], asset.EGAPalette[i][2], 255}
	}
	e, err := porgame.New(filepath.Join(porDir(t), "data", "assets"), filepath.Join(porDir(t), "data"))
	if err != nil {
		t.Skip("no game files")
	}
	e.Init()                 // the font
	ui.UseGameData(e.Data()) // the label inks
	files, _ := filepath.Glob(filepath.Join(porDir(t), "presentdata", "testdata", "map_*.json"))
	if len(files) == 0 {
		t.Fatal("no recorded maps")
	}
	for _, f := range files {
		var rec struct { // as recorded: a label's ink, not its kind
			Squares [16][16]struct {
				Seen   ui.Seen
				Walls  [4]ui.MapWall
				Ink    int
				Letter byte
			}
			Party, Ahead image.Point
		}
		b, _ := os.ReadFile(f)
		if err := json.Unmarshal(b, &rec); err != nil {
			t.Fatal(err)
		}
		m := ui.MapSquares{Party: rec.Party, Ahead: rec.Ahead}
		for y, row := range rec.Squares {
			for x, q := range row {
				m.Squares[y][x] = ui.MapSquare{Seen: q.Seen, Walls: q.Walls, Kind: slices.Index(e.Data().LabelInks, q.Ink), Letter: q.Letter}
			}
		}
		r, err := os.Open(strings.TrimSuffix(f, ".json") + ".png")
		if err != nil {
			t.Fatal(err)
		}
		want, err := png.Decode(r)
		r.Close()
		if err != nil {
			t.Fatal(err)
		}
		cell, glyph := want.Bounds().Dx()/16, ui.PoolGlyph("Acer710_CGA") // the font the pngs were drawn in
		if strings.HasSuffix(f, "_mini.json") {
			glyph = nil
		}
		got := ui.MapImage(&m, cell, ink, glyph)
		for y := range want.Bounds().Dy() {
			for x := range want.Bounds().Dx() {
				if w := color.RGBAModel.Convert(want.At(x, y)); got.RGBAAt(x, y) != w {
					t.Fatalf("%s: pixel %v is %v, want %v", f, image.Pt(x, y), got.RGBAAt(x, y), w)
				}
			}
		}
	}
}
