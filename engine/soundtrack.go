package engine

import (
	"gobox/adventure"
	"path/filepath"
)

// The soundtrack (extra, the Esc menu's Audio "Soundtrack"): recordings in
// data/assets/soundtrack play by area and building (areaSong), in place
// of the Amiga's. They are free music from OpenGameArt, converted to Ogg
// Vorbis at 44100 Hz and levelled to -20 LUFS; CREDITS.txt there names them.
// A song without its file is silent. The recordings play in the presenter
// (present.Soundtrack); the rules choose the song.

const soundtrackDir = "soundtrack"

// SoundtrackFiles are the recordings' names (without .ogg) by song.
var SoundtrackFiles = map[int]string{
	adventure.SongTitle: "title", adventure.SongNewPhlan: "newphlan", adventure.SongDungeon1: "dungeon1", adventure.SongDungeon2: "dungeon2",
	adventure.SongDiogenes: "diogenes", adventure.SongDungeon3: "dungeon3", adventure.SongYarash: "yarash", adventure.SongGraveyard: "graveyard",
	adventure.SongValjevo: "valjevo", adventure.SongLair: "lair", adventure.SongCombat: "combat", adventure.SongTyranthraxus: "tyranthraxus",
	adventure.SongWilderness: "wilderness", adventure.SongShop: "shop", adventure.SongEnding: "ending", adventure.SongGameOver: "gameover",
	adventure.SongEncounter: "encounter", adventure.SongVictory: "victory", 19: "victory", adventure.SongTemple: "temple", adventure.SongInn: "inn",
	22: "traininghall", adventure.SongTrainingHall: "traininghall", adventure.SongCityHall: "cityhall", adventure.SongTavern: "tavern",
	adventure.SongGypsy: "gypsy", adventure.SongBattleBegins: "battlebegins",
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
