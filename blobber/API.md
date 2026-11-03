# gobox/blobber API (first-person grid crawler, pure 2D quads; needs gpu)

Deepnight look: every surface a hand-transformed flat quad, painted far to near; the texture shader steps darkness into fog. Two ways in: the standalone `Game` (owns loop and rules, cmd/) and `View` (host owns both). Maps live in `maps/` (*.map; `maps.Levels` embed.FS).

## View (the host-driven renderer)
- `NewView(textures map[string]image.Image) (*View, error)`: after the window runs; nil = embedded PoR walls. BLOBBER_NIGHT=1 starts at night.
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
`Settings Options` (+ `OnSettings`), `Glyph` (8x8 font, drawn 2x), `FontNames/FontAt/SetFont/FontGlyph/MenuGlyph`, `Window` (Esc window look), `Held`, `Repeats`, `Voice`, `OnSaveLoad`, `OnMainMenu`, `UISteps`. Text: `Print`, `Width`, `LineH=16`, `Hotkeys`, `KeyJustPressed/KeyRepeated`.

## How a game feeds it
Host builds a Scene per map (ParseScene plus its own state lines), `Show`s it on entry, moves the Camera itself (or via Walk), calls Update per tick and Draw into its image. Back: Camera from Walk, menu state; Options changes via OnSettings.

## Gotchas
- Global state: one View per process; scene, camera, Faces/Edges/Water/Objects, W/H, Settings are package vars (ponytail note in View's doc). Layout sets W,H.
- Only the Esc menu reads the keyboard through View; everything else is host-fed.
- Not every Options field is blobber's: most are host settings it stores and persists.
