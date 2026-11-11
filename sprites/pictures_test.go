package sprites

import (
	"encoding/json"
	"maps"
	"testing"
	"testing/fstest"
)

func TestPicturesRoundTrip(t *testing.T) {
	for _, game := range []string{"por", "cotab", "sotsb", "pod", "buckrogers"} {
		p := GamePictures(game)
		if len(p) == 0 {
			t.Fatalf("%s: empty table", game)
		}
		b, err := json.Marshal(p)
		if err != nil {
			t.Fatal(err)
		}
		got, err := ReadPictures(fstest.MapFS{PicturesFile: {Data: b}})
		if err != nil || !maps.Equal(got, p) {
			t.Fatalf("%s: round trip differs (%v)", game, err)
		}
		for file, blocks := range Origins()[game] {
			for key, name := range blocks {
				if id := p.ID(parseRef(file, key)); string(id) != name {
					t.Fatalf("%s %s %s: %q, want %q", game, file, key, id, name)
				}
			}
		}
	}
}

// TestAmigaPictures: a game's Amiga archives (origins game-amiga) are in its
// table apart from the DOS ones of the same name (CBODY.DAX), round trip included.
func TestAmigaPictures(t *testing.T) {
	for _, game := range []string{"por", "cotab", "sotsb", "pod"} {
		p := GamePictures(game)
		b, _ := json.Marshal(p)
		got, err := ReadPictures(fstest.MapFS{PicturesFile: {Data: b}})
		if err != nil || !maps.Equal(got, p) {
			t.Fatalf("%s: round trip differs (%v)", game, err)
		}
		n := 0
		for file, blocks := range Origins()[game+originSuffix[PlatformAmiga]] {
			for key, name := range blocks {
				r := parseRef(file, key)
				r.Platform = PlatformAmiga
				if id := p.ID(r); string(id) != name {
					t.Fatalf("%s amiga %s %s: %q, want %q", game, file, key, id, name)
				}
				n++
			}
		}
		if n == 0 {
			t.Fatalf("%s: no Amiga pictures", game)
		}
	}
}
