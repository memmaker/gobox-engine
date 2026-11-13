package importer

import (
	"gobox/adventure"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// TestEndPictures: the End's pictures are in the imported art with the frames it counts
// (art/dos/PIC6.txt, BIGPIC6.txt).
func TestEndPictures(t *testing.T) {
	frames := func(file string, block int) int {
		b, err := os.ReadFile(filepath.Join("..", "..", "data", "cotab", "sources", "assets", "art", "dos", file+".txt"))
		if err != nil {
			t.Skip("no imported art:", err)
		}
		n, in := 0, false
		for _, l := range strings.Split(string(b), "\n") {
			if m := regexp.MustCompile(`^block (\d+) `).FindStringSubmatch(l); m != nil {
				in = m[1] == strconv.Itoa(block)
			} else if in && strings.HasPrefix(l, "frame ") {
				n++
			}
		}
		return n
	}
	steps := cotabEnd().Steps
	if len(steps) != 5 {
		t.Fatalf("%d steps", len(steps))
	}
	for _, st := range steps {
		if len(st.Text) == 0 {
			t.Errorf("step %v says nothing", st.Kind)
		}
		if len(st.Layers) == 0 { // a text step
			continue
		}
		l := st.Layers[0]
		if want := frames(l.File, l.Block); want == 0 || (len(l.Frames) != 0 && len(l.Frames) != want) {
			t.Errorf("%s %d: %d frames in the art, %d listed", l.File, l.Block, want, len(l.Frames))
		}
	}
}

// TestEndEffects: the ending heals the party and opens the hall, and says it is won.
func TestEndEffects(t *testing.T) {
	e := cotabEnd()
	if !slices.Equal(e.After, []adventure.Effect{adventure.EffectHealParty, adventure.EffectOpenHall}) || e.Note != endNote {
		t.Fatalf("After %v, Note %q", e.After, e.Note)
	}
}
