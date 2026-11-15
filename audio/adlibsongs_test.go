package audio

import (
	"bytes"
	"os"
	"testing"

	"gobox/asset/orig"
)

// TestAdLibPlaysImportedSongs: a song given as its imported MIDI and voices
// (AdLibData.Songs, as the import's sound/music) sounds sample for sample as
// MUSIC.DAX's block; a MIDI that does not read is NewAdLib's error.
func TestAdLibPlaysImportedSongs(t *testing.T) {
	dax, err := os.ReadFile("testdata/MUSIC.DAX")
	if err != nil {
		t.Fatal(err)
	}
	blocks, err := orig.GlbrBlocks(dax)
	if err != nil {
		t.Fatal(err)
	}
	songs := map[int]orig.Song{}
	for id, b := range blocks {
		if mid, voices, err := orig.SongMIDI(b, 0); err == nil {
			songs[id] = orig.Song{MIDI: mid, Voices: voices}
		}
	}
	if len(songs) < 2 {
		t.Fatalf("%d songs", len(songs))
	}
	render := func(d AdLibData, n int) []byte {
		a, err := NewAdLib(d)
		if err != nil {
			t.Fatal(err)
		}
		a.Start(n)
		if a.Song() != n {
			t.Fatalf("song %d did not start", n)
		}
		m := make([]byte, 4*adlibRate) // a second
		a.Mix(m, false)
		return m
	}
	for id := range songs {
		want := render(AdLibData{Music: dax}, id)
		if got := render(AdLibData{Music: dax, Songs: songs}, id); !bytes.Equal(got, want) {
			t.Errorf("song %d: the MIDI sounds unlike the block", id)
		}
	}
	if _, err := NewAdLib(AdLibData{Music: dax, Songs: map[int]orig.Song{1: {MIDI: []byte("no midi")}}}); err == nil {
		t.Error("a broken MIDI was taken")
	}
}
