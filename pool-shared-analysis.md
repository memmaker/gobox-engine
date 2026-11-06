# Pool series: shared game-logic analysis

Read-only analysis, 2026-11-06. Scope: the `rules/` packages of por-go (POR), cotab-go (COTAB),
sotsb-go (SOTSB) and pod-go (POD). The game/UI plumbing (Ask/Decision/Event/Mode/wait/then) is
out of scope because the passive migration already covers it.

## Method

- **Function census.** I listed every top-level func and method name in each `rules/` (non-test
  files). POR has 1716, COTAB 1260, SOTSB 1166 and POD 1409.
  - 238 names exist in all four games.
  - 334 more exist in all three of COTAB, SOTSB and POD.
  - 259 exist only in the twins (SOTSB and POD).
- **Body comparison.** For each common name I extracted the body and stripped whitespace and
  comments. Then I took a line-diff similarity for the pairs POR/COTAB, COTAB/SOTSB, SOTSB/POD
  and POR/POD. The figures below are written `pc/cs/sp/pp %`, in that pair order.
- **Reading the numbers.** Above about 85% is near-identical and differs only in field names or
  constants. Between 50% and 85% is the same logic with a few divergent lines. Below 50% is the
  same idea with divergent code. The diff is line-based, so if a game reorders cases or splits a
  function, its number looks worse than the real similarity.
- **Limits.** I checked the ranked candidates by reading the code. I did not read all 238.
  The tools are in the session scratchpad (`fx.sh` extracts a body, `cmp.sh` scores it). They can
  be rerun after each migration step.

**Lineage.** COTAB, SOTSB and POD share one file layout:

- affects, eclcmds*, eclvars, classcalc, charimport, spellids, spellmenu, treasure/coins,
  stats, records.

POR has a different layout (ecl.go/eclops*.go, tables.go, cmdbar, ...), but it uses the same
function names wherever the original DOS code is shared. So the real split is POR versus the
COTAB line, and the twins are the closest pair.

## What gobox already has, and who ignores it

| gobox package | Provides | POR | COTAB | SOTSB | POD |
|---|---|---|---|---|---|
| `gobox/rules` | `AttackHits`, `TrapHits`, `SavingThrow`, `RollDice`, `Thac0At`, `StrengthCode`, `ConHitPoints`, `DamagePlayer`, `Purse`/`Rates` | Purse only; **own AttackHits/SavingThrow** (rules.go:232/254) | uses all | uses all | Purse, Thac0At, StrengthCode, ConHitPoints; **own AttackHits/SavingThrow/RollDice/DamagePlayer** (rules.go:445/468) |
| `gobox/ecl` | VM core, operand decode, shared ops | **ignores it**: own VM in ecl.go/eclops.go/eclops2.go/ecllist.go (3088 lines) | uses it | uses it | uses it (eclvm.go) but **clears `vm.Ops` and re-registers every command itself** |
| `gobox/char` | byte-record `Layout`/`Rec`, `Stats` | ignores | test import only | 1 import | 1 import |
| `gobox/geo` | GEO/WALLDEF parse, `WallAt`/`DoorAt`/`MapCell` | uses | uses | uses | uses. Every game still keeps its own wrapper `WallAt/DoorAt/MapCell/LoadGeo/LoadWalldef`: 5-26 lines each, 8-84% alike |
| `gobox/dax` | DAX loader | uses | uses | uses | uses |

**Quick wins.**
- POR has to adopt `grules.AttackHits`/`SavingThrow`.
- POD has to adopt them too. POD's version adds a Dancing-affect check, a difficulty table and
  `critical`. These should become `Hooks` fields or `Config` data in gobox/rules, not a fork.

---

## Ranked candidates

The ranking is value (lines removed × bug-risk reduction) against effort. For each candidate, the
"Migration" line says when the work can start relative to the passive migration.

### 1. RNG and dice (`Random`, `Randomize`, `Dice`, `RollDice`)

- **Locations.** `engine.go:Random/Randomize` in all four games. `Dice` is in rules.go in all four.
- **Grade:** identical.
  - `Randomize` is 88/88/100/100%.
  - `Random` is the same LCG in all four (`seed*0x08088405+1`, `>>16 % n`). It scores lower only
    because COTAB, SOTSB and POD wrap it with trace/replay hooks (29-30 lines).
  - `Dice` is 58/58/100/100%.
- **To parameterize:** nothing. The `FixedSeed` and the trace hook become fields.
- **Proposed home:** `gobox/rules/rng.go` with this API:
  - `type LCG struct{ Seed uint32; Fixed *uint32; Trace func(n, r int) }`
  - methods `Seed()`, `Random(n) int`, `Dice(n, sides) int`
  - It already satisfies the existing `rules.RNG`.
- **Saved:** about 120 lines. **Risk cut:** high, because the RNG call order is the basis of all
  oracle tests.
- **Migration:** independent of it. Can be done now.

### 2. Combat geometry and pathing

- **Functions covered:**
  - `lineOfSight` (95/95/100/100)
  - `combatInView` (95/93/97/95)
  - `combatOnScreen` (96/96/96/96)
  - `combatSquareFree` (87/90/95/92)
  - `moveReach` (89/91/94/92)
  - `inArc` (89/97/100/93)
  - `buildNear` (95/99/85/81)
  - `nearSort`, `rangeTo` (64/95/100/69)
  - `cellSide`/`cellEdge`, `directionTo`, `step` (46/97/100/49)
  - `stepSquare`, `placeAt`, `placeOne` (79/98/100/80)
  - `groundAt` (75/97/82/73)
  - `boltPath`, `combatX/Y`, `footprint`
  - `coneTargets`: COTAB, SOTSB and POD only, 98/98
- **Locations:** combatmap.go, combatsetup.go and combatturn1.go in each game.
- **Grade:** identical, or the same logic with differing constants. The differences are the map
  size, the `groundTable`/`groundInfo` tables (POR has 66×4, POD has 70×4), and the
  `stepX`/`stepY`/`footprints` tables.
- **Proposed home:** `gobox/combat/grid` (pure functions, no game types). The API is:
  - `type Grid struct{ W, H int; Ground func(x, y int) byte; Occupant func(x, y int) int }`
  - `LineOfSight(g, a, b)`, `Reach(g, from, moves)`, `InArc`, `Cone(g, origin, dir, len)`,
    `Bolt(g, from, to)`, `Place(g, formation, ...)`
  - The tables are passed in as `GroundInfo [][4]byte`.
- **Saved:** about 1,100 lines (roughly 400 lines × 3 removable copies).
- **Risk:** POR's edge fixes (cmap_edge_test) and POD's tile fixes stop drifting apart.
- **Most complete version:** SOTSB/POD (the twins score 100%), plus POR's `cmap_edge` fix
  (check it).
- **Migration:** pure functions, so this can be done now.

### 3. Spell-effect / affect engine

- **Functions covered:**
  - `tickEffects` (92/87/99/85)
  - `checkAffect` (89/84/95/78)
  - `AddAffect`, `addAffectNow`, `FindAffect` (88/88/100/88)
  - `RemoveAffect` (75/92/96/75)
  - `removeAffectFromSide`, `callAffect`, `affectGroup`, `applySpell` (80/97/98/81)
  - `applySpellToSide`, `CureAffect`, `blockAffect`, `healTick`, `hourTick` (53/95/90/48)
  - `readAffects` (100%)
  - the fx* handlers: `fxSleep` 91/94, `fxNeutralizePoison` 91/96, `fxDispelMagic` 96 (cot/sot)
  - `charmHandler`, `confusedHandler`
- **Locations:** affects.go plus affecttypes.go. COTAB/SOTSB/POD each have about 1,700 lines.
  POR has the same handlers in affects1.go/affects2.go.
- **Grade:** the same logic, with differing data.
  - `setupAffectHandlers` (116-141 lines each) scores 1-65% only because the tables are
    differently ordered. It is a type→handler map per game.
  - Affect type numbers differ per game (affecttypes.go is 80% sot/pod and 4% por/sot).
- **To parameterize:** the affect id numbering becomes a per-game table that maps the game's id
  to a canonical handler name. Handlers that only one game has stay registered by that game,
  through data, not branches.
- **Proposed home:** `gobox/rules/affect`, with this API:
  - `type Affect struct{Type byte; Minutes uint16; Data byte; Call bool; Next *Affect}`
  - `Engine.Add/Find/Remove/Tick(minutes)`
  - `Registry map[string]Handler` plus `IDs []string` per game
- **Saved:** about 2,500 lines across the four games.
- **Risk cut:** high. Affect bugs come up repeatedly in the bugclass tests.
- **Most complete version:** POD (1,864 lines, adds dancing and death-ray handlers) and COTAB
  (`auraAffect` with real logic, 37 lines versus 3 elsewhere).
- **Migration:** handlers that emit events touch the plumbing. Do this after each game's flows
  are passive, starting with POD as the reference.

### 4. Character derived stats: STR/DEX tables, HP, level, saves, thief skills

- **Functions covered:**
  - `strHit`/`strDamage`/`strIndex` (100% cot/sot/pod, 50-59% por)
  - `dexAC`/`dexMissile` (80/100/100/80)
  - `maxHitPoints`/`minHitPoints`, `savingThrows` (cot/sot 76, sot/pod 100)
  - `thiefSkills` (sot/pod 95)
  - `hitDiceWeight` (100/100)
  - `conBonusForClass` (96/89)
  - `recalcClassBonuses` (sot/pod 100)
  - `computeExp` (44/83/65/42)
  - `giveExp` (49/92/95/49)
  - `atRaceLimit`, `mayTakeClass` (sot/pod 100)
  - `Level`, `casterLevel` (sot/pod 97)
  - `modifyRaise/Lower` (88/100)
  - `parseStat`
- **Locations:** charcalc.go, classcalc.go, stats.go, train.go and tables.go in each game.
- **Grade:** the same logic, with differing data. The tables are all byte arrays from each EXE:
  - `thac0Table`, `saveTableData`, `thiefBaseData`/`thiefRaceData`/`thiefDexData`
  - `expTable`/`expSteps`, `classMaxHitDice`/`classHitDiceSides`, `conHitPointBonus`
  - `raceAbilityLimits`, `classStatMinimum`, `raceAges`/`ageSteps`
  - the slot tables

  POR uses 8 classes and differently named masks (`hpClassMask`, `hitDieCount`). The COTAB line
  uses 7 base classes plus multi-class combos.
- **Proposed home:** extend `gobox/rules` (Thac0At and ConHitPoints are already there). Add
  `rules.Tables` (a struct of the arrays above, loaded per game from Go literals or JSON). Then
  add `rules.Saves(t, class, level)`, `ThiefSkills(t, race, dex, level)`, `StrHit/StrDamage`,
  `DexAC`, `ExpFor(t, class, level)`, `LevelFor(t, class, exp)` and `MaxHP`.
- **Saved:** about 900 lines of code. The data tables stay one per game.
- **Risk:** medium-high. POR's signed-save fix (rules.go:265, "compared unsigned") has to be
  applied everywhere.
- **Migration:** independent. Can be done now.

### 5. Combat AI

- **Functions covered:**
  - `aiAttackLoop` (71/49/78/38)
  - `aiMoveStep` (70/48/84/46)
  - `aiPickWeapon` (73/38/83/35)
  - `aiCanStep`, `aiFleeCheck` (sot/pod 86), `aiGiveUp`
  - `weaponScore`, `weaponStats` (cot/sot 94, sot/pod 97)
  - `nearEnemies`, `pickAttackTarget` (91/61)
  - `chooseSpellTarget(s)` (91/94)
  - `avoidDistantAttack`, `rangedWeapon`, `throwable`
- **Location:** combatai.go in each game (657-902 lines).
- **Grade:** the same logic with divergent detail. It is clearly one DOS routine per game, with
  per-game patches. POR and SOTSB/POD are surprisingly close (70-84%), while COTAB has drifted.
- **To parameterize:** the per-game knobs become a data struct (`AIConfig`: flee threshold, ammo
  table, spell preference list).
- **Proposed home:** `gobox/combat/ai`, operating on a `Combatant` interface (`Pos()`, `HP()`,
  `Weapons()`, `Side()`). It needs item 2 first.
- **Saved:** about 1,500 lines.
- **Risk:** high value. AI bugs are frequent (aimoffer/aioffhand/aipath tests in POR, aiweapon in
  POD).
- **Migration:** after item 2 and after combat flows are passive. SOTSB/POD go first.

### 6. Attack resolution and the swing loop

- **Functions covered:**
  - `attackSlots` (66/91/90/62)
  - `fullAttack` (59/80/95/59)
  - `attackText`, `hitMessage` (66/86)
  - `KillCharacter`, `killWithMessage`
  - `damagePerson` (93/64), `damageWithMessage`
  - `cmdDamage` (88)
  - `strikeSquare`, `leaveAttacks` (sot/pod 94)
  - `poisonAttack`, `paralyseTarget`
  - `combatResult` (cot/sot 100, sot/pod 85)
  - `cleanupCombatants`, `combatEnd`
- **Locations:** combatturn1.go, combatturn2.go and damage.go in each game.
- **Grade:** the same logic. The message text differs, so it becomes `Data.Texts`.
- **Proposed home:** `gobox/combat/resolve`. The API is
  `Swing(r *rules.Roll, rng, att, def Combatant, cfg) Outcome`, which returns data (hits, damage,
  status changes, message keys).
- **Saved:** about 800 lines.
- **Migration:** these functions emit events. Do it after the passive combat flow (the POD
  reference).

### 7. ECL VM: finish the existing gobox/ecl adoption

- **Locations:**
  - POR: ecl.go, eclops.go, eclops2.go, ecllist.go (3,088 lines, does not use `gobox/ecl` at all)
  - COTAB/SOTSB/POD: eclvm.go, eclrun.go, eclvars.go, eclcmds*.go (about 1,700-2,500 lines
    each, using `gobox/ecl`)
- **Grade:**
  - The VM core is the same idea.
  - The opcodes are the same logic with different numbering: POD's `newScriptVM` remaps the
    operand counts and zeroes `vm.Ops`.
  - `getVar`/`SetVar`/`readTextVar` are 80% cot/sot but only 16% sot/pod, because POD rewrote
    them for the passive VM.
- **To parameterize:**
  - the opcode→semantic table, `[256]OpName` per game
  - the operand counts (already data in POD)
  - the memory map (`ecl.Layout`, already data)
- **Proposed home:** `gobox/ecl`. Move the commands that are shared by meaning (goto, gosub,
  compare, add/sub/mul/div, and/or, random, save/load var, if*, on-goto, clear/print text,
  delay, ...) into `gobox/ecl/ops.go` keyed by name. Each game supplies the
  `[256]string` opcode→name table.
- **Saved:** about 3,000 lines (mostly POR's own VM, plus duplicated cmds in the twins).
- **Migration: blocked.**
  - Do not touch this until the SOTSB VM is aligned with POD's passive shape (handover row SOTSB:
    "next: align with POD").
  - POR's RunECL frame-stack conversion (11 sites) also has to land first.
  - Then lift POD's VM (`b68f757`) into gobox, and port POR and COTAB onto it last.

### 8. Spell tables, spell lists and memorization

- **Functions covered:**
  - `spellTableInit` (sot/pod 83, others below 10% because of table layout)
  - `spellData`, `spellLevel`, `spellClass`, `buildSpellList` (77/71)
  - `addSpellLine` (77/85), `addSpellSorted` (95/97)
  - `sortSpells` (90/100)
  - `canStudy` (94/94), `memorizeOne` (82/80), `scribeOne`, `forgetSpell`
  - `calcSpellSlots`, `spellReduce`, `spellRemoveCurse` (90/94), `spellTargetSetup` (88/89/81)
  - `addScrollSpells`, `scrollUsed`
  - spellids.go (81/77)
- **Grade:** the same logic, with differing data (spell ids, slot tables).
- **Proposed home:** `gobox/rules/spells`. The API is a `Spell{ID, Class, Level, Range, Target,
  Effect string}` table per game, plus `Slots(t, class, level)`, `Memorize`, `Scribe`, `Forget`
  and `SortByLevel`.
- **Saved:** about 900 lines.
- **Migration:** memorize/scribe is a camp flow, so do it after the camp flows are passive. The
  data half can be done now.

### 9. Items: AC, readying, joining and identifying

- **Functions covered:**
  - `itemAC` (97% cot/sot), `ApplyItems` (79/76), `readyItem` (89 sot/pod)
  - `readiedWeapon`, `joinItems` (85), `halveItem` (sot/pod 83)
  - `carried`, `canLoseItem`, `armourMove` (94/100), `setSpeed` (98)
  - `ammoFor` (95/95), `isScroll`, `identifyItem`
  - `itemEffect`/`itemReadyEffect` (sot/pod 95-100)
  - `AddItemCopy`, `readItems`, `savedItems`, `stow`
- **Locations:** items.go, slots.go, itemlook.go, pack.go and charcalc.go in each game.
- **Grade:** the same logic, with data differences (item-type tables, `itemWords` with 256 entries
  in POR and 126 in POD).
- **Gap:** POR's `items.go` (199 lines) and `slots.go` (332 lines) are the most complete. The
  shared party inventory (a user decision) makes one implementation more valuable.
- **Proposed home:** `gobox/rules/items`. The API is
  `ItemType{Hands, Slot, Ammo, ACBonus, ...}` plus `Ready(c, i)`, `AC(c)`, `Join` and `Split`.
- **Saved:** about 700 lines.
- **Migration:** independent of the plumbing. Do it after the shared-inventory work.

### 10. Money and treasure

- **Functions covered:**
  - `appendPurse` (84/100/100), `purseCoin`/`setPurseCoin` (sot/pod 100)
  - `readPurse` (100 sot/pod), `takeRecordCoins`
  - coins.go (174/103/260 lines, 42-50%)
  - treasure.go (COTAB/SOTSB only, 38%)
  - `freeTreasureItems`
  - the temple/shop price helpers
- **Grade:** the same idea. `gobox/rules.Purse` exists, but each game still keeps its own coin
  indexing on top of it.
- **Proposed home:** finish `gobox/rules/purse.go`. Add `Purse.Coin(i)`/`SetCoin`, the split
  (`Split(n) []Purse`, the treasure split) and `FromRecord(layout)`.
- **Saved:** about 400 lines.
- **Migration:** can be done now. The purse is already shared state.

### 11. Party records and the save format

- **Functions covered:**
  - `readRecords`, `writeRecords`, `record`, `setRecord`, `SetName` (100% all four)
  - `makeFileName`, `removeSubstr`, `upper`, `cut`, `sign`, `abs`, `b2i`, `targetOf` (100% all
    four)
  - `LoadCharacter`/`SaveCharacter` (20-75%)
  - `Save`/`Load`
  - records.go (85/100), charfiles.go
- **Grade:**
  - The helpers are identical.
  - The record I/O is the same logic with per-game layouts. That is exactly what `gobox/char.Rec`
    with a `Layout` was built for, but no game uses it beyond tests.
- **Proposed home:** move the identical helpers to `gobox/rules/util` (or delete them in favour of
  `strings`/`cmp`). Characters load and save through `char.Rec` plus a per-game layout JSON.
- **Saved:** about 600 lines.
- **Risk:** high value for import. `charimport.go` (COTAB 544, SOTSB 294, POD 249 lines, only
  12-46% alike) translates POR→COTAB→SOTSB→POD records by hand. With `char.Layout` for every
  game, an import is one layout-to-layout field copy plus a rules hook.
- **Migration:** load/save touches the app/saves plumbing. Do the helpers now and the records
  after the migration.

### 12. Clock, rest and time

- **Functions covered:**
  - `AdvanceClock` (79/93/72/72), `normalizeClock` (80/82/100/75)
  - `normalizeRest` (91 cot/sot), `restFor` (72/62)
  - `fixRestTime`, `hourTick` (95/90)
  - clock.go
- **Grade:** the same logic.
- **Proposed home:** `gobox/rules/clock`. The API is
  `Clock{Min, Hour, Day, ...}.Advance(minutes, onHour func())`.
- **Saved:** about 200 lines.
- **Migration:** can be done now. Rest interruption is part of the camp flow, so that part waits.

### 13. Temples, shops, training and the vault

- **Locations:** temple.go (20-42% alike), shop.go (23-52%), train.go (21-58%), vault.go
  (twins 77%).
- **Grade:** the same idea, with divergent code. Most of each file is flow, which the migration is
  rewriting anyway.
- **What stays as logic:** the temple price table, heal/cure/raise pricing, training cost and
  level-up (which reuses item 4), and the shop buy/sell price ratio.
- **Proposal:** after the passive migration, extract only `rules.TemplePrice(service, level)` and
  `ShopPrice(item, ratio)` into gobox/rules. Prices become data.
- **Saved:** about 300 lines. Lower priority.

---

## Twins: SOTSB and POD

Of the 259 functions that only the twins share, most are 85-100% alike:

- campflow/itemflow/lootflow/partyflow/rulesflow, alterflow, vault, digitized, names
- `charstats.go`, `porstate.go`, `records.go` and `itemlook.go` are 100% identical files.

POD's VM rewrite has caused the main divergence, which sits in eclcmds*/eclrun/eclvars (17-36%).
Once SOTSB is aligned with POD's passive VM (already the plan), the remaining non-VM `rules/`
code is roughly 70-90% common. That is about 8-10k lines that could go to a shared
`gobox/pool2` package, or simply into the gobox packages above.

**Recommendation:** after the passive migration, diff SOTSB against POD file by file. Make POD
canonical (it has more tests and fixes) and move the twin-only flows into gobox.

## Identical trivial helpers (delete now)

These are 100% the same in all four games:

- `abs`, `sign`, `b2i`, `cut`, `upper`, `removeSubstr`, `makeFileName`, `targetOf`
- `formationEdge`, `combatInfo`, `readAffects`, `readRecords`, `writeRecords`, `record`,
  `setRecord`, `SetName`

That is about 100 lines. Most can become the standard library (`strings.ToUpper`,
`strings.Replace`) or one `gobox/rules/util.go`.

## Bug-class checks to broadcast

- **Signed save comparison.** POR fixed `Saves[typ] <= int8(save)`, which used to be compared
  unsigned. Grep every game's `SavingThrow`/`saveTargetOf` for an unsigned compare.
  - SOTSB and COTAB go through `grules.SavingThrow`.
  - POD's `saveTargetOf(c, typ) <= e.rv.Save` compares bytes. Check whether a negative bonus can
    wrap there.
- **Forked shared routines.** Any function listed above that scores below 90% between twins
  without a decomp comment explaining it is a likely bug. Rerun `cmp.sh` on the `all4` and
  `three` lists after each migration step.

## Top 10 (value ÷ effort)

| # | Candidate | Grade | Lines saved | Can start |
|---|---|---|---|---|
| 1 | RNG/dice to `gobox/rules.LCG` | identical | ~120 | now |
| 2 | Combat grid/LOS/reach/cone/placement to `gobox/combat/grid` | identical / constants | ~1,100 | now |
| 3 | Derived stats (STR/DEX/HP/saves/thief/exp/level) data-driven in `gobox/rules` | same logic, data | ~900 | now |
| 4 | POR and POD adopt the existing `grules.AttackHits`/`SavingThrow` (add Hooks for Dancing and difficulty) | same logic | ~100 + bug fix | now |
| 5 | Affect engine plus handler registry to `gobox/rules/affect` | same logic, data | ~2,500 | after flows passive |
| 6 | ECL VM: lift POD's passive VM to `gobox/ecl`, named-op tables, POR off its own VM | same idea / data | ~3,000 | **after** SOTSB aligns to POD and POR's RunECL stack |
| 7 | Combat AI to `gobox/combat/ai` | same logic, divergent detail | ~1,500 | after 2 and passive combat |
| 8 | Spell tables, memorize and scribe to `gobox/rules/spells` | same logic, data | ~900 | data now, flows later |
| 9 | Char records via `gobox/char.Layout`, generic import, shared trivial helpers | identical helpers / layouts | ~600 + charimport ~1,000 | helpers now, records later |
| 10 | Items, purse coins and treasure split on `gobox/rules` | same logic, data | ~1,100 | after shared-inventory work |

After the migration, the SOTSB/POD twin merge adds roughly another 8-10k lines.
