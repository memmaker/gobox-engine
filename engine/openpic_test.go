package engine

import (
	"gobox/game"
	"strings"
	"testing"
)

// eventPic is s's scripted picture up (PIC, or a HEAD/BODY portrait's
// head), "" none.
func eventPic(s game.State) string {
	c := commonOf(s)
	if pt := c.Portrait; pt != nil {
		return pt.Head.File
	}
	if p := c.Picture; p != nil && strings.HasPrefix(strings.ToUpper(p.File), "PIC") {
		return p.File
	}
	return ""
}

// TestOpeningPicture: a new game's opening event (Rolf's tour of Phlan)
// has its picture in State while its page is up, and the picture is gone
// once the party is back to exploring.
func TestOpeningPicture(t *testing.T) {
	e := quickStarted(t)
	e.Start(e.GameLoop)
	seen := ""
	for range 200 {
		e.stepOn(&game.Outcome{}, nil)
		s := e.build()
		p := eventPic(s)
		if x, ok := s.(game.Exploring); p != "" {
			seen = p
		} else if seen != "" && ok && x.Ask == nil {
			return
		}
		press(e)
	}
	if seen == "" {
		t.Fatal("the opening event showed no picture in State")
	}
	t.Fatalf("the picture %s stayed after the event", seen)
}

// paneOf is the text pane's lines, each its runs' text.
func paneOf(e *Engine) []string {
	var out []string
	if p := e.statePage(); p != nil {
		for _, s := range p.Pane {
			if s = strings.TrimRight(s, " "); s != "" {
				out = append(out, s)
			}
		}
	}
	return out
}

// TestPrintContinues: PRINT goes on at the text cursor, on the same line;
// PRINTCLEAR empties the pane first.
func TestPrintContinues(t *testing.T) {
	e := quickStarted(t)
	e.scriptText(true, "THE DOOR ")
	e.scriptText(false, "OPENS.")
	if got := paneOf(e); len(got) != 1 || got[0] != "THE DOOR OPENS." {
		t.Errorf("after PRINT: %q, want one line \"THE DOOR OPENS.\"", got)
	}
	e.scriptText(true, "GONE.")
	if got := paneOf(e); len(got) != 1 || got[0] != "GONE." {
		t.Errorf("after PRINTCLEAR: %q, want just \"GONE.\"", got)
	}
}

// TestSkyDay: Flags.Day follows the sky the original draws (8, dark grey,
// is night, at any hour), not the clock.
func TestSkyDay(t *testing.T) {
	e := quickStarted(t)
	e.world.In3D = 1
	e.state().Area.Clock[3], e.state().Area.Clock[2] = 1, 2 // 1:20, night by the clock: the sky decides
	for sky, want := range map[uint16]bool{11: true, 15: true, 8: false} {
		e.world.Sky = sky
		if got := e.flagDay(false); got != want {
			t.Errorf("sky %d: day %v, want %v", sky, got, want)
		}
	}
	e.world.SkyInside = 8
	if e.world.Sky = 11; e.flagDay(true) {
		t.Error("inside under a night SkyInside: day")
	}
}
