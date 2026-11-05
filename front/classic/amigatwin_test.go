package classic

import (
	"strings"
	"testing"

	"gobox/sprites"
)

// Every POR DOS portrait and combat icon has its Amiga twin (but set 6's heads 1
// and 3, which the Amiga lacks), and poolImage draws it only while the Amiga art shows.
func TestAmigaTwin(t *testing.T) {
	GameID = "por"
	defer func() { GameID, AmigaArt = "", nil }()
	o := sprites.Origins()["por"]
	for f, refs := range o {
		if !strings.HasPrefix(f, "HEAD") && !strings.HasPrefix(f, "BODY") && f != "CHEAD.DAX" && f != "CBODY.DAX" {
			continue
		}
		for r, n := range refs {
			if f == "HEAD6.DAX" && (r == "1" || r == "3") {
				continue
			}
			if tw := amigaTwin(n); !strings.HasPrefix(tw, "amiga_") {
				t.Errorf("%s %s (%s): twin %q", f, r, n, tw)
			}
		}
	}
	n := o["HEAD3.DAX"]["0"]
	dos := sprites.Image(Series, "portraits", n)
	if poolImage("portraits", n) != dos {
		t.Fatal("DOS art: not the DOS picture")
	}
	AmigaArt = func() bool { return true }
	if im := poolImage("portraits", n); im == nil || im == dos {
		t.Fatal("Amiga art: not the Amiga twin")
	}
}
