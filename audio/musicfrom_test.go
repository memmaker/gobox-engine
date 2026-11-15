package audio

import "testing"

// fakeSongs is a Tracks and an Amiga player that only remember their song.
type fakeSongs struct{ song int }

func (f *fakeSongs) Len() int           { return 9 }
func (f *fakeSongs) Has(n int) bool     { return n > 0 }
func (f *fakeSongs) Song() int          { return f.song }
func (f *fakeSongs) Reset()             { f.song = 0 }
func (f *fakeSongs) Start(n int)        { f.song = n }
func (f *fakeSongs) Playing() bool      { return f.song != 0 }
func (f *fakeSongs) Mix([]byte, bool)   {}
func (f *fakeSongs) Effect(int)         {}
func (f *fakeSongs) Render([]byte, int) {}
func (f *fakeSongs) Sounding() bool     { return f.song != 0 }
func (f *fakeSongs) amigaSong(n int)    { f.song = n }

type fakeAmiga struct{ *fakeSongs }

func (f fakeAmiga) Song(n int) { f.amigaSong(n) }

// TestMusicFrom: a game with both releases' music plays its songs on the one
// picked, and the song playing moves over when the pick changes.
func TestMusicFrom(t *testing.T) {
	d := newDriver(t, 'T')
	dos, amiga := &fakeSongs{}, &fakeSongs{}
	d.Attach(fakeAmiga{amiga}, nil, nil, nil, nil, nil)
	d.SetTracks(dos)
	if d.Song(3, 0, false) {
		t.Fatal("the DOS songs took a song with the Amiga's picked")
	}
	d.AmigaSong(3)
	if amiga.song != 3 || dos.song != 0 {
		t.Fatalf("Amiga picked: amiga %d dos %d", amiga.song, dos.song)
	}
	d.MusicFrom(true)
	if amiga.song != 0 || dos.song != 3 {
		t.Fatalf("DOS picked: amiga %d dos %d, want the song moved over", amiga.song, dos.song)
	}
	if !d.Song(5, 0, false) || dos.song != 5 {
		t.Fatalf("DOS picked: song 5 not taken (%d)", dos.song)
	}
	d.MusicFrom(false)
	if amiga.song != 5 || dos.song != 0 {
		t.Fatalf("Amiga picked again: amiga %d dos %d", amiga.song, dos.song)
	}
}
