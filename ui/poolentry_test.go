package ui

import (
	"image"
	"strings"
	"testing"

	"gobox/asset"
)

// Every pool entry is unique, resolves to a picture under both arts, and an
// entry with a twin shows the DOS picture under DOS and the twin under Amiga
// whichever art showed when it was picked; a legacy amiga_ name loads to an
// entry with that twin. IDs come sorted, so SpriteNames order never matters.
func TestPoolEntriesArtNeutral(t *testing.T) {
	same := func(a, b image.Image) bool { return a != nil && a == b }
	for kind, prefixes := range map[string][]string{"portraits": {"head_", "body_"}, "combat": {"combat_head_", "combat_body_"}} {
		for _, prefix := range prefixes {
			ids := poolEntries(kind, prefix)
			seen := map[string]bool{}
			var both, dosOnly, amigaOnly int
			for _, id := range ids {
				if seen[id] {
					t.Errorf("%s: duplicate ID", id)
				}
				seen[id] = true
				if entryPicture(kind, id, false) == nil || entryPicture(kind, id, true) == nil {
					t.Errorf("%s: no picture under one art", id)
				}
				tw := amigaTwin(id)
				switch {
				case strings.HasPrefix(id, amigaPrefix):
					amigaOnly++
				case tw == "":
					dosOnly++
				default:
					both++
					if l := poolEntryID(tw); l != id && amigaTwin(l) != tw { // several DOS pictures may share one twin: it loads to the first

						t.Errorf("%s: legacy %s loads to %s", id, tw, l)
					}
					if !same(entryPicture(kind, tw, false), asset.SpriteImage(Series, kind, poolEntryID(tw))) {
						t.Errorf("%s: picked under Amiga, DOS does not show the DOS picture", id)
					}
					if !same(entryPicture(kind, id, true), asset.SpriteImage(Series, kind, tw)) {
						t.Errorf("%s: picked under DOS, Amiga does not show %s", id, tw)
					}
				}
			}
			t.Logf("%s %s: %d entries, %d both, %d DOS-only, %d Amiga-only", kind, prefix, len(ids), both, dosOnly, amigaOnly)
		}
	}
}
