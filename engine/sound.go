package engine

import (
	"fmt"
	"os"
	"path/filepath"

	"gobox/adventure"
	"gobox/asset/orig"
	"gobox/audio"
	"gobox/game"
	"gobox/rules"
)

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
	if a := e.music().AmigaSongs; n >= 0 && n < len(a) && e.songMusic(a[n], 0) {
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
	return e.snd.d != nil && e.snd.d.Song(n, next, n == adventure.SongBattleBegins)
}

// titleMusic reports whether the title song plays, which plays on into the
// main menu.
func (e *Engine) titleMusic() bool {
	if e.snd.d == nil {
		return false
	}
	on, song := e.snd.d.Track()
	return on && song == adventure.SongTitle
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

// music is the game's AreaMusic.
func (e *Engine) music() adventure.AreaMusic {
	if m := e.game().Music; m != nil {
		return *m
	}
	return adventure.AreaMusic{}
}

// areaSong is the song for where the party is: the wilderness on the
// overland map, a labelled building's (MapLabel.Song), else the area's (0: none).
func (e *Engine) areaSong() int {
	if e.viewMode > 1 {
		return adventure.SongWilderness
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
	case e.songMusic(adventure.SongBattleBegins, adventure.SongCombat):
	default:
		e.Music(2) // Amiga 0x364E2
	}
}

// SoundParts loads what the driver plays besides the synth, for
// audio.Driver.Attach: the Amiga players and their effects, the sound pack,
// the voice-over's speaking and the volume sliders.
func (e *Engine) SoundParts() (*orig.Player, []int, interface{ Mix([]byte) }, func() bool, func(string) float64) {
	var am *orig.Player
	if img, err := os.ReadFile(filepath.Join(e.adv, adventure.SoundDir, "SOUND.BIN")); err == nil && len(img) == orig.End-orig.Base {
		am = orig.NewPlayer(img)
		am.SongOn = 0x667d5
	}
	e.snd.dasp = loadDasp(filepath.Join(e.adv, adventure.SoundDir, daspDir))
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
	e.event(game.Sounded{Sound: string(c)})
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
		b, err := os.ReadFile(filepath.Join(e.adv, adventure.DataDir, adventure.SoundFile))
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
