# gobox/blobber API (first-person grid crawler, pure 2D quads; needs gpu)

Deepnight look: every surface a hand-transformed flat quad, painted far to near; the texture shader steps darkness into fog. Two ways in: the standalone `Game` (owns loop and rules, cmd/) and `View` (host owns both). Maps are the game's (`LoadMap` reads the .map format from any io.Reader).

## View (the host-driven renderer)
- `NewView(textures map[string]image.Image) (*View, error)`: after the window runs; nil = only generated materials; the game fills `Textures` and `Painted` (wall char -> texture name). BLOBBER_NIGHT=1 starts at night.
- Fields: `Textures` (by name: FacePart.Tex, "floor_NAME", "ceiling_NAME"; call `ResetTextures` after changing), `Night` (stars, moonlight), `SunUp` + `SunAz` (compass rad, 0 N, pi/2 E) + `SunEl`, `Mountains` (range on the north horizon).
- `Show(*Scene, Camera)`, `Follow(Camera)` (neighbour step/turn animates, further = jump), `Update()` once per tick, `Draw(dst)`, `DrawMenu(dst)` over it.
- `SetFlats(floor, sky)` flats of the camera square; package `SetFlats(floor, ceiling)` the same ("" = default), call on each map change; it survives texture regeneration. `FlatTexture(kind,name)`, `ResetTextures()`.
- Free walk: `Walk(at, open [4]bool, FreeInput{Forward,Right,Turn,Pitch}, ready) Camera`, `FreeOff()`. Fluid keys: `Flow(held, ready)`, `Fluid()`.
- Esc menu: `OpenOptions/OptionsOpen/UpdateOptions`, `PadKey(k)`, `PointOptions(line,click,back,wheel)`.

## Maps and scene
- `LoadMap(r) (rows, extra, err)`: '.' floor, other chars walls by material (`MatColors`); `Door 'D'`, `Locked 'L'`; non-grid lines come back in extra.
- `ParseScene(r) (*Scene, extra, err)`; `Scene{Rows, Faces, Edges, NoMap, Water, Pond, Sea, Floor, Ceiling, Props}`; `Scene.Show(objs...)`. Directives: face, edge, nomap, water, pond, sea edge, prop, well, floor, ceiling.
- `Faces map[[3]int][]FacePart{Mat,T0,T1,U0,U1,Tex}`, `Edges` (thin walls, key {x,y,side}), `Water`, `Pond`, `Sea`, `NoMap`: package globals that Show swaps.
- `Camera{X,Y,Dir}` (0 N,1 E,2 S,3 W). `Object{X,Y,Img,Size,At,Front,Sink,Ring}` billboard; `Objects` global list. `SetMap(lvl,x,y)`, `Face(dir)`.

## Host hooks (package vars, set before NewView)
`Settings Options` (+ `OnSettings`), `Glyph` (8x8 font, drawn 2x), `FontNames/FontAt/SetFont/FontGlyph/MenuGlyph`, `Window` (Esc window look), `Held`, `Repeats`, `OnSaveLoad`, `OnMainMenu`. Text: `Print`, `Width`, `LineH=16`, `Hotkeys`, `KeyJustPressed/KeyRepeated`.

## How a game feeds it
Host builds a Scene per map (ParseScene plus its own state lines), `Show`s it on entry, moves the Camera itself (or via Walk), calls Update per tick and Draw into its image. Back: Camera from Walk, menu state; Options changes via OnSettings.

## Gotchas
- Global state: one View per process; scene, camera, Faces/Edges/Water/Objects, W/H, Settings are package vars (ponytail note in View's doc). Layout sets W,H.
- Only the Esc menu reads the keyboard through View; everything else is host-fed.
- Options holds only what blobber reads (Smooth, Fluid, Dither, Fullscreen, Torch, Speed, NoGrid); hosts keep their own settings.

## Migrate
Embedded game assets were removed from blobber. Last commit that has them: `b8cbba982153eafcde2ccd3266300f7e481f732a`.
Each game now supplies:
- Textures: `View.Textures` (name -> image) and `View.Painted` (wall char -> name). Names are plain now: `floor_water`, `ceiling_clouds` (were `*_phlan_*`), `stone`, `floor`, `ceiling`, `wall_*`, `prop_*`.
- Maps: the .map files, read with `LoadMap(io.Reader)`; `maps.Levels` is gone.
- Menu: the Esc menu lost the host rows (portraits, minimap, artwork, audio page and volumes, deskew, 4e, zoom, explicit) and `Voice`, `UISteps`; the game keeps those settings in its own struct and puts the rows back with `blobber.ExtraRows` (`map[page][]Row{Label, At, Get func() string, Change func(d int)}`: `Gameplay` and `Graphics` rows go before built-in row `At`; any other key is a new page on the root menu, after Graphics, sorted). `Change` gets 0 Enter, -1 Left, 1 Right; `OnSettings` runs after it.
- Ponds: `View.PondHalf/PondNeck/PondMud` default to 0.46/0.36/0.12 (por's).
- por copies its assets (run in por-go):
  `git -C ../gobox-engine archive b8cbba982153eafcde2ccd3266300f7e481f732a blobber/assets blobber/maps blobber/tools blobber/create_textures.sh | tar -x --strip-components=1 -C <dest>`
  or one file: `git -C ../gobox-engine show b8cbba982153eafcde2ccd3266300f7e481f732a:blobber/assets/walls/stone.png > stone.png`.
  Rename `floor_phlan_water.png` -> `floor_water.png`, `ceiling_phlan_clouds.png` -> `ceiling_clouds.png`. The ignored `art/` dir was never in git.
