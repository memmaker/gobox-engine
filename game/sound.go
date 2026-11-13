package game

// What the engine plays through: the host (cmd/gobox) makes them in
// gobox/audio and hands them to the engine, so the engine never imports
// audio (REFACTOR.md E2).

// Sound is the sound driver the rules call synchronously (audio.Driver;
// a fake in tests). Cards: 0 Tandy, 1 speaker, 2 none.
type Sound interface {
	Effect(n int, quiet bool) // PlaySound n; quiet: hook only, start nothing
	CutAfter(frames int)      // stop and unhook after frames output frames
	Wait(n int)               // 0 hooks, n > 0 the cut mode
	Reset()
	SetCard(c int)
	Song(n, next int, again bool) bool // soundtrack song n then next; false: no soundtrack
	AmigaSong(n int)
	SetSoundtrack(on bool, n, music int)
	Track() (on bool, song int) // the soundtrack plays, its song
	Sounding() bool             // the Amiga players sound
}

// SoundPack plays the sound pack's events (audio.SoundPack).
type SoundPack interface {
	Play(event string) bool // false: the pack has nothing for event
}

// Voice plays the voice-over's recordings (audio.Voice).
type Voice interface {
	Say(ogg []byte) error // queue an Ogg Vorbis recording
	Stop()                // silence and drop the queue
}
