// pool.go starts a Gold Box game's pool engine (engine.New over
// rules1e) on an imported adventure folder: the one host loop
// every game on pool shares. What differs between games is the manifest's
// (importer.Manifest).
package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"gobox/asset"
	"gobox/audio"
	"gobox/engine"
	"gobox/game"
	"gobox/importer"
	"gobox/rules"
	"gobox/rules4e"
	"gobox/ui"
)

// tps: Update runs 120 times a second, so a key waits at most ~8 ms before
// the engine sees it.
const tps = 120

// Data is the engine's game.Data with the remastered art (asset.Remaster) of game id, if it has any.
func poolData(eng *engine.Engine, id string) game.Data {
	d := eng.Data()
	d.Remaster, d.AreaMaps = asset.Remaster(id)
	return d
}

// playPool starts the game of the adventure folder s.Adventure.
func playPool(s start) (ui.Launch, error) {
	s.Explore = s.Explore || s.Capture.Slot != "" // -slot starts play from that save (Explore reads it), -explore or not

	host := &ui.HostSettings // the shared settings in use; ui.Play loads them from SETTINGS.JSON
	f, err := importer.Read(s.Adventure)
	if err != nil {
		return ui.Launch{}, err
	}
	if f.Record == rules.NoRecord {
		return ui.Launch{}, fmt.Errorf("adventure %s: manifest has no character record layout: re-import it", f.Dir)
	}
	host.FullIntro = f.Manifest.Start != nil && f.Manifest.Start.Full // the adventure's default
	eng, err := engine.NewPool(s.Adventure, s.Data)
	if err != nil {
		return ui.Launch{}, err
	}
	if f.UserSaves {
		saves, err := homeSaves(f.ID, s.Adventure)
		if err != nil {
			return ui.Launch{}, err
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
		return ui.Launch{}, err
	}
	if err := eng.SetEnd(f.Manifest.End); err != nil {
		return ui.Launch{}, err
	}
	apply := func(h ui.HostOptions) { // the Esc menu's options (minimap, picture box, speed, fullscreen ...) kept across runs
		eng.Deskew.Store(h.Deskew)
		eng.Explicit.Store(h.Explicit)
		eng.QuickSearch.Store(h.QuickSearch)
		eng.DaySky.Store(h.DaySky)
		eng.FullIntro.Store(h.FullIntro && !s.Explore)
		if h.Combat4e && f.Combat4e {
			eng.HandIn(engine.PoolSettings(h.PathSearch, rules4e.New()))
		} else {
			eng.HandIn(engine.PoolSettings(h.PathSearch, nil))
		}
		eng.Dasp.Store(h.Dasp)
		eng.SetSoundtrack(h.Soundtrack)
		eng.NoSteps.Store(!h.Steps)
		engine.SetVolumes(h.Music, h.Effects, h.Voice)
	}
	tables, err := eng.SoundData()
	if err != nil {
		return ui.Launch{}, err
	}
	drv, err := audio.NewDriver(tables, eng.Cfg.Sound)
	if err != nil {
		return ui.Launch{}, err
	}
	am, fx, pack, speaking, vol := eng.SoundParts()
	drv.Attach(am, fx, ui.NewSoundtrack(eng.SoundtrackDir(), eng.SoundtrackFiles()), pack, speaking, vol) // the presenter decodes and plays the recordings
	eng.SetSound(drv)
	audio.Output(drv) // speaker and voice share one context
	audio.Output(eng.VoiceReader())
	eng.Cfg.Param = s.Param
	if s.Explore {
		eng.Explore = true
	}
	if err := devViews(eng, s); err != nil {
		return ui.Launch{}, err
	}
	eng.FreeTraining = s.Capture.Dev.FreeTrain
	seed := uint32(s.Seed)
	eng.FixedSeed = &seed

	data := poolData(eng, f.ID)
	data.OpenAt = s.Open
	data.OwnMusic = filepath.Join(s.Data, "music") // the player's own music, by trigger
	cg := ui.NewGame(eng, data)
	g := ui.NewFront(cg, data, drv) // the rules' data (series, spell tips, overland art); one driver: the rules' calls and the fight's events
	g.OnStart = func() { engine.DOSArt.Store(host.DOSArt) }
	g.P.Sound = eng.PlayEventSound // a fight's sounds with what they belong to
	w, h := window(f)
	return ui.Launch{
		Front: g,
		Title: f.Title,
		W:     w,
		H:     h,
		Art: &ui.Arts{ // the Amiga or the DOS art (F12, the Graphics page)
			Want:   func() bool { return host.DOSArt },
			Toggle: func() { host.DOSArt = !host.DOSArt },
			Set:    func(dos bool) { engine.DOSArt.Store(dos); g.ArtNew = true },
		},
		FontPrev: ui.KeyFontPrev,
		Apply:    apply,
		Voice:    eng.ToggleVoice,
		SaveLoad: func() { g.P.OpenSlots() }, // as F10
		MainMenu: func() { cg.EndGame() },    // through classic, so its view is fresh
		TPS:      tps,
	}, nil
}

// HomeSaves is the save folder of game id under the user's config dir
// (gobox/<id>); the first time it is seeded with the adventure adv's shipped
// characters and saves (its saves/); the adventure folder is never written.
func homeSaves(id, adv string) (string, error) {
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
		return engine.ReadRoster(other.Record, filepath.Join(cfg, "gobox", from))
	}
}
