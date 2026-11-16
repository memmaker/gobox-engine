package importer

import (
	"strconv"

	"gobox/adventure"
	"gobox/rules"
)

// brSoundLimit is the first Sound number the script's command plays nothing for: BR1's command 67 (rules/eclops.go,
// "R7DA_0000; 128 and up none") reports the numbers 0..127 to the front, which plays its own files by number.
const brSoundLimit = 128

// brScriptSounds is BR's Sound command: number n is the cue named by the number itself, as the old front's EvSound's
// Sound was (strconv.Itoa(n)); the original's sound driver (SOUNDDRV, START.EXE's tables) is not decoded.
func brScriptSounds() map[byte]adventure.ScriptSound {
	m := make(map[byte]adventure.ScriptSound, brSoundLimit)
	for n := range brSoundLimit {
		m[byte(n)] = adventure.ScriptSound{Cue: rules.SoundCue(strconv.Itoa(n))}
	}
	return m
}
