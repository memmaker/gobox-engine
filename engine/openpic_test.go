package engine

import (
	"gobox/game"
	"strings"
	"testing"
)

// eventPic is the first layer of s's page drawn from a scripted picture
// (PIC, or a HEAD/BODY portrait), "" none.
func eventPic(s game.State) string {
	layers := commonOf(s).Layers
	if sc, ok := s.(game.Scene); ok {
		layers = sc.Layers
	}
	for _, l := range layers {
		if f := strings.ToUpper(l.File); strings.HasPrefix(f, "PIC") || strings.HasPrefix(f, "HEAD") {
			return l.File
		}
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
		for _, ln := range p.Pane {
			s := ""
			for _, r := range ln {
				s += r.Text
			}
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
	e.state().Area.In3D = 1
	e.state().Area.Clock[3], e.state().Area.Clock[2] = 1, 2 // 1:20, night by the clock: the sky decides
	for sky, want := range map[uint16]bool{11: true, 15: true, 8: false} {
		e.state().Area.Sky = sky
		if got := e.state().SkyDay(false); got != want {
			t.Errorf("sky %d: day %v, want %v", sky, got, want)
		}
	}
	e.state().Area.SkyInside = 8
	if e.state().Area.Sky = 11; e.state().SkyDay(true) {
		t.Error("inside under a night SkyInside: day")
	}
}
