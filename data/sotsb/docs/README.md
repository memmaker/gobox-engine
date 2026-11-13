# sotsb-go

*Secret of the Silver Blades* (SSI, 1990) in Go. This repo holds only the
game's import (`importsotsb/`: the originals into an adventure folder, with the
game's data). The game runs on gobox's shared engine (`gobox/pool` over
`rules1e`, through `gobox/pool/playpool`). Drawing, input, menus, sound playback, saves and the app
shell come from gobox (`../gobox-engine`: `app`, `front`, `front/present`,
`gpu`, `blobber`, `combat`, ...), so the presentation is modern and native,
while the content and rules come from the original game. The source of
truth for the rules is the Turbo Pascal reconstruction in `../sotsb-re`.

No game data is included. The game reads the original files (`*.DAX`,
`START.EXE`, `GAME.OVR`, `ITEMS`, `BLADES.CFG`) from your own copy.

## Import the game

Run `gameimport` (`../gameimport/cmd/gameimport`) in this folder. It reads
`import.toml` and asks for the game: a folder, a zip, a CD image or the GOG
installer. Then it copies the files into `data/`.

## Build and run

```
go build ./cmd/blades
./blades -data data
```

- `-saves DIR`: the save folder (default `DIR/SAVE` if it exists, else `DIR`)
- `-load NAME` or `-slot NAME`: start from a named save slot
- `-scale N`: the window scale (default 2)
- gobox `app.Flags`, for captures and measuring: `-keys` (keys to play; escapes
  `\r \e \U \D \L \R \F`), `-shot FILE` (writes a PNG after the keys and exits),
  `-pointer`, `-size`, `-freeze`, `-quit`, `-fps`, `-offscreen`

`go test ./rules` drives the engine headless through `Apply` and needs game
data with save A in `$SOTSB_DATA` (default `../sotsb-re/gog`).

## Layout

- `importsotsb/`: `gobox import sotsb`: scripts with the game's fixes, maps,
  records, art, the game's data (gamedata.json, gametables.json, rules.json).
- `playsotsb/`: starts the game on `gobox/pool/playpool`.
- `cmd/blades`: the native binary.
- `cmd/dsbytes`, `cmd/dsfind`, `cmd/itemdump`: porting aids (tablefind: gobox exe/cmd/tablefind)
  over the data segment and the item files.
- `cmd/fixture`: builds save fixtures in another area.
- `import.toml`: what `gameimport` copies.

## What's left

See `MODERNIZE.md`, which has the open D items: the adventure Walk, the sheet,
shops, temples and training as rule calls, 3D exploration, sound in a window,
and por-go's UI extras.
