# gobox/amiga: Amiga Pool of Radiance / CotAB / SotSB sound (Paula + Wally Beben player)

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
1. go.mod: `require gobox/amiga v0.0.0` and `replace gobox/amiga => ../gobox-engine/amiga`.
2. Change imports `"<game>/amiga"` to `"gobox/amiga"`; same names (New, Base, End, Image, Title, Render, Effect, Song, Playing).
3. por: nothing else (uses Sounding). cotab/sotsb: calls unchanged (Playing, Effect, Render); they never call Title.
4. Delete `<game>/amiga` and its imports; `go vet ./... && go test ./...`.
5. All games: set `p.SongOn` after New (no default any more; 0x667d5 in por's SOUND.BIN) -- Sounding and the music player read it.

## Migrate
- `New` no longer sets `SongOn`: set `p.SongOn` after it (was 0x667d5). The player now uses the field everywhere (it was a literal in the song start/stop code).
- Other 0x667xx addresses in the player are the same RAM block of the one SOUND.BIN image; a game whose image puts the block elsewhere needs those parametrised too (only SongOn is, so far).
