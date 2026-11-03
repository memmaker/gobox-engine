# gobox-engine: design

One shared source for the remastered Gold Box games: por-go (Pool of
Radiance), cotab-go (Curse of the Azure Bonds), sotsb-go (Secret of the
Silver Blades). Each game keeps its engine (`engine.Apply(decision)`,
synchronous, no goroutines, no Ebiten) and its data reading (DAX, ECL, GEO,
WALLDEF). gobox-engine holds only what presents a game: it never imports a
game, and a game feeds it plain Go values built from its own data.

## Modules

Each directory is its own Go module. A game uses them with `replace`:

```
require gobox/gpu v0.0.0
replace gobox/gpu => ../gobox-engine/gpu
```

Data and sound (each module's API.md has its exported API and the switch
steps per game):

```
adf       Amiga disk images                       identical in all three
ega       EGA palette, W/H, Amiga palettes        palette only; sotsb's Screen stays
dax       DAX files, images, sprites               Block copies; BlockShared = por's cached slice
exe       EXEPACK START.EXE                        Load(path, ds): por 0xC7C, cotab 2750, sotsb 3554
geo       GEO/WALLDEF parsing                      bytes in; LoadGeo/LoadWalldef, engine state stay in the game
amiga     Amiga SOUND.BIN player                   por's player; SongOn address a field
audio     PC speaker/Tandy synth, oto Output       NewPlayer(Config), Play(n); data JSON from audio/cmd/sounddata
fixture   save fixtures for tests                  the game supplies callbacks (Open, ParseDecision ...)
scripts   build_n_run.sh GAMEDIR                   one line wrapper per game
tools/remaster  texture pipeline                   per-game remaster.json + areas_data.py
ui        widgets: scroll bars, menu rows, key tables, pager core, map images, results
```

Presentation:

```
gpu       window (glfw), input, 2D batches, WGSL shaders, DrawMesh (wgpu)
input     Act: what a key does (bindings live in the game's front end)
blobber   first-person 3D view          needs gpu
combat    3D combat tabletop            needs gpu, input
window    window manager (phase 2)      needs gpu, input
```

| Module | Takes | Returns |
|---|---|---|
| `gpu` | images, triangles, meshes, shaders | frames on screen, key/pad/mouse state |
| `input` | — | the `Act` names screens switch on |
| `blobber` | wall textures, map layouts (`maps`), a `Camera` | the drawn view; the camera tween (movement) |
| `combat` | a `Board` snapshot: ground, tiles, units with icons, effects, the turn | the drawn tabletop; `Decision`s (walk, attack, aim, bar button) |
| `window` | text, images, menu definitions, sprites, tiles, fonts | window overlays drawn over a view; `Decision`s |

Rules for every module:
- No game imports, no env vars, no game loop of its own (the game's `cmd` owns it).
- Snapshot in, decisions out. A module never changes game state; the game maps
  a returned decision onto its own engine calls (por-go: rule calls such as
  `TurnChoice`, `CombatStep`, queued by the presenter; sotsb-go: `Apply`).
- Per-game differences are data in the snapshot or a `Config`, never a
  build tag or a game switch inside the module.

## Per-game differences (as data)

| Difference | Where it goes |
|---|---|
| wall-set sizes, symbol blocks (WALLDEF) | the game builds `blobber` wall textures and `maps` from its own WALLDEF/GEO |
| palettes | the game converts to RGBA before handing images over (tiles, icons, walls) |
| combat icons, tile sets | `combat.Board.Tiles`, `Unit.Icon` (already RGBA) |
| DungCom's stone tile count (`DungeonTiles`, PoR 0x19) | `combat.Board.WallTiles` |
| SotSB fights only in dungeons | the game always sets `Board.Dungeon`; nothing in `combat` assumes outdoors |
| which bar buttons exist | `Turn.Can` lists the `combat.Kind`s offered, in menu order |

## combat: the seam

`combat` owns its snapshot types (a subset of por-go's `engine.Battle`):
`Board`, `Square`, `Unit`, `Turn`, `Step`, `Effect`, `Strike`, and the
decision type `Decision{Kind, N, X, Y}` with its own `Kind` enum (Walk,
Attack, Aim, Point, Close, Yes, No, Wait, End, Quick, Guard, Bandage, Flee,
Turn, Cast, Items, Ranged, Maneuver). Each game writes one small adapter
(its battle snapshot → `combat.Board`, `combat.Decision` → its engine calls).

Confirms: a choice the rules refuse until the player is sure (Flee?, Attack
Ally?) comes back to the game as a refusal; the game sets `Board.Confirm`
(the question) and `combat` returns `Yes` or `No`, then the game calls again
with `sure`. A maneuver may pre-aim a square before the attack aim
(`Maneuver.Move`, Deft Strike, Evasive Strike): `combat` aims twice and
returns both squares in `Decision.Pre` and `X, Y`.
Game-only extras (the weapon item, spells, maneuvers) stay in the game; the
front end passes icons and labels for them through `Turn.Icons`/`Labels`.

Units cover `Squares` (1x1 or 2x2) with an `IconSize` of their own (24 or
48 px); icons come in already mirrored for facing. `Board` carries the
aiming cursor (`Cursor`, `CursorOn`, `CursorAt`) and `View`, the original's
scrolled window, as the camera's framing hint. `Turn.Can` is data, taken in
order from the game's own menu (PoR's bar, CotAB's MOVE VIEW AIM TURN QUICK
DONE QMenu); `Turn.Labels` and `Turn.Icons` are optional per kind. Kinds
include `View` and `Step` (one keypad step, N = direction 1..9, for games that
walk a square at a time, CotAB's QMove) beside `Walk` (a path to a Reach
square).

## blobber

The host fills the textures: `Textures map[string]image.Image` by name (the
remastered PNGs, with the game's WALLDEF front view as the fallback); the
embedded PoR walls and `walls.txt` stay with por-go. Faces are
`Faces[{x,y,side}][]FacePart{Tex, T1, U1}` with Edges, NoMap, Water, Pond.
CotAB's sky (`celestial.go`: sun by the hour, night, northern mountains) and
the per-square floor/sky flats (`SetFlats`) move from package globals onto
the `View`/`Camera` as fields. Maps are the game's (CotAB: 16x16, 3 wall
slots of up to 15 walls).

## Overlays

`blobber` and `combat` draw into the caller's `*gpu.Image` and nothing after
it, so a game draws its 2D layer on top (CotAB's 80x50 console in 8X8D, the
window manager) and can `ReadPixels` the result for shots.

## window (phase 2)

por-go's `present/remaster` is 33 files written straight against PoR's
`engine.View`. It moves in pieces once the shared parts are clear: first the
generic widgets (pager, scrollbar, list, question, picker, text and fonts,
portrait cards), then screens that the three games share in shape (sheet,
inventory, shop, camp). Screen code that reads a game's `View` stays in the
game. Until then cotab-go and sotsb-go keep their copies of the screens.

## Status

- [x] gpu, input, blobber extracted; por-go builds against them
- [x] combat extracted with its own types; por-go adapter
- [x] interfaces agreed with the POR, COTAB and SOTSB sessions
- [x] por-go, cotab-go switched; sotsb-go on the data modules (audio, blobber, combat at D11–D14)
- [ ] window (phase 2)
- [ ] combat bar icons for Move, View and Aim (`combat/icons/{move,view,aim}.png`, 64px like the others); letters until then. They're art: ask before spending on an image model
- [ ] dax: remove `Loader.OneDisk`. Make the other-disk fallback opt-in (por sets it) and drop `OneDisk = true` from cotab and sotsb

combat status: extracted into `combat/` (module gobox/combat), kinds are named
`ActWalk` etc. (`Step` and `Turn` are also types). TODO: `Board.Confirm`,
`Decision.Pre` and `ActStep`/`ActView` are fields/kinds only; no behaviour yet
(no Yes/No from Confirm, no double aim, no keypad step, Turn.Labels/Icons not
drawn, Turn.Moves unused).
