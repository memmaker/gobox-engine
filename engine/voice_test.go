package engine

import (
	"testing"

	"gobox/game"
	"gobox/ui"
)

// voxLog counts the recordings queued and the silences.
type voxLog struct{ clips, stops int }

func (s *voxLog) Say([]byte) error { s.clips++; return nil }
func (s *voxLog) Stop()            { s.stops++ }

// TestJournalPageSpoken: the journal page gobox lays out is voiced through
// Speak: Journal Entry 11's recording (the whole entry, under its title) is
// queued after the page voiced before is silenced.
func TestJournalPageSpoken(t *testing.T) {
	e := testEngine(t)
	e.Cfg.SaveDir = t.TempDir() // no VOICE.OFF
	vox := &voxLog{}
	e.vox = vox
	const title = "Journal Entry 11"
	for _, en := range e.entries() {
		if en.Title != title {
			continue
		}
		keys := ui.SpeakKeys(e.journalReadTexts(en.Book, []string{en.Title}, en.Voiced), nil, 0, 40, 25)
		if len(keys) != 1 || keys[0] != "Journal_Entry_11" {
			t.Fatalf("page 0's clips %v, want the entry's", keys)
		}
		if _, err := e.Do(game.Speak{Keys: keys}); err != nil {
			t.Fatal(err)
		}
		if vox.clips != 1 || vox.stops != 1 {
			t.Fatalf("%d recordings queued after %d silences, want 1 after 1", vox.clips, vox.stops)
		}
		return
	}
	t.Fatal("no entry", title)
}
