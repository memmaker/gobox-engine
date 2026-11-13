# gobox/ui API: combat (3D tabletop combat map; needs gpu, input)

Reads a `Board` snapshot, draws figurines on a wooden table, returns `Decision`s. The camera turns in eighths (XCOM style).

## Snapshot types (game -> combat)
- `Board`: Ground [25][50]Square, Tiles []image.Image, WallTiles, StoneTile (the wall tile used as stone texture; was fixed 3), Floor, Dungeon, Units, Cursor/CursorOn/CursorAt, Marks, Reach, Paths map[Point][]Step, Targets (unit IDs), Turn *Turn (nil = nobody acts), Panel, Log, Area/AreaAt/Skipped, AimWhy, Effects, View, Strike, Confirm.
- `Square{Tile,Wall,Height}`, `Step{Point,Free}` (Free: leaving provokes), `Effect{X,Y,Pic}`, `Strike{Seq,From,To,Hit,Damage}` (new attack = new Seq).
- `Unit`: ID, X/Y/Size, Squares (1x1 or 2x2), Party (blue base, else red), Facing 0..7, Icon + IconSize (24 or 48), Active, Name, HP/MaxHP/AC/Hit/Dmg/Status.
- `Turn`: Unit, `Can []Kind` (buttons in the game's menu order), Aim, Casting, Moves, `Labels map[Kind]string` (button words). (`Turn.Icons` was removed: use `Presenter.Icons`.)

## Decisions (combat -> game)
- `Decision{Kind, N, X, Y, Pre}`. Kinds: ActNone, ActWalk, ActStep, ActAttack, ActAim, ActPoint, ActClose, ActYes, ActNo, ActWait, ActEnd, ActQuick, ActGuard, ActBandage, ActFlee, ActTurn, ActCast, ActItems, ActRanged, ActManeuver, ActView.

## Presenter (one per combat; zero value works)
- `Draw(dst, *Board)`; `DrawBar(dst, glyph, *Board, k, bottom)` the action bar at scale k; `HideBar()`.
- Input: `Key(act input.Act, k uint16) bool` (W A S D pan, arrows cursor, Q/E turn, Home reset, digits/letters buttons, Tab cycles targets, Enter, Esc = No, Backspace ends turn), `Point(x,y)`, `Click(x,y) bool`, `Zoom(in, speed)`, `Pick(x,y)`, `Hovered()`.
- Out: `Actions() []Decision` (drain once a frame). `Float(id,n,heal)` starts a damage number.
- Motion (motion.go): figures walk cell by cell at `StepTime` seconds a step (0: `DefaultStepTime`; front/present sets it from an EvMoved's hold), along the last board's `Paths` else the line of squares; new units rise, dropped ones sink. Never a jump (TestNoTeleport).
- Fields the host sets: `Theme` (colours, set before the first Draw; `DefaultTheme` is por's current look, games copy and change it), `Shift` (Shift held, set each frame: health bars, Shift-Tab, Shift-U), `Icons map[Kind]image.Image` (overrides the embedded default icon of an action) + `IconKey` (redraw cue), `Glyph`, `Pointer`, `Stick` (pad right stick pans), `Picking`/`PickKind` (spell list up: bar off), `Asking` (bar shows Yes/No).

## Adapter pattern
Each game writes one small adapter: its battle -> `*combat.Board` (every frame), `combat.Decision` -> its engine call. Examples: por-go/present/remaster/combat.go (`toBoard`, `toAction`), cotab-go/cmd/cotab/combat.go (`board`, `decision`, `input`; it offers ActStep/ActView from its own menu). Game-only extras (weapon, spells) stay in the game; pass their words and icons via `Turn.Labels` and `Presenter.Icons`.

## Status (see ../DESIGN.md)
Fields/kinds only, no behaviour in combat: `Board.Confirm`, `Decision.Pre`, `ActStep`, `ActView`, `Turn.Moves`. `Turn.Labels` IS used by the bar (DESIGN.md says not; stale).

## Gotchas
- Draw needs a window (GPU); create the Presenter lazily after it runs. Units are drawn from their Icon extruded: icons must be RGBA.
- W A S D panning still reads gpu keys; `Key` takes the raw code for digits and bar letters.
- Presenter holds all GPU state; Board is read-only and rebuilt by the game each frame.

## Migrate
The bar icons stay embedded in combat (`icons/*.png`) as the shared default set, so games need not copy them. `Presenter.Icons[kind]` overrides a default; with neither, the hotkey letter shows. `Turn.Icons` is removed (nothing read it). Each game:
- Set `p.Theme = combat.DefaultTheme` (por's look, pixel-identical; cotab/sotsb may copy and change fields) before the first Draw. The zero Theme draws transparent black.
- Set `p.Shift = gpu.IsKeyPressed(gpu.KeyShift)` each frame before Draw and Key.
- Set `Board.StoneTile = 3` (the old fixed value) wherever WallTiles is set.
- Drop any `Turn.Icons` use.
