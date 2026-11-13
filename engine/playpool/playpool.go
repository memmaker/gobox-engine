// Package playpool starts a Gold Box game's pool engine (engine.New over
// rules1e) on an imported adventure folder for cmd/gobox: the one host loop
// every game on pool shares. What differs between games is the manifest's
// (importer.Manifest) or Extras.
package playpool

import (
	"bytes"
	"errors"
	"fmt"
	"gobox/rules1e"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"gobox/app"
	"gobox/asset"
	"gobox/audio"
	"gobox/engine"
	"gobox/game"
	"gobox/importer"
	"gobox/rules"
	"gobox/ui"
)

// tps: Update runs 120 times a second, so a key waits at most ~8 ms before
// the engine sees it.
const tps = 120

// Extras are what a game adds in code, the rest of its host settings being
// its manifest's (importer.Manifest: title, window, saves, foreign rosters).
type Extras struct {
	Combat4e rules1e.CombatRules // the Gameplay page's "4e Combat" (nil: the option does nothing)
}

// Data is the engine's game.Data with the remastered art (asset.Remaster) of game id, if it has any.
func Data(eng *engine.Engine, id string) game.Data {
	d := eng.Data()
	d.Remaster, d.AreaMaps = asset.Remaster(id)
	return d
}

// Play runs the game of the adventure folder s.Adventure until the window closes.
func Play(s app.Start, gm Extras) error {
	s.Explore = s.Explore || s.Capture.Slot != "" // -slot starts play from that save (Explore reads it), -explore or not

	host := ui.DefaultHostOptions // the settings in use; app.Menu reads them from SETTINGS.JSON
	f, err := importer.Read(s.Adventure)
	if err != nil {
		return err
	}
	if f.Record == rules.NoRecord {
		return fmt.Errorf("adventure %s: manifest has no character record layout: re-import it", f.Dir)
	}
	host.FullIntro = f.Manifest.Start != nil && f.Manifest.Start.Full // the adventure's default
	ui.Host = &host.Options
	eng, err := engine.NewPool(s.Adventure, s.Data)
	if err != nil {
		return err
	}
	if f.UserSaves {
		saves, err := HomeSaves(f.ID, s.Adventure)
		if err != nil {
			return err
		}
		eng.Cfg.SaveDir = saves
	}
	if d := s.Capture.Dev.Saves; d != "" {
		eng.Cfg.SaveDir = d
	}
	if len(f.Foreign) > 0 {
		eng.Cfg.Foreign = homeRosters(filepath.Dir(s.Adventure), f.Foreign)
	}
	if err := eng.SetStart(f.Manifest.Start); err != nil {
		return err
	}
	if err := eng.SetEnd(f.Manifest.End); err != nil {
		return err
	}
	apply := func() { // the Esc menu's options (minimap, picture box, speed, fullscreen ...) kept across runs
		h := host
		eng.Deskew.Store(h.Deskew)
		eng.Explicit.Store(h.Explicit)
		eng.QuickSearch.Store(h.QuickSearch)
		eng.DaySky.Store(h.DaySky)
		eng.FullIntro.Store(h.FullIntro && !s.Explore)
		rs := rules1e.Settings{PathSearch: h.PathSearch}
		if h.Combat4e {
			rs.Combat = gm.Combat4e
		}
		eng.HandIn(rs)
		eng.Dasp.Store(h.Dasp)
		eng.SetSoundtrack(h.Soundtrack)
		eng.NoSteps.Store(!h.Steps)
		engine.SetVolumes(h.Music, h.Effects, h.Voice)
	}
	tables, err := eng.SoundData()
	if err != nil {
		return err
	}
	drv, err := audio.NewDriver(tables, eng.Cfg.Sound)
	if err != nil {
		return err
	}
	am, fx, pack, speaking, vol := eng.SoundParts()
	drv.Attach(am, fx, ui.NewSoundtrack(eng.SoundtrackDir(), engine.SoundtrackFiles), pack, speaking, vol) // the presenter decodes and plays the recordings
	eng.SetSound(drv)
	audio.Output(drv) // speaker and voice share one context
	audio.Output(eng.VoiceReader())
	eng.Cfg.Param = s.Param
	if s.Explore {
		eng.Explore = true
	}
	capt := s.Capture
	if capt.PNG != "" { // the -png picture the portrait dev scenes show
		b, err := os.ReadFile(capt.PNG)
		if err == nil {
			_, err = png.Decode(bytes.NewReader(b))
		}
		if err != nil {
			return fmt.Errorf("-png: %w", err)
		}
		eng.DevPortrait = b
	}
	if s.Dev != "" {
		eng.DevAction = eng.DevScene(s.Dev)
	}
	if s.Battle {
		eng.DevAction = eng.DevScene("autocombat")
	}
	dev := s.Capture.Dev
	if dev.Fight > 0 {
		eng.DevMonster = dev.Fight
		eng.DevAction = eng.DevScene("combat")
	}
	if a := dev.At; a != nil {
		eng.DevAction = func() { eng.DevAt(a.Block, a.X, a.Y) }
	}
	eng.FreeTraining = dev.FreeTrain
	if dev.ECLTrace {
		eng.TraceScripts()
	}
	if s.Seed >= 0 {
		seed := uint32(s.Seed)
		eng.FixedSeed = &seed
	}

	data := Data(eng, f.ID)
	data.OpenAt = s.Open
	data.OwnMusic = filepath.Join(s.Data, "music") // the player's own music, by trigger
	cg := ui.NewGame(eng, data)
	g := ui.NewFront(cg, data, drv) // the rules' data (series, spell tips, overland art); one driver: the rules' calls and the fight's events
	g.OnStart = func() { engine.DOSArt.Store(host.DOSArt) }
	g.P.Sound = eng.PlayEventSound // a fight's sounds with what they belong to
	ui.Scripted = capt.Scripted()
	return app.Run(app.Config{
		Front: g,
		Art: &app.Art{ // the Amiga or the DOS art (F12, the Graphics page)
			Want:   func() bool { return host.DOSArt },
			Toggle: func() { host.DOSArt = !host.DOSArt },
			Set:    func(dos bool) { engine.DOSArt.Store(dos); g.ArtNew = true },
		},
		FontPrev: ui.KeyFontPrev,
		Menu: &app.Menu{
			Prefs:    app.Prefs{Dir: eng.Cfg.SaveDir},
			Extra:    &host,
			Rows:     ui.HostRows(&host.Options, ui.HostOptionRows(&host, eng.ToggleVoice)...),
			Apply:    apply,
			SaveLoad: func() { g.P.OpenSlots() }, // as F10
			MainMenu: func() { cg.EndGame() },    // through classic, so its view is fresh
		},
		Theme:    app.Theme{Title: f.Title, W: f.Window.W, H: f.Window.H},
		Settings: s.Settings,
		Capture:  capt,
		TPS:      tps,
	})
}

// HomeSaves is the save folder of game id under the user's config dir
// (gobox/<id>); the first time it is seeded with the adventure adv's shipped
// characters and saves (its saves/); the adventure folder is never written.
func HomeSaves(id, adv string) (string, error) {
	cfg, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	home := filepath.Join(cfg, "gobox", id)
	if _, err := os.Stat(home); err == nil {
		return home, nil
	}
	if err := os.MkdirAll(home, 0o755); err != nil {
		return "", err
	}
	if err := os.CopyFS(home, os.DirFS(filepath.Join(adv, "saves"))); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}
	return home, nil
}

// homeRosters is a pool.Config.Foreign over the saves of the games named in
// names under the user's config dir (gobox/<id>), each read by the character
// record layout its adventure folder root/<id> declares.
func homeRosters(root string, names []string) func(string) ([]rules.CharacterFile, error) {
	return func(from string) ([]rules.CharacterFile, error) {
		if !slices.Contains(names, from) {
			return nil, fmt.Errorf("no roster known of game %q", from)
		}
		other, err := importer.Read(filepath.Join(root, from))
		if err != nil {
			return nil, err
		}
		cfg, err := os.UserConfigDir()
		if err != nil {
			return nil, err
		}
		return rules1e.ReadRoster(other.Record, filepath.Join(cfg, "gobox", from))
	}
}
