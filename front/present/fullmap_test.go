package present

import (
	"image"
	"slices"
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// mapState is the party out on a 16x16 map; marked has the player's own mark on (3,4).
var (
	mapState = play.Exploring{Common: play.Common{Automap: game.Automap{Size: image.Pt(16, 16)}}}
	marked   = play.Exploring{Common: play.Common{Automap: game.Automap{Size: image.Pt(16, 16), Marked: []image.Point{{3, 4}}, Notes: []string{"hi"}, NoteKinds: []string{""}}}}
)

// TestFullMapMarkUnmark: the cell menu's Mark types a label that Enter sends
// to Game.Mark; Remove calls Game.Unmark.
func TestFullMapMarkUnmark(t *testing.T) {
	g, f := playGame(mapState)
	p := &Presenter{}
	p.SetEngine(g)
	p.fullMap = &fullMapUI{cursor: image.Pt(3, 4)}
	p.mapPick(1) // Walk, Mark, Remove
	for _, k := range []uint16{'h', 'i', 13} {
		p.mapKey(k)
	}
	for _, c := range p.Calls() {
		c(g)
	}
	f.st = marked // the mark made
	p.fullMap.menu = true
	p.mapPick(1) // Walk, Remove: a labelled square has no Mark
	for _, c := range p.Calls() {
		c(g)
	}
	want := []play.Action{play.Mark{At: image.Pt(3, 4), Text: "hi"}, play.Unmark{At: image.Pt(3, 4)}}
	if !slices.Equal(f.sent, want) {
		t.Fatalf("sent %v, want %v", f.sent, want)
	}
}
