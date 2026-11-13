package present

import (
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
	"gobox/gpu"
	"image"
)

// counting is a fake that counts the State reads (every View build reads it).
type counting struct {
	*fake
	reads int
}

func (f *counting) State() play.State { f.reads++; return f.fake.State() }

// countGame is the presenter's game over a counting fake.
func countGame() (*classicGame, *counting) {
	f := &counting{fake: &fake{st: play.Exploring{}}}
	return &classicGame{Game: NewGame(f, game.Data{})}, f
}

// The shell asks for the View several times a frame (Screen, Free, Act's
// pacing, Draw): it is built once until something may change it.
func TestFrontBuildsViewOncePerFrame(t *testing.T) {
	g, c := countGame()
	f := &Front{Game: g, started: true}
	f.Screen()
	one := c.reads // one View build's reads
	f.Free()
	f.Free()
	if c.reads != one {
		t.Fatalf("Free built the View again: %d reads, want %d", c.reads, one)
	}
	f.Act() // nothing to run: the View stays
	f.Screen()
	f.Free()
	if c.reads != one {
		t.Fatalf("after an idle Act: %d reads, want %d", c.reads, one)
	}
	f.Key(' ') // a key may change it: the next ask builds it again
	f.Screen()
	if c.reads <= one { // (the derived reads stay: the State is the same)
		t.Fatalf("after Key: %d reads, want more than %d", c.reads, one)
	}
}

// The minimap is drawn and uploaded when the map changes, not for every
// frame (each builds a new MapSquares).
func TestMinimapUploadsOnChange(t *testing.T) {
	var p Presenter
	dst := gpu.NewImage(640, 400)
	font := func(byte, bool) [8]byte { return [8]byte{} }
	st := func(party image.Point) play.State {
		return play.Exploring{Common: play.Common{Where: game.Where{X: party.X, Y: party.Y}}, Map: play.ExploreMap{Automap: game.Automap{Size: image.Pt(16, 16)}}}
	}
	g, f := playGame(st(image.Pt(1, 1)))
	p.SetEngine(g)
	p.drawMap(dst, egaInk, font)
	first := p.mapSrc
	f.st = st(image.Pt(1, 1))
	p.drawMap(dst, egaInk, font)
	if p.mapSrc != first {
		t.Error("an unchanged map was drawn again")
	}
	f.st = st(image.Pt(2, 1))
	p.drawMap(dst, egaInk, font)
	if p.mapSrc == first {
		t.Error("a changed map was not drawn again")
	}
}

// A key pressed while the fight's events play is kept, not dropped: it
// acts once they are played (Draw hands it on).
func TestKeyDuringPlayKept(t *testing.T) {
	g, _ := countGame()
	f := &Front{Game: g, started: true}
	f.P.Play([]Moment{{Ms: 5000, Board: &play.Combat{}}})
	f.Key('x')
	if len(f.held) != 1 || f.held[0] != 'x' || !f.P.Playing() {
		t.Fatalf("held %v, playing %v", f.held, f.P.Playing())
	}
	f.Key(' ') // Space still cuts the moment short
	if len(f.held) != 1 {
		t.Fatalf("held %v", f.held)
	}
}

// The camera's keys (Q, E, Home) turn it while the fight's events play: not
// held for after, and the moment plays on.
func TestCameraKeyDuringPlay(t *testing.T) {
	g, _ := countGame()
	f := &Front{Game: g, started: true}
	f.P.fight = true
	f.P.Play([]Moment{{Ms: 5000, Board: &play.Combat{}}})
	f.Key('q')
	if len(f.held) != 0 || !f.P.Playing() || f.P.replay.skip {
		t.Fatalf("held %v, playing %v, skip %v", f.held, f.P.Playing(), f.P.replay.skip)
	}
}

// artCounting is a fake whose Art() is read by every board build and View build.
type artCounting struct {
	*fake
	arts int
}

func (f *artCounting) Art() map[string]game.Picture { f.arts++; return nil }

// The game is turn-based: idle frames of a fight draw the board built once;
// a new State of the game (its Version) builds it again.
func TestFightBoardBuiltOncePerChange(t *testing.T) {
	fight := play.Fighting{Common: allow(play.EndTurn{}),
		Field: play.Battlefield{Turn: 1, Size: image.Pt(8, 8), Units: []game.Unit{{ID: 1, Member: -1}}}}
	g := &artCounting{fake: &fake{st: fight}}
	f := &Front{Game: &classicGame{Game: NewGame(g, game.Data{})}, started: true}
	f.Glyph, f.Font = func(byte) [8]byte { return [8]byte{} }, func(byte, bool) [8]byte { return [8]byte{} }
	f.P.SetEngine(f.Game)
	dst := gpu.NewImage(640, 400)
	f.Draw(dst)
	if !f.P.fight {
		t.Fatal("no fight drawn")
	}
	one, board := g.arts, f.P.fightBoard
	for range 5 {
		f.Draw(dst)
	}
	if g.arts != one || f.P.fightBoard != board {
		t.Fatalf("idle draws rebuilt: %d Art reads, want %d", g.arts, one)
	}
	fight.Field.Units = append(fight.Field.Units, game.Unit{ID: 2, Member: -1, Foe: true})
	g.st = fight // the game's State is built anew: a new Version
	f.stale()
	f.Draw(dst)
	if f.P.fightBoard == board {
		t.Fatal("a change did not rebuild the board")
	}
}

// The 3D view's walls and the minimap's squares are derived from the State
// once per Version: idle frames read the same ones, a new State builds them again.
func TestWallsAndMinimapBuiltOncePerChange(t *testing.T) {
	st := func(x int) play.State {
		return play.Exploring{Common: play.Common{Where: game.Where{X: x}}, Map: play.ExploreMap{Automap: game.Automap{Size: image.Pt(16, 16)},
			Walls: play.Walls{Side: [][][4]int{{{1, 1, 1, 1}}}, Pool: map[int]string{1: "w"}}}}
	}
	g, f := playGame(st(1))
	var p Presenter
	p.SetEngine(g)
	dst := gpu.NewImage(640, 400)
	font := func(byte, bool) [8]byte { return [8]byte{} }
	w, m := p.walls(), p.eng.r.derived().minimap
	if w == nil || m == nil {
		t.Fatalf("walls %v, minimap %v", w, m)
	}
	for range 5 {
		p.scene()
		p.sceneName()
		p.drawMap(dst, egaInk, font)
	}
	if p.walls() != w || p.eng.r.derived().minimap != m {
		t.Fatal("idle frames derived them again")
	}
	f.st = st(2)
	p.scene()
	if p.walls() == w || p.eng.r.derived().minimap == m {
		t.Fatal("a new State did not derive them again")
	}
}

// The window border's pieces are decoded from the art once per Version:
// the frame's setup and every box ask for them again and again.
func TestBorderBuiltOncePerChange(t *testing.T) {
	g := &artCounting{fake: &fake{st: play.Exploring{}}}
	var p Presenter
	p.SetEngine(&classicGame{Game: NewGame(g, game.Data{})})
	p.windowBorder()
	one := g.arts
	for range 5 {
		p.windowBorder()
	}
	if g.arts != one {
		t.Fatalf("idle asks read the art again: %d reads, want %d", g.arts, one)
	}
	g.st = play.Exploring{Common: play.Common{Where: game.Where{X: 1}}}
	p.windowBorder()
	if g.arts != one+1 {
		t.Fatalf("a new State: %d art reads, want %d", g.arts, one+1)
	}
}
