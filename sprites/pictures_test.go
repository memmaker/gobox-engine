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
