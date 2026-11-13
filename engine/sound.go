package engine

import (
	"fmt"
	"os"
	"path/filepath"

	"gobox/asset/orig"
	"gobox/audio"
	"gobox/audio/sounddata"
	"gobox/game/play"
	"gobox/rules"
)

// SoundFile is the driver's songs, scripts and vibrato waveforms, decoded from
// START.EXE at import (ImportData, sounddata.Decode at soundLayout) and
// synthesized at run time.
const SoundFile = "sound.json"

// SoundLayout is where START.EXE's driver (segment 08AE) keeps its tables (gobox audio API.md).
var SoundLayout = sounddata.Layout{Segment: 0x8AE, JumpTable: 545, SpeakerSet: 557, TandySet: 967,
	SpeakerSongs: 1375, TandySongs: 1543, Songs: 21, Enabled: 525,
	Commands: map[int]string{5710: "reset", 5804: "return", 5811: "call", 5701: "switch", 5678: "loop", 5759: "set"}}

// The sound driver of START.EXE (its segment 08AE) is gobox audio's Driver,
// over the tables in soundFile, which the rules call as Sound;
// this file keeps the driver's calls as the game makes them (the 0.1 s cut,
// Ctrl-S, the Amiga effects, the soundtrack and the sound pack). Op 0 of the program ends a
// script: the Tandy's song 18 (sound 19, which the game never plays) returns without
// a call, which in the original runs on into the driver's own code.
//
// Nothing is audible, and nothing waits, until SetSound is called, so the
// headless tools stay silent and fast.

const sndRate = 44100

// The game's sound card (DS:5372).
const (
	sndTandy = iota
	sndSpeaker
	sndNone
)

// sound is the rules' side of the driver: the calls go to d (nil until
// SetSound), DS:5372..5374 stay here.
type sound struct {
	d     audio.Sound
	typ   int  // DS:5372: sndTandy, sndSpeaker or sndNone
	saved int  // DS:5373: typ before Ctrl-S turned the sound off
	wait  int  // DS:5374
	music int  // the Amiga song Music started last
	stOn  bool // the Soundtrack option
	dasp  *dasp
}

// amigaFx maps PlaySound 2..12 to Amiga effects: the Amiga's effect table parallels the
// DOS sound numbers.
var amigaFx = [...]int{5, 9, 8, 10, 12, 11, 6, 11, 4, 7, 0}

// Music = the Amiga's playMusic (0x62E54): 1 title and camp (loops), 2 battle, 3 party
// menu, 5 party killed (one-shot jingles of ~7.7 s), 0 stop. These are its only nine
// callers (the intro, a separate program, has no player); songs 4 (15 s jingle) and
// 6 (looping) are in the player's data but the program never starts them, and the
// table's slot 7 points into note data, so no song), and any effect takes the channels from the music. The
// battle song is the word at 0x8A8C0 (2, never written) unless the byte at 0x8D2D1
// (never written, so 0) is set. The DOS game has no music.
func (e *Engine) Music(n int) {
	s := &e.snd
	if n >= 0 && n < len(amigaSong) && e.songMusic(amigaSong[n], 0) {
		s.music = n
		return
	}
	if s.d != nil {
		s.d.AmigaSong(n)
	}
	s.music = n
}

// songMusic (extra) plays soundtrack song n (0 stops), then song next
// when n is a one-shot that ends; false without the soundtrack, so the
// caller falls back to the Amiga's. A Battle Begins starts again.
func (e *Engine) songMusic(n, next int) bool {
	return e.snd.d != nil && e.snd.d.Song(n, next, n == songBattleBegins)
}

// titleMusic reports whether the title song plays, which plays on into the
// main menu.
func (e *Engine) titleMusic() bool {
	if e.snd.d == nil {
		return false
	}
	on, song := e.snd.d.Track()
	return on && song == songTitle
}

// areaMusic (extra; the Amiga program never starts songs 4 and 6): the
// overland map plays song 6, starting it again when an effect or another
// song took the channels; leaving the map stops it.
func (e *Engine) areaMusic() {
	s := &e.snd
	if s.d == nil {
		return
	}
	if on, cur := s.d.Track(); on {
		if want := e.areaSong(); want != 0 && want != cur {
			e.songMusic(want, 0)
		}
		return
	}
	going := s.d.Sounding()
	switch {
	case e.viewMode > 1 && (s.music != 6 || !going):
		e.Music(6)
	case e.viewMode <= 1 && s.music == 6:
		e.Music(0)
	}
}

// The soundtrack's songs, numbered in the Famicom cartridge's order.
const (
	songTitle        = 1
	songNewPhlan     = 2
	songDungeon1     = 3
	songDungeon2     = 4
	songDiogenes     = 5
	songDungeon3     = 6
	songYarash       = 7
	songGraveyard    = 8
	songValjevo      = 9 // Stojanow Gate and Valjevo Castle
	songLair         = 10
	songCombat       = 11
	songTyranthraxus = 12
	songWilderness   = 13
	songShop         = 14
	songEnding       = 15
	songGameOver     = 16
	songEncounter    = 17
	songVictory      = 18 // with the fanfare; 19 is the loop alone
	songTemple       = 20
	songInn          = 21
	songTrainingHall = 24 // with the fanfare; 22 is the loop alone
	songCityHall     = 23
	songTavern       = 25
	songGypsy        = 26
	songBattleBegins = 27 // a two-second sting
)

// amigaSong (extra) maps the Amiga songs Music starts to the soundtrack's:
// title the title, battle the combat theme, the party menu the training
// hall, Phlan New Phlan, party killed the game over, the overland map the
// wilderness.
var amigaSong = [...]int{0: 0, 1: songTitle, 2: songCombat, 3: songTrainingHall, 4: songNewPhlan, 5: songGameOver, 6: songWilderness}

// AreaMusic is a game's soundtrack by area (GameTables.Music; nil: none):
// each GEO block's song, and the songs of the battles fought there.
type AreaMusic struct {
	Songs       map[int]int // GEO block -> its song
	BattleSongs map[int]int // GEO block -> its battles' song; none: A Battle Begins into the combat theme
}

// PorMusic (extra) is POR's: the cartridge's own area table (bank 59 $8C46)
// where its areas match the PC's, the themes' names elsewhere; Tyranthraxus's
// lair (GEO 7) has his own battle song.
var PorMusic = AreaMusic{
	Songs: map[int]int{
		0: songNewPhlan, 1: songDungeon3, 2: songDungeon1,
		3: songValjevo, 4: songValjevo, 5: songValjevo, 6: songValjevo, 7: songLair,
		9: songValjevo, 10: songGraveyard, 13: songDungeon1, 14: songDungeon2, 15: songDungeon1,
		16: songYarash, 17: songGypsy, 18: songDungeon1, 19: songDiogenes, 20: songDungeon1,
		21: songDungeon2, 22: songYarash, 23: songYarash, 24: songDungeon2,
		25: songWilderness, 26: songWilderness, 27: songWilderness, 28: songDungeon3,
		29: songDungeon1, 30: songYarash, 31: songDungeon2, 32: songDungeon1,
	},
	BattleSongs: map[int]int{7: songTyranthraxus},
}

// music is the game's AreaMusic.
func (e *Engine) music() AreaMusic {
	if m := e.game().Music; m != nil {
		return *m
	}
	return AreaMusic{}
}

// areaSong is the song for where the party is: the wilderness on the
// overland map, a labelled building's (MapLabel.Song), else the area's (0: none).
func (e *Engine) areaSong() int {
	if e.viewMode > 1 {
		return songWilderness
	}
	k := e.areaKey()
	for _, l := range e.game().MapLabels {
		if l.Area == k && l.X == e.posX && l.Y == e.posY && l.Song != 0 {
			return l.Song
		}
	}
	return e.music().Songs[e.worldWord(rules.WordGeo)]
}

// battleMusic (extra) starts a battle's music: A Battle Begins into the
// combat theme, Tyranthraxus's own in his lair; the Amiga's battle song
// without the soundtrack.
func (e *Engine) battleMusic() {
	switch {
	case e.music().BattleSongs[e.worldWord(rules.WordGeo)] != 0 && e.songMusic(e.music().BattleSongs[e.worldWord(rules.WordGeo)], 0):
	case e.songMusic(songBattleBegins, songCombat):
	default:
		e.Music(2) // Amiga 0x364E2
	}
}

// SoundParts loads what the driver plays besides the synth, for
// audio.Driver.Attach: the Amiga players and their effects, the sound pack,
// the voice-over's speaking and the volume sliders.
func (e *Engine) SoundParts() (*orig.Player, []int, interface{ Mix([]byte) }, func() bool, func(string) float64) {
	var am *orig.Player
	if img, err := os.ReadFile(filepath.Join(e.adv, SoundDir, "SOUND.BIN")); err == nil && len(img) == orig.End-orig.Base {
		am = orig.NewPlayer(img)
		am.SongOn = 0x667d5
	}
	e.snd.dasp = loadDasp(filepath.Join(e.adv, SoundDir, daspDir))
	var pack interface{ Mix([]byte) }
	if e.snd.dasp != nil {
		pack = e.snd.dasp
	}
	return am, amigaFx[:], pack, e.vox.speaking, volume
}

// SetSound hands the rules the attached driver (audio.Driver, made on
// Cfg.Sound): main 0x4574 picks the card, SOUND.OFF keeps Ctrl-S's.
func (e *Engine) SetSound(d audio.Sound) {
	s := &e.snd
	s.d = d
	switch e.Cfg.Sound { // main 0x4574
	case 'T':
		s.typ = sndTandy
	case 'P':
		s.typ = sndSpeaker
	default:
		s.typ = sndNone
	}
	s.wait = 100 // soundWait(100) at start-up
	d.Wait(100)
	if e.marked(soundOffFile) { // Ctrl-S from the last run
		s.saved, s.typ = s.typ, sndNone
	}
	d.SetCard(s.typ)
	if s.stOn {
		e.SetSoundtrack(true)
	}
}

// PlaySound = 08AE:00DC (n): sndMute stops everything and mutes, sndUnmute unmutes,
// sndStop stops, anything else starts song n-1. While DS:5374 is non-zero the sound
// is played for 0.1 s and cut; after soundWait(0) it plays on until
// soundWait(n) unhooks the interrupt.
func (e *Engine) PlaySound(n int) { e.playSound(n, string(daspFx[n])) }

// cue (extra) is PlaySound(n) out of a fight as an event (Events):
// the presenter plays it when it draws the View that follows
// (PlayEventSound), as it does a fight's.
func (e *Engine) cue(c rules.SoundCue) {
	e.event(play.Sounded{Sound: string(c)})
}

// playSound is PlaySound with the sound pack's event (extra) in place of
// sound n: the pack plays it, and the pacing stays the original's.
func (e *Engine) playSound(n int, event string) {
	if e.inFight() && !e.soundNow && n > sndUnmute && n != sndStop { // in a fight: when the presenter plays the events up to here
		e.emit(rules.Sounded{Cue: rules.SoundCue(event)})
		return
	}
	s := &e.snd
	if s.d == nil || s.typ == sndNone || n == sndStep && e.NoSteps.Load() {
		return
	}
	quiet := n > sndUnmute && n != sndStop && e.dasp(event) || // extra: the sound pack's instead
		n == sndStep && s.d.Sounding() // extra: a step does not take the channels from a song
	s.d.Effect(n, quiet)
	if s.wait > 0 {
		s.d.CutAfter(sndRate / 10) // ten hundredths
		if !e.soundNow {           // the presenter plays a fight's events at its own pace
		}
	}
}

// soundWait = 08AE:01A5 (n): 0 hooks the timer interrupt so sounds play
// while the caller animates; n > 0 unhooks it (when it was hooked) and sets
// the cut-after-0.1 s mode used by PlaySound.
func (e *Engine) soundWait(n int) {
	s := &e.snd
	if s.d == nil || s.typ == sndNone {
		return
	}
	s.d.Wait(n)
	s.wait = n
}

// SoundReset = 08AE:017B: silence everything (used by the sound toggle and
// at start-up). It also copies DS:5372 to DS:5373.
func (e *Engine) SoundReset() {
	s := &e.snd
	s.saved = s.typ
	if s.d != nil {
		s.d.Reset()
	}
}

// soundOffFile in the save folder marks the sound as switched off (Ctrl-S).
const soundOffFile = "SOUND.OFF"

// soundToggle = 0899:006E (Ctrl-S in the key reader): sound on means
// remember DS:5372 in DS:5373, stop the song (PlaySound(sndStop))
// and set 5372 to sndNone, which mutes every driver call; off again restores the
// saved type (the set test [0,1] always holds) and resets the driver.
func (e *Engine) soundToggle() {
	s := &e.snd
	defer func() { e.mark(soundOffFile, s.typ == sndNone) }()
	if s.typ != sndNone {
		s.saved = s.typ
		e.PlaySound(sndStop)
		s.setCard(sndNone)
		return
	}
	if s.saved == sndTandy || s.saved == sndSpeaker {
		s.setCard(s.saved)
		e.SoundReset()
	}
}

func (s *sound) setCard(c int) {
	s.typ = c
	if s.d != nil {
		s.d.SetCard(c)
	}
}

// SoundData is the driver's tables (soundFile), for audio.NewDriver; Data's
// SoundTables after it.
func (e *Engine) SoundData() ([]byte, error) {
	if e.sounds == nil {
		b, err := os.ReadFile(filepath.Join(e.adv, DataDir, SoundFile))
		if err != nil {
			return nil, fmt.Errorf("the sound tables are not imported (run the import): %w", err)
		}
		e.sounds = b
	}
	return e.sounds, nil
}

// The sound numbers the game passes to PlaySound (the table at DS:2612).
const (
	sndStop       = 255 // stop the song
	sndMute       = 0   // stop and mute
	sndUnmute     = 1
	sndSpell      = 2 // a spell flies (the spells but 47 and 51)
	sndHit        = 3 // a weapon hits
	sndMagicHit   = 4 // a magical weapon hits
	sndFall       = 5 // a combatant falls
	sndMissile    = 6 // a missile flies
	sndBump       = 7 // walking into a wall
	sndSpell51    = 8
	sndMiss       = 9 // an attack hits nobody
	sndStep       = 10
	sndSpell47    = 11
	sndMissileEnd = 12 // a missile arrives
)

// daspFx is the sound pack's event that replaces each sound number; the
// rest (a step) keep the original's sound.
var daspFx = map[int]rules.SoundCue{
	sndSpell:      rules.CueSpell,
	sndHit:        rules.CueHit,
	sndMagicHit:   rules.CueMagicHit,
	sndFall:       rules.CueFall,
	sndMissile:    rules.CueMissile,
	sndBump:       rules.CueBump,
	sndSpell51:    rules.CueLightning,
	sndMiss:       rules.CueMiss,
	sndSpell47:    rules.CueFireball,
	sndMissileEnd: rules.CueMissileEnd,
}
