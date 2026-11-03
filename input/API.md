# gobox/input API (pure Go, no window; module gobox/input)

Names what a key does. The game owns the key table (code -> Act by context); screens switch on Act, never on the code.

## Types and consts
- `Act` (uint8): None, Up/Down/Left/Right, PgUp/PgDn/Home/End, Enter, Space, PadA (pad A: Enter except where Enter opens a menu), Esc,
  Member1.. (Member1+i = F(i+1)), `Members = 8`, PrevMember/NextMember, More/Less,
  adventuring: Minimap, Overview, Map, Party; any time: Help, SaveLoad, PartyMenu, Sound, Font, FontPrev, Backspace, Delete, Note;
  combat: CursorUp/Down/Left/Right, TurnLeft/TurnRight, Tab, EndTurn, Options (pad Start), Log.

## Funcs
- `Cursor(a Act) bool`: a is a cursor key (arrows, PgUp/PgDn, Home/End).
- `Repeat(held, delay, rate int) bool`: held = ticks the key has been down (1 on the first tick, 0 up). True on 1, again at delay, then every rate ticks.
  Pure: the caller counts ticks and picks delay/rate (por: tps/3 and tps*3/20).

## How a game feeds it
The host keeps the binding table and a held-ticks counter per key/button, calls Repeat each tick, and passes the resulting Act on to ui (MenuNav, Pager) or combat (`Presenter.Key(act, rawKey)`). Nothing comes back but a bool.

## Gotchas
- The table itself is NOT here (por: present/remaster/bindings.go); gobox/ui `Table` holds a game's data.
- Act has no state; "held" must be counted by the caller. No global state, no gpu import.
- `Act` values are shared across games: add new ones at the end of their group only if every game's table still compiles.
