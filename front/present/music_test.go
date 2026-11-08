package present

import (
	"gobox/front/classic"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"gobox/game"
)

// A place's own trigger plays over the generic one, which plays where it has none.
func TestOwnMusicTriggers(t *testing.T) {
	town := classic.MusicCue{Mode: game.ModeExplore, Where: game.Where{Mode: "town", Music: "new-phlan", Hour: 12}}
	if c := musicChain(town); !slices.Equal(c, []string{"new-phlan", "peaceful-town", "exploring"}) {
		t.Fatalf("town chain %v", c)
	}
	town.Fighting = true
	if c := musicChain(town); !slices.Equal(c, []string{"combat-new-phlan", "combat-town", "combat"}) {
		t.Fatalf("combat chain %v", c)
	}
	dir := t.TempDir()
	m := newOwnMusic(dir, []string{"new-phlan"})
	for _, d := range []string{"peaceful-town", "new-phlan", "combat"} {
		if _, err := os.Stat(filepath.Join(dir, d)); err != nil {
			t.Fatalf("no folder %s", d)
		}
	}
	m.pools["peaceful-town"] = []string{"a.ogg"}
	m.Want(musicChain(classic.MusicCue{Where: game.Where{Mode: "town", Music: "slums", Hour: 12}}))
	if m.trig != "peaceful-town" || !m.Playing() {
		t.Fatalf("generic: %q", m.trig)
	}
	m.pools["new-phlan"] = []string{"b.mp3"}
	m.Want(musicChain(classic.MusicCue{Where: game.Where{Mode: "town", Music: "new-phlan", Hour: 12}}))
	if m.trig != "new-phlan" {
		t.Fatalf("specific: %q", m.trig)
	}
	m.Want(nil)
	if m.Playing() {
		t.Fatal("no trigger: the game's music again")
	}
}
