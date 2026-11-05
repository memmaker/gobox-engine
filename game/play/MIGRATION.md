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

## Order

1. gobox: an adapter `front` can drive from `play.Game`, beside the old path
   (a game is one or the other). `Legacy{Decision}` lets a game move one
   state at a time.
2. COTAB first (newest walking contract): `Exploring` + `Step`/`Face`, then
   `Talking`/`Reading`, then `Fighting` + `EndTurn`, then shop, loot, camp,
   creation.
3. SOTSB, POD, BR1, BR2 the same way, one state per commit, `go test` each.
4. POR last and only if the user lifts its freeze; until then gobox keeps the
   old path for it.
5. When no game uses an old field or `Legacy`, delete it from game.go.

## Open points

- Each game-specific verb (cast, turn undead, appraise, ...) gets a typed
  action in `play` once two games share it; until then `Legacy`.
- Saves: gobox saves; the game needs `Save() []byte` / `Load([]byte)`
  (not in the draft yet).
