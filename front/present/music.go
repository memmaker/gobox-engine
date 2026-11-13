// music.go: the player's own music. data/music/<trigger>/ holds .mp3 and
// .ogg files; the most specific trigger with files plays, a random song of
// its pool at a time, over the game's silenced songs.

package present

import (
	"errors"
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"gobox/audio"
	"gobox/game"
)

// musicTriggers are gobox's generic triggers, made as folders for every game.
var musicTriggers = []string{
	"title", "party-creation",
	"peaceful-town", "peaceful-town-night", "dungeon", "wilderness", "wilderness-night", "exploring",
	"combat", "combat-town", "combat-dungeon", "combat-wilderness", "victory",
	"camp", "shop", "temple", "training-hall", "vault",
}

// musicChain is the triggers for s, most specific first: the place's own
// (Where.Music) before the generic.
func musicChain(s MusicCue) []string {
	place := s.Where.Music
	kind := "exploring"
	switch m := s.Where.Mode; {
	case m == "town" || m == "city":
		kind = "peaceful-town"
	case m == "dungeon" || m == "" && s.Where.Inside:
		kind = "dungeon"
	case m == "wilderness" || m == "overland" || m == "outdoors":
		kind = "wilderness"
	}
	night := !s.Where.Inside && (s.Where.Hour >= 20 || s.Where.Hour < 6) && kind != "dungeon"
	var c []string
	add := func(t ...string) {
		for _, x := range t {
			if x != "" && !strings.HasPrefix(x, "-") && !slices.Contains(c, x) {
				c = append(c, x)
			}
		}
	}
	explore := func() {
		if night {
			add(place + "-night")
		}
		add(place)
		if night {
			add(kind + "-night")
		}
		add(kind, "exploring")
	}
	switch {
	case s.Title:
		add("title")
	case s.Creating:
		add("party-creation", "title")
	case s.Results:
		add("victory")
	case s.Fighting:
		ck := map[string]string{"peaceful-town": "town", "dungeon": "dungeon", "wilderness": "wilderness"}[kind]
		if s.Dungeon {
			ck = "dungeon"
		}
		add("combat-"+place, "combat-"+ck, "combat")
	case s.Mode == game.ModeShop:
		add("shop-"+place, "shop")
		explore()
	case s.Mode == game.ModeTemple:
		add("temple-"+place, "temple")
		explore()
	case s.Mode == game.ModeTraining:
		add("training-hall-"+place, "training-hall")
		explore()
	case s.Mode == game.ModeVault:
		add("vault")
		explore()
	case s.Mode == game.ModeCamp:
		add("camp-"+place, "camp")
		explore()
	default:
		explore()
	}
	return c
}

// ownMus is the player's own music (nil: none, headless).
var ownMus *ownMusic

// ownMusic plays the pools of data/music. Safe from the game and the audio thread.
type ownMusic struct {
	mu    sync.Mutex
	pools map[string][]string // trigger -> its files
	trig  string              // the trigger playing ("": none)
	song  string              // the file pcm is (or is being decoded) for
	pcm   []byte              // s16 stereo, nil while decoding
	pos   int
}

// newOwnMusic makes dir's trigger folders (gobox's and places) where they
// are missing and finds their files.
// A folder that cannot be made (a read-only data folder) is the error returned
// with the music still found in the folders there are.
func newOwnMusic(dir string, places []string) (*ownMusic, error) {
	m := &ownMusic{pools: map[string][]string{}}
	var failed error
	for _, t := range append(slices.Clone(musicTriggers), places...) {
		failed = errors.Join(failed, os.MkdirAll(filepath.Join(dir, t), 0o755))
	}
	es, _ := os.ReadDir(dir)
	for _, e := range es {
		fs, _ := os.ReadDir(filepath.Join(dir, e.Name()))
		for _, f := range fs {
			if x := strings.ToLower(filepath.Ext(f.Name())); !f.IsDir() && (x == ".mp3" || x == ".ogg") {
				m.pools[e.Name()] = append(m.pools[e.Name()], filepath.Join(dir, e.Name(), f.Name()))
			}
		}
	}
	return m, failed
}

// Want plays the first trigger of chain with files, a song of the trigger
// playing going on; none: silence, the game's songs heard again. It is the
// trigger now playing ("": none).
func (m *ownMusic) Want(chain []string) string {
	t := ""
	for _, c := range chain {
		if len(m.pools[c]) > 0 {
			t = c
			break
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if t != m.trig {
		m.trig = t
		m.next()
	}
	return t
}

// next starts a random song of the trigger's pool, another than the last when it can (mu held).
func (m *ownMusic) next() {
	m.pcm, m.pos, m.song = nil, 0, ""
	pool := m.pools[m.trig]
	if len(pool) == 0 {
		return
	}
	song := pool[rand.IntN(len(pool))]
	if len(pool) > 1 {
		for song == m.song {
			song = pool[rand.IntN(len(pool))]
		}
	}
	m.song = song
	go func() {
		b, err := os.ReadFile(song)
		var pcm []byte
		if err == nil {
			pcm, _ = audio.MusicPCM(song, b)
		}
		if pcm == nil {
			pcm = []byte{} // unreadable: it ends at once, the next one plays
		}
		m.mu.Lock()
		if m.song == song && m.pcm == nil {
			m.pcm = pcm
		}
		m.mu.Unlock()
	}()
}

// Playing: a trigger's pool plays (the game's songs are silent).
func (m *ownMusic) Playing() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.trig != ""
}

// Read renders the music (s16 stereo, 44100 Hz) at half level, the next
// song of the pool after each one's end.
func (m *ownMusic) Read(p []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	clear(p)
	for i := 0; i+1 < len(p) && m.pcm != nil; i += 2 {
		if m.pos+1 >= len(m.pcm) {
			last := m.song
			m.next()
			if m.song == "" || m.song == last && len(m.pcm) == 0 {
				break
			}
			continue // decoding: silence until it is ready
		}
		v := int16(uint16(m.pcm[m.pos]) | uint16(m.pcm[m.pos+1])<<8)
		v /= 2 // ponytail: a fixed half level, a slider when there is a music volume option
		p[i], p[i+1] = byte(v), byte(uint16(v)>>8)
		m.pos += 2
	}
	return len(p), nil
}
