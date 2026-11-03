# gobox/input API (pure Go, no window; module gobox/input)

Names what a key does. The game owns the key table (code -> Act by context); screens switch on Act, never on the code.

## Types and consts
- `Act` (uint8): None, Up/Down/Left/Right, PgUp/PgDn/Home/End, Enter, Space, PadA (pad A: Enter except where Enter opens a menu), Esc,
  Member1.. (Member1+i = F(i+1)), `Members = 8`, PrevMember/NextMember, More/Less,
  adventuring: Minimap, Overview, Map, Party; any time: Help, SaveLoad, PartyMenu, Sound, Font, FontPrev, Backspace, Delete, Note;
  combat: CursorUp/Down/Left/Right, TurnLeft/TurnRight, Tab, EndTurn, Options (pad Start), Log.

## Funcs
- `Cursor(a Act) bool`: a is a cursor key (arrows, PgUp/PgDn, Home/End).
- `Repeat(held, tps int) bool`: held = ticks the key has been down (1 on the first tick, 0 up), tps = the host's ticks a second. True on 1, again after `RepeatDelay` (0.3 s), then every `RepeatRate` (0.1 s): package vars in seconds, one for all games, rate at least 1 tick.
  Pure: the caller counts ticks.

## How a game feeds it
The host keeps the binding table and a held-ticks counter per key/button, calls Repeat each tick, and passes the resulting Act on to ui (MenuNav, Pager) or combat (`Presenter.Key(act, rawKey)`). Nothing comes back but a bool.

## Gotchas
- The table itself is NOT here (por: present/remaster/bindings.go); gobox/ui `Table` holds a game's data.
- Act has no state; "held" must be counted by the caller. No global state, no gpu import.
- `Act` values are shared across games: add new ones at the end of their group only if every game's table still compiles.

## Migrate
`Repeat(held, delay, rate)` became `Repeat(held, tps)`; the repeat is now `RepeatDelay` 0.3 s, then every `RepeatRate` 0.1 s in every game (PROPOSED, awaiting approval) (was por: tps/3, tps*3/20; cotab: 16 then 4 ticks at 60 tps). Each game:
- por-go: wherever it called `input.Repeat(d, tps/3, tps*3/20)`: `input.Repeat(d, tps)` (tps = 120).
- cotab-go: `cmd/cotab/keys.go` lines 97 and 102: `input.Repeat(x, 16, 4)` -> `input.Repeat(x, 60)` (gpu.SetTPS(60)).
- sotsb-go: same, with `tps` (120).
