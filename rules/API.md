# rules

Shared rules, copied from the games and checked against them (`oracle_test.go`). RNG calls
happen in the games' order. Characters come in as `char.Stats`; the game's RNG is `rules.RNG`
(`Dice(n, sides int) int`). Rule globals the affect handlers read live in `Roll` (the games' `rv`).

- `RollDice(rv, rng, n, sides)`
- `AttackHits(rv, rng, att, adj, need, Hooks)` - `adj` is the side modifier the game picks; hooks:
  `Begin` (remove invisibility), `Groups` (attacker/defender affects), `Hit(over)` (por's rules4e `Hits`; nil = accept).
- `TrapHits(rv, rng, c, bonus, Hooks{Group})`
- `SavingThrow(rv, rng, c, typ, bonus, Hooks{Group})` - signed compare for every game.
- `Thac0At(table [][]byte, class, level)`
- `StrengthCode`, `ConHitPoints(c, con, maxHitDice)` (class order: cleric druid fighter paladin ranger ...)
- `DamagePlayer(c, dmg) (over, dying, down)` - the game applies its combat bookkeeping.
- `Config{TrapSkipStoned, TrapRollEach}` with `TrapSkips(status)` and `TrapNextDamage(dmg, roll)`.

Not moved (see RULE2.md): ability-score effects with items/affects, ApplyItems and item effects,
coins/carry, encumbrance (cut), expThreshold, XP split, shop pricing.

## Migrate

### cotab-go
- `e.RollDice` -> `rules.RollDice(&e.rv, e, n, sides)` (rv fields map to `rules.Roll`; or embed one).
- `AttackHits` body -> `rules.AttackHits(&rv, e, att, adj, need, Hooks{Begin: removeInvisibility, Groups: affectGroups})`.
- `SavingThrow` -> `rules.SavingThrow(..., Hooks{Group: func(){ e.affectGroup(c, grpSave) }})`. Behaviour change: signed compare.
- `trapHits` -> `rules.TrapHits`; `damagePlayer` status maths -> `rules.DamagePlayer`.
- `thac0At` -> `rules.Thac0At(rows, ...)` (build `[][]byte` from `thac0Table` once).
- `strengthCode`, `conHitPoints` (maths) -> `rules.StrengthCode`, `rules.ConHitPoints(c, con, classMaxHitDice[:])`.
- Delete: `encumbrance`, `strCarry`-based move limits, `carryLimit` checks, `coinWeight` (see RULE2.md).
- `Config{}` zero value.

### sotsb-go
Same as cotab, plus `Config{TrapSkipStoned: true, TrapRollEach: true}`:
`damageWithMessage` returns early when `cfg.TrapSkips(c.Status)`; the all-members loop sets `dmg = cfg.TrapNextDamage(dmg, roll)`.
`expThreshold` and `giveExp` stay.

### por-go
- `SavingThrow` -> `rules.SavingThrow` (identical, already signed).
- `AttackHits` -> `rules.AttackHits` with `Hooks.Hit = func(over int) bool { return e.rules().Hits(e, att, def, over) }`.
- `RollHit` -> `rules.TrapHits` (pass `byte(bonus)`).
- `RollDice` -> `rules.RollDice`.
