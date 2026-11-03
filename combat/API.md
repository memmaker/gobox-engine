# gobox/combat API (3D tabletop combat map; needs gpu, input)

Reads a `Board` snapshot, draws figurines on a wooden table, returns `Decision`s. The camera turns in eighths (XCOM style).

## Snapshot types (game -> combat)
- `Board`: Ground [25][50]Square, Tiles []image.Image, WallTiles, Floor, Dungeon, Units, Cursor/CursorOn/CursorAt, Marks, Reach, Paths map[Point][]Step, Targets (unit IDs), Turn *Turn (nil = nobody acts), Panel, Log, Area/AreaAt/Skipped, AimWhy, Effects, View, Strike, Confirm.
- `Square{Tile,Wall,Height}`, `Step{Point,Free}` (Free: leaving provokes), `Effect{X,Y,Pic}`, `Strike{Seq,From,To,Hit,Damage}` (new attack = new Seq).
- `Unit`: ID, X/Y/Size, Squares (1x1 or 2x2), Party (blue base, else red), Facing 0..7, Icon + IconSize (24 or 48), Active, Name, HP/MaxHP/AC/Hit/Dmg/Status.
- `Turn`: Unit, `Can []Kind` (buttons in the game's menu order), Aim, Casting, Moves, `Labels map[Kind]string` (button words), `Icons map[Kind]image.Image`.

## Decisions (combat -> game)
- `Decision{Kind, N, X, Y, Pre}`. Kinds: ActNone, ActWalk, ActStep, ActAttack, ActAim, ActPoint, ActClose, ActYes, ActNo, ActWait, ActEnd, ActQuick, ActGuard, ActBandage, ActFlee, ActTurn, ActCast, ActItems, ActRanged, ActManeuver, ActView.

## Presenter (one per combat; zero value works)
- `Draw(dst, *Board)`; `DrawBar(dst, glyph, *Board, k, bottom)` the action bar at scale k; `HideBar()`.
- Input: `Key(act input.Act, k uint16) bool` (W A S D pan, arrows cursor, Q/E turn, Home reset, digits/letters buttons, Tab cycles targets, Enter, Esc = No, Backspace ends turn), `Point(x,y)`, `Click(x,y) bool`, `Zoom(in, speed)`, `Pick(x,y)`, `Hovered()`.
- Out: `Actions() []Decision` (drain once a frame). `Float(id,n,heal)` starts a damage number.
- Fields the host sets: `Icons map[Kind]image.Image` + `IconKey` (redraw cue), `Glyph`, `Pointer`, `Stick` (pad right stick pans), `Picking`/`PickKind` (spell list up: bar off), `Asking` (bar shows Yes/No).

## Adapter pattern
Each game writes one small adapter: its battle -> `*combat.Board` (every frame), `combat.Decision` -> its engine call. Examples: por-go/present/remaster/combat.go (`toBoard`, `toAction`), cotab-go/cmd/cotab/combat.go (`board`, `decision`, `input`; it offers ActStep/ActView from its own menu). Game-only extras (weapon, spells) stay in the game; pass their words and icons via `Turn.Labels` and `Presenter.Icons`.

## Status (see ../DESIGN.md)
Fields/kinds only, no behaviour in combat: `Board.Confirm`, `Decision.Pre`, `ActStep`, `ActView`, `Turn.Moves`. `Turn.Icons` is not read (use `Presenter.Icons`). `Turn.Labels` IS used by the bar (DESIGN.md says not; stale).

## Gotchas
- Draw needs a window (GPU); create the Presenter lazily after it runs. Units are drawn from their Icon extruded: icons must be RGBA.
- Keys read raw gpu state too (Shift for Shift-Tab and Shift-U); `Key` takes the raw code for digits and bar letters.
- Presenter holds all GPU state; Board is read-only and rebuilt by the game each frame.
