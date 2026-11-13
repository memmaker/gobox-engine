package engine

import "path/filepath"

// The soundtrack (extra, the Esc menu's Audio "Soundtrack"): recordings in
// data/assets/soundtrack play by area and building (areaSong), in place
// of the Amiga's. They are free music from OpenGameArt, converted to Ogg
// Vorbis at 44100 Hz and levelled to -20 LUFS; CREDITS.txt there names them.
// A song without its file is silent. The recordings play in the presenter
// (present.Soundtrack); the rules choose the song.

const soundtrackDir = "soundtrack"

// SoundtrackFiles are the recordings' names (without .ogg) by song.
var SoundtrackFiles = map[int]string{
	songTitle: "title", songNewPhlan: "newphlan", songDungeon1: "dungeon1", songDungeon2: "dungeon2",
	songDiogenes: "diogenes", songDungeon3: "dungeon3", songYarash: "yarash", songGraveyard: "graveyard",
	songValjevo: "valjevo", songLair: "lair", songCombat: "combat", songTyranthraxus: "tyranthraxus",
	songWilderness: "wilderness", songShop: "shop", songEnding: "ending", songGameOver: "gameover",
	songEncounter: "encounter", songVictory: "victory", 19: "victory", songTemple: "temple", songInn: "inn",
	22: "traininghall", songTrainingHall: "traininghall", songCityHall: "cityhall", songTavern: "tavern",
	songGypsy: "gypsy", songBattleBegins: "battlebegins",
}

// SoundtrackDir is where the recordings are.
func (e *Engine) SoundtrackDir() string { return filepath.Join(e.adv, soundtrackDir) }

// SetSoundtrack turns the soundtrack on or off, restarting the song for
// where the party is in the music now chosen.
func (e *Engine) SetSoundtrack(on bool) {
	s := &e.snd
	s.stOn = on
	if s.d == nil {
		return // SetSound takes it up
	}
	n := 0
	if s.music < len(amigaSong) {
		n = amigaSong[s.music] // the Amiga's song's counterpart; areaMusic follows up
	}
	s.d.SetSoundtrack(on, n, s.music)
}
