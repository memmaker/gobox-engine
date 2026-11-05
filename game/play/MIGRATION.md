# Moving to the typed contract (gobox/game/play)

Goal: the game is a typed, purely reactive interface. gobox reads `State()`,
switches over its type, offers the state's `Choices`, sends one `Action` to
`Do`, plays the `Outcome`, reads `State()` again. The game never acts alone.

## What goes away from game.Rules / game.State

| Old | New |
|---|---|
| `Step()` (game runs time, AI, scripts) | `EndTurn` outcome carries the monsters' moves; `AdvanceTimeBy` for rest/waiting; scripts advance on `Continue`/`Reply` |
| `Pending() *Ask`, `State.Pending`, `Ask`, `Answer`, `Cancel` | the state's `Choices`; gobox builds the dialog |
| `Refusal{Ask}` (door question, LootLeave "sure?") | plain error from `Do`; the follow-up actions are in `Choices` |
| `Mode`, `Window`, `Holding`, `Screen`, `Pacing`, `Flags` | the State's type (gobox picks screen and pacing) |
| `Read`, `Page.Prompt`, `PaneRows`, `Intro`, `Sequence` | `Reading` / `Talking` states; paging and timing are gobox's |
| `Map.Edge` walls checked in gobox walk.go | gobox sends `Step`; `Do` refuses at a wall |

Old decisions with no action, because the choice is gobox's: `Answer`,
`Cancel`, `Select` (menus), `Items`, `Maneuver` (open a menu), `Wait`,
`Ranged`, `Aim` (now `Shoot`/`Cast` with the square), `RunEvent`, `Arrive`
(now `Step`), `Rest`/`Camp`/`LeaveCamp` (now `AdvanceTimeBy`, `MakeCamp`,
`BreakCamp`), `Save`/`Load` (now `Game.Save`/`Load`), `SetSpeed` (pacing is
gobox's), `SetIcon`/`Create` (now `SetChoice`), `KeepDraft` (`KeepCharacter`),
`AddChar`/`DropChar` (`AddCharacter`/`DropMember`), `Trade` (`Give`),
`TakeLoot` (`Take`), `Cure` (`Heal`).

## Order

There is no bridge: every action and state is one explicit type in `play`,
and a game moves over whole.

1. gobox: a front path driven by `play.Game`, beside the old one; a game
   runs on one or the other.
2. COTAB first: its rules package implements `play.Game` completely, tests
   pass, its cmd switches to the new path.
3. SOTSB, POD, BR1, BR2 the same way, one game per change.
4. POR last and only if the user lifts its freeze; until then gobox keeps the
   old path for it.
5. When no game uses the old path, delete game.Rules, Decision and the
   flow fields from game.go.

## Open points

- A verb only one game has still gets its own type in `play`.
- Sound and music cues ride in `Outcome.Events`; where the music is set by
  the place, `Where.Music` stays state.
