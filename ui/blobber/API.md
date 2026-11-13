# gobox/ui/blobber API (the first-person 3D view; needs gpu only, never imports ui)

Deepnight look: every surface a hand-transformed flat quad, painted far to near; the texture shader steps darkness into fog. The one way in is `View`: package ui (the host) owns the loop, the rules, the input and the Esc menu. Maps are the game's (`LoadMap` reads the .map format from any io.Reader).

## View (the host-driven renderer)
- `NewView(textures map[string]image.Image, o *Options) (*View, error)`: after the window runs; nil = only generated materials; `o` is the host's one copy of the view's options (ui hands `&ui.Settings.Options`), read live; the game fills `Textures` and `Painted` (wall char -> texture name). BLOBBER_NIGHT=1 starts at night.
- Fields: `Textures` (by name: FacePart.Tex, "floor_NAME", "ceiling_NAME"; call `ResetTextures` after changing), `Night` (stars, moonlight), `SunUp` + `SunAz` (compass rad, 0 N, pi/2 E) + `SunEl`, `Mountains` (range on the north horizon).
- `Show(*Scene, Camera)`, `Follow(Camera)` (neighbour step/turn animates, further = jump), `Update()` once per tick, `Draw(dst)`.
- `SetFlats(floor, sky)` flats of the camera square; package `SetFlats(floor, ceiling)` the same ("" = default), call on each map change; it survives texture regeneration. `FlatTexture(kind,name)`, `ResetTextures()`.
- Free walk: `Walk(at, open [4]bool, FreeInput{Forward,Right,Turn,Pitch}, ready) Camera`, `FreeOff()`. Fluid keys: `Flow(held, ready)`, `Fluid()`.
- `Options{Smooth, Fluid, Dither, Torch, Speed, NoGrid}`, `MinSpeed/MaxSpeed`. `Tint(img, c)`.

## Maps and scene
- `LoadMap(r) (rows, extra, err)`: '.' floor, other chars walls by material (`MatColors`); `Door 'D'`, `Locked 'L'`; non-grid lines come back in extra.
- `ParseScene(r) (*Scene, extra, err)`; `Scene{Rows, Faces, Edges, NoMap, Water, Pond, Sea, Floor, Ceiling, Props}`; `Scene.Show(objs...)`. Directives: face, edge, nomap, water, pond, sea edge, prop, well, floor, ceiling.
- `Faces map[[3]int][]FacePart{Mat,T0,T1,U0,U1,Tex}`, `Edges` (thin walls, key {x,y,side}), `Water`, `Pond`, `Sea`, `NoMap`: package globals that Show swaps.
- `Camera{X,Y,Dir}` (0 N,1 E,2 S,3 W). `Object{X,Y,Img,Size,At,Front,Sink,Ring}` billboard; `Objects` global list. `SetMap(lvl,x,y)`, `Face(dir)`.

## Not here
The Esc menu (rows, pages, fonts, keys, Settings, OnSettings, ExtraRows, Hotkeys, QuitRow) is package ui's (ui/options.go); it reads the keys (ui/pad.go) and draws in the game's text window.

## How a game feeds it
Host builds a Scene per map (ParseScene plus its own state lines), `Show`s it on entry, moves the Camera itself (or via Walk), calls Update per tick and Draw into its image. Back: Camera from Walk.

## Gotchas
- Global state: one View per process; scene, camera, Faces/Edges/Water/Objects, W/H and the Options pointer are package vars (ponytail note in View's doc). Layout sets W,H.
- Nothing here reads the keyboard; everything is host-fed (FreeInput is the decoded walk).
- Options holds only what the view reads; Fullscreen is ui.MenuOptions'.

## Migrate
Embedded game assets were removed from blobber. Last commit that has them: `b8cbba982153eafcde2ccd3266300f7e481f732a`.
Each game now supplies:
- Textures: `View.Textures` (name -> image) and `View.Painted` (wall char -> name). Names are plain now: `floor_water`, `ceiling_clouds` (were `*_phlan_*`), `stone`, `floor`, `ceiling`, `wall_*`, `prop_*`.
- Maps: the .map files, read with `LoadMap(io.Reader)`; `maps.Levels` is gone.
- Menu: the Esc menu lost the host rows (portraits, minimap, artwork, audio page and volumes, deskew, 4e, zoom, explicit) and `Voice`, `UISteps`; the game keeps those settings in its own struct and puts the rows back with `ui.ExtraRows` (`map[page][]Row{Label, At, Get func() string, Change func(d int)}`: `Gameplay` and `Graphics` rows go before built-in row `At`; any other key is a new page on the root menu, after Graphics, sorted). `Change` gets 0 Enter, -1 Left, 1 Right; `OnSettings` runs after it.
- Ponds: `View.PondHalf/PondNeck/PondMud` default to 0.46/0.36/0.12 (por's).
- por copies its assets (run in por-go):
  `git -C ../gobox-engine archive b8cbba982153eafcde2ccd3266300f7e481f732a blobber/assets blobber/maps blobber/tools blobber/create_textures.sh | tar -x --strip-components=1 -C <dest>`
  or one file: `git -C ../gobox-engine show b8cbba982153eafcde2ccd3266300f7e481f732a:blobber/assets/walls/stone.png > stone.png`.
  Rename `floor_phlan_water.png` -> `floor_water.png`, `ceiling_phlan_clouds.png` -> `ceiling_clouds.png`. The ignored `art/` dir was never in git.
