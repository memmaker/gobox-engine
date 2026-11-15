# gobox/asset/orig (amiga): Amiga Pool of Radiance / CotAB / SotSB sound (Paula + Wally Beben player)

Pure Go, no deps. Plays the game's own 68000 sound code over its memory image.

```
amiga.Base, amiga.End          image bounds (0x63000..0x91000)
amiga.Image(exe) []byte        flatten the "program" hunk file; nil if not this game (importers)
amiga.Title(intro) image.Image PoR painted title from disk 1 "intro"; nil if not that file
p := amiga.New(img)            img = Base..End bytes (len End-Base)
p.SongOn uint32                address of the "song running" byte (no default; the host must set it)
p.Song(n) / p.Effect(n)        start song n (0 stops) / sound effect n
p.Render(out, rate)            fill out with 16-bit LE stereo PCM, advancing the players
p.Playing() / p.Sounding()     music has the channels / a song (not a one-shot jingle) is going
```

Test: `AMIGA_SOUND=path/to/SOUND.BIN go test` checks testdata/golden.txt (skipped otherwise).

## Switch steps (per game)
1. go.mod: `require gobox v0.0.0`, `replace gobox => ../gobox-engine` and `replace gobox/rules => ../gobox-engine/rules` (one module: every package below imports as `gobox/<dir>`).
2. Change imports `"<game>/amiga"` to `"gobox/asset/orig"`; same names (NewPlayer, AmigaImage, ...).
3. por: nothing else (uses Sounding). cotab/sotsb: calls unchanged (Playing, Effect, Render); they never call Title.
4. Delete `<game>/amiga` and its imports; `go vet ./... && go test ./...`.
5. All games: set `p.SongOn` after New (no default any more; 0x667d5 in por's SOUND.BIN) -- Sounding and the music player read it.

## Migrate
- `New` no longer sets `SongOn`: set `p.SongOn` after it (was 0x667d5). The player now uses the field everywhere (it was a literal in the song start/stop code).
- Other 0x667xx addresses in the player are the same RAM block of the one SOUND.BIN image; a game whose image puts the block elsewhere needs those parametrised too (only SongOn is, so far).

## Music-X (SotSB, PoD)
```
m, err := amiga.NewMX(mx, glb)  SECRET.MX / pod.mx (MXTX) and sounds.glb (GLIB of DIG8; may be nil)
m.Songs map[int]int             Song n -> file song (0-based); nil: n-1
m.Song(n) / m.Effect(i)         song once (0 stops) / GLIB entry i on two channels at volume 45
m.Render, Playing, Sounding     as Player; m.Len() songs in the file
```
Front: game.Data.AmigaMusic{Songs, Effects, SongOf, EffectOf}; amigaimport Spec.Files copies the files.

Traced from SotSB's `Secret` (code hunk at 0x10000; player 0x5a20c..0x5c3a8). PoD's `Pools of Darkness` has the same code at +0xf1e (0x5b12a..0x5d2c6), differing only in relocated data addresses, so one player serves both.
- Voices (0x5a252, 0x5a33a): voice n is Paula channel n (0, 3 left; 1, 2 right). Even MIDI channels prefer left, odd right (0x5a71c; controller 10 changes it). The note goes to the other side if both voices of its own are held and the other has one free or releasing; of a side's two voices the one in the lower state (free < stopping < releasing < sustain < attack < new), taking turns when equal. A song note always gets a voice: priorities and locked voices only serve the player's own sample calls, which the games do not use (0x571ee sends its own audio.device requests).
- Volume (0x5a20c): 64 scaled by velocity+1 (only with header flag bit 1, 0x5c156; else velocity just gates), envelope and channel volume (controller 7), each /128.
- Pitch (0x5a376): note + transpose + bend to a log2 period; a new note moves to the next octave chunk (twice the length, half the period) while the period is over 0x8fd77 and keeps that chunk.
- Events (0x5b4c0): note, 0x80 tempo (ramp over arg ticks; arg >= 0x8000 is at once, the compare being signed; the ramp takes the change unsigned), 0xa0 meta (2 n / 3: repeat, 4 deep; 0 section mark, 1 signals the game: not acted on), 0xb0 controllers 6 (bend range, RPN 0), 7, 10, 0x40, 0x64, 0x65, 0x79, 0x7b, 0xc0 instrument, 0xe0 pitch bend (24 semitones at rest), 0xff end.
- Not ported (in neither file): portamento and mono legato (controllers 5, 0x25, 0x41, 0x7e, 0x7f; 0x5af90..0x5b00c, 0x5a38e), the LED filter (header flag bit 0 and controller 0x51 -> 0x5b39c, CIA 0xbfe001 bit 1; both files clear it), the loader's limit of 8 songs (0x5c1bc), looped play (0x5b91c) and start at a section (0x5b7fc).
- In the files: both have header flags 0. SECRET.MX: notes and one tempo per song only, one octave chunk each, 4 notes at most. pod.mx: controller 7 on every used channel, tempo changes of >= 0x8000 ticks in song 1, instrument 3 with two octave chunks, up to 7 notes at once.
