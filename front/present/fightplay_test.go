package present

import (
	"image"
	"image/color"
	"image/draw"
	"testing"

	"gobox/front"
)

func TestPlayMakesSnapshotTiles(t *testing.T) {
	var p Presenter
	pc := front.Picture{Pix: []byte{1, 2, 3, 4}, PixW: 2}
	p.Play([]front.CombatEvent{{Ms: 500, Battle: &front.Battle{TilePix: []front.Picture{pc}}}})
	if b := p.replay.q[0].Battle; len(b.Tiles) != 1 || b.Tiles[0] == nil {
		t.Fatal("snapshot tiles not made")
	}
}

// Space during the computer's turn passes every queued timed moment; a
// moment waiting for a key still waits.
func TestSpaceFastForward(t *testing.T) {
	var p Presenter
	b := &front.Battle{}
	p.Play([]front.CombatEvent{{Ms: 5000, Battle: b}, {Ms: 5000, Battle: b}, {Ms: 5000, Battle: b}, {Key: true, Battle: b}})
	if !p.Playing() {
		t.Fatal("not playing")
	}
	p.Key(' ', front.View{})
	if ev := p.playing(); ev == nil || !ev.Key {
		t.Fatalf("after Space: %+v, want the key moment", ev)
	}
}

// A 5-frame blast (5 timed scenes of 60 ms) plays, and Space passes it.
func TestBlastPlaysAndSpacePasses(t *testing.T) {
	var p Presenter
	var evs []front.CombatEvent
	for i := 0; i < 5; i++ {
		evs = append(evs, front.CombatEvent{Kind: front.EvFrame, Ms: 60, Scene: &front.View{}})
	}
	p.Play(evs)
	if p.playing() == nil {
		t.Fatal("blast not shown")
	}
	p.Key(' ', front.View{})
	if ev := p.playing(); ev != nil || p.Playing() {
		t.Fatalf("after Space: %+v", ev)
	}
}

// A blow plays as a held swing (the attacker posed), then its line shows on
// the pane: what the game said is seen, not only logged.
func TestBlowPlays(t *testing.T) {
	var p Presenter
	posed := &front.Battle{Units: []front.Unit{{ID: 1, Active: true}}}
	p.Play([]front.CombatEvent{{Kind: front.EvAttack, Unit: 1, Target: 2, Ms: 300, Battle: posed},
		{Kind: front.EvHit, Unit: 1, Target: 2}, {Kind: front.EvDamage, Unit: 2, Amount: 5},
		{Kind: front.EvMessage, Text: "ARIN HITS FOR 5"}})
	if ev := p.playing(); ev == nil || ev.Kind != front.EvAttack || ev.Battle != posed {
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
	b := pixBattle(&front.Battle{TilePix: []front.Picture{{Img: flat}}}, egaInk)
	if _, still := flatColour(b.Tiles[0]); still || wallImage("floor") == nil {
		t.Fatal("still a flat fill")
	}
}

// Class 20b: a blow floats once. The classic path's events for POR seed 3
// (attack+hit+damage, attack+miss, a heal): the swing is no MISS, EvHit and
// EvDamage make one number.
func TestBlowFloatsOnce(t *testing.T) {
	es := []front.CombatEvent{
		{Kind: front.EvAttack, Unit: 1, Target: 2}, {Kind: front.EvHit, Unit: 1, Target: 2, Amount: 5}, {Kind: front.EvDamage, Unit: 2, Target: 2, Amount: 5},
		{Kind: front.EvAttack, Unit: 2, Target: 1}, {Kind: front.EvMiss, Unit: 2, Target: 1},
		{Kind: front.EvHeal, Unit: 1, Amount: 3},
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
	if nums != 2 || misses != 1 {
		t.Fatalf("numbers %d misses %d, want 2 and 1", nums, misses)
	}
}
