package present

import (
	"gobox/blobber"
	"image"
	"image/color"
	"image/draw"
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

// A replayed board's tiles are made from the game's art when it is drawn.
func TestPlayMakesSnapshotTiles(t *testing.T) {
	g := NewGame(&fake{st: play.Exploring{}}, game.Data{Art: map[string]game.Picture{"t": {W: 2, H: 2, Pixels: []byte{1, 2, 3, 4}, Clear: -1}}})
	if b := presenterOf(g).boardOf(&play.Combat{Tiles: []string{"t"}}, nil, nil); len(b.Tiles) == 0 || b.Tiles[0] == nil {
		t.Fatal("snapshot tiles not made")
	}
}

// Space during the computer's turn passes every queued timed moment; a
// moment waiting for a key still waits.
func TestSpaceFastForward(t *testing.T) {
	var p Presenter
	b := &play.Combat{}
	p.Play([]Moment{{Ms: 5000, Board: b}, {Ms: 5000, Board: b}, {Ms: 5000, Board: b}, {Key: true, Board: b}})
	if !p.Playing() {
		t.Fatal("not playing")
	}
	p.Key(' ')
	if ev := p.playing(); ev == nil || !ev.Key {
		t.Fatalf("after Space: %+v, want the key moment", ev)
	}
}

// A 5-frame blast (5 timed scenes of 60 ms) plays, and Space passes it to
// its last frame (the state the rest leave).
func TestBlastPlaysAndSpacePasses(t *testing.T) {
	var p Presenter
	var evs []Moment
	for i := 0; i < 5; i++ {
		evs = append(evs, Moment{Report: play.Report{Page: &game.Page{}}, Ms: 60})
	}
	p.Play(evs)
	if p.playing() == nil {
		t.Fatal("blast not shown")
	}
	p.Key(' ')
	if ev := p.playing(); ev == nil || len(p.replay.q) != 1 {
		t.Fatalf("after Space: %+v, %d left", ev, len(p.replay.q))
	}
}

// Space passes the ending fight's moments but its last board (the party
// fallen, the foes walked) is still shown before the page.
func TestSpaceKeepsEndBoard(t *testing.T) {
	var p Presenter
	walk, end := &play.Combat{}, &play.Combat{}
	p.Play([]Moment{{Report: play.Report{Kind: game.EvMove}, Ms: 300, Board: walk}, {Report: play.Report{Kind: game.EvDown}, Ms: 300, Board: end}, {Report: play.Report{Kind: game.EvMessage, Text: "x"}}})
	p.Key(' ')
	if ev := p.playing(); ev == nil || ev.Board != end {
		t.Fatalf("after Space: %+v, want the end board", ev)
	}
}

// A blow plays as a held swing (the attacker posed), then its line shows on
// the pane: what the game said is seen, not only logged.
func TestBlowPlays(t *testing.T) {
	var p Presenter
	posed := &play.Combat{Units: []game.Unit{{ID: 1, Active: true}}}
	p.Play([]Moment{{Report: play.Report{Kind: game.EvAttack, Unit: 1, Target: 2}, Ms: 300, Board: posed},
		{Report: play.Report{Kind: game.EvHit, Unit: 1, Target: 2}}, {Report: play.Report{Kind: game.EvDamage, Unit: 2, N: 5}},
		{Report: play.Report{Kind: game.EvMessage, Text: "ARIN HITS FOR 5"}}})
	if ev := p.playing(); ev == nil || ev.Kind != game.EvAttack || ev.Board != posed {
		t.Fatalf("swing not held: %+v", ev)
	}
	p.replay.skip = true
	for p.playing() != nil {
	}
	if p.Playing() || p.blowLine != "ARIN HITS FOR 5" {
		t.Fatalf("line %q", p.blowLine)
	}
}

// A flat-colour combat tile (a Pool floor) draws textured, in its colour.
func TestFlatTileTextured(t *testing.T) {
	pc := front.Picture{Img: image.NewUniform(color.RGBA{130, 65, 45, 255})}
	flat := image.NewRGBA(image.Rect(0, 0, 24, 24))
	draw.Draw(flat, flat.Bounds(), pc.Img, image.Point{}, draw.Src)
	if _, still := flatColour(textured(flat)); still || wallImage("floor") == nil {
		t.Fatal("still a flat fill")
	}
}

// Class 20b: a blow floats once. The classic path's events for POR seed 3
// (attack+hit+damage, attack+miss): the swing is no MISS, EvHit and
// EvDamage make one number.
func TestBlowFloatsOnce(t *testing.T) {
	es := []Moment{
		{Report: play.Report{Kind: game.EvAttack, Unit: 1, Target: 2}}, {Report: play.Report{Kind: game.EvHit, Unit: 1, Target: 2, N: 5}}, {Report: play.Report{Kind: game.EvDamage, Unit: 2, Target: 2, N: 5}},
		{Report: play.Report{Kind: game.EvAttack, Unit: 2, Target: 1}}, {Report: play.Report{Kind: game.EvMiss, Unit: 2, Target: 1}},
	}
	var nums, misses int
	for _, ev := range es {
		_, n, w := blowFloat(ev)
		if n > 0 {
			nums++
		}
		if w == "MISS" {
			misses++
		}
	}
	if nums != 1 || misses != 1 {
		t.Fatalf("numbers %d misses %d, want 1 and 1", nums, misses)
	}
}

// Class 12: no Walls combination draws a flat-colour floor or sky. A flat
// Backdrop never shows, and a named flat resolves on its own.
func TestNoFlatSky(t *testing.T) {
	cream := image.NewRGBA(image.Rect(0, 0, 8, 8))
	draw.Draw(cream, cream.Bounds(), image.NewUniform(color.RGBA{230, 220, 180, 255}), image.Point{}, draw.Src)
	for _, w := range []game.Walls{{}, {Floor: "nosuch", Sky: "clouds"}, {Floor: "nosuch", Sky: "nosuch"}} {
		if !flatsPainted(&w, cream) {
			t.Errorf("%q/%q: flat backdrop shown", w.Floor, w.Sky)
		}
	}
	art := image.NewRGBA(image.Rect(0, 0, 4, 1))
	art.Pix[0] = 9
	if !flatsPainted(&game.Walls{Sky: "clouds"}, art) {
		t.Error("clouds resolve without a floor name")
	}
	if _, still := flatColour(blobber.Tint(wallImage("ceiling"), color.RGBA{230, 220, 180, 255})); still {
		t.Error("a flat colour stays flat")
	}
}
