package front

// Sound is the sound driver the rules call synchronously (gobox/audio.Driver;
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
