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

- `Purse{Gold, Copper, Gems, Jewelry uint64}`: the one shared party purse (no per-member coins). `Add(q, Rates)`,
  `AddCopper(n, Rates)` (copper rolls into gold), `Pay(gold) bool`, `Empty()`, and `FromCoins(coins []uint64, Rates)`
  which turns a record's coins (kinds in `Rates.Coin`, then gems, jewelry) into a Purse. `Rates{CopperPerGold, Coin}` is data;
  `GoldBoxRates` = 200 copper per gold, coins 1/10/100/200/1000 copper. Oracle: por's money.go in `purse_test.go`.

Not moved (see RULE2.md): ability-score effects with items/affects, ApplyItems and item effects,
carry (cut), expThreshold, XP split, shop pricing.

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
- Purse: `purse` -> `rules.Purse`, fields exported (`gold` -> `Gold`), `add`/`addCopper` take `rules.GoldBoxRates`, `coinsToPurse`
  -> `FromCoins`. `pay` keeps its message and calls `Pay`. Save file stays byte-identical: `appendPurse` writes the uint64
  fields saturated to uint32 (ponytail ceiling: a purse over 4294967295 saves as that; widen the format when it matters).

### Purse in cotab-go and sotsb-go (one shared party purse)
Coins in the record (7 uint16 at +251: copper, silver, electrum, gold, platinum, gems, jewelry) and the treasure pool
(`treasureCoins [7]int32`) both become `rules.Purse`; `coinValues`, `poolGold`, `goldWorth`, `setGold`, `setPoolGold`,
`addPlatinum`, `moveCoins`, `poolMoney`, `sharers`-based splitting and `coinWeight` go.
- Load (party join, .SAV/.GUY read): `e.purse.Add(rules.FromCoins(widen(rec.Coins), rules.GoldBoxRates), rules.GoldBoxRates)`, then
  zero `rec.Coins`. Every member's coins end up in the one purse.
- Write-back (.SAV/.GUY): write the record's coin words as zero; the purse is not in the record. The .SAV game save keeps the
  purse in the game's own trailer (as por: gold, gems, jewelry, copper as uint32, saturated; ponytail ceiling uint32). A .GUY
  exported on leaving the party carries no coins; the purse stays with the party. An imported .GUY with coins (cotab's Pool
  character has 300 platinum at +251) pays its coins into the purse on join.
- Shop prices: `shopPrice` is unchanged (price in gold, uint16 from the item); buying is `e.pay(uint64(price))` against
  `purse.Gold` (fold copper: it is under one gold, so gold alone decides). The member-then-pool fallback in `shopBuy` goes: one
  purse. Selling adds `AddCopper(value)`. cotab's `(copper + 100) / 200` rounding in `goldWorth` goes with it.
- Treasure pool: the pool is a Purse (`e.loot`, as por). ECL treasure commands set it with
  `FromCoins(args, GoldBoxRates)`; taking the treasure does `e.purse.Add(e.loot, rates)` and clears it. The take/share/drop
  per-kind screens and `treasureExpDivisor`/`Factor` experience (computed from the pool's kinds before Add) are the places
  that need the kind split: compute experience from the raw `coins` before `FromCoins`. Scripts that read or set a member's
  coin words use por's `purseCoin`/`setPurseCoin` (platinum, gold, copper view of the purse).
