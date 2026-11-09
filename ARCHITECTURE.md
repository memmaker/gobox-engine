# Architecture: engine, rules, adventure

DRAFT for Felix's review, revision 2 with his 11 rulings. Nothing uses `core/`; no migration. Settled parts: ONE-PATH.md "Next: three components".

## Dependencies

```
 gobox/core        Rules, Adventure, neutral types and id types, RNG; imports nothing of ours
 core/dnd  core/xxvc   the concepts of one world: races, classes, spells, monster and item rosters
     |          |
 rules1e    rulesxxvc   adventures: POR, COTAB, SotSB, POD | BR1, BR2     (leaves: core + one world)
     \________ engine (gobox): main loop, one script machine, combat and ship loops, front, saves ________/
 cmd/<game>: main wires one Rules and one Adventure.   cmd/import: originals to canonical, once.
```

Leaves know neither gobox nor each other (`imports_test.go`, list empty until a leaf exists). A rule never calls back: all it needs arrives as a value. Sizes: `Rules` 7 + `CharacterRules` 11 + `CombatRules` 9; optional `ShipRules` 5; `Adventure` 9.

**What "any adventure on any ruleset" now means: it does not.** An adventure names `dnd.Hobgoblin` or `xxvc.Terran`; `Mismatches(Needs, Catalog)` fails at load on any id the ruleset lacks. POR runs on any ruleset that fills the `dnd` Catalog (1e, rules4e), never on XXVc. Crossing worlds would take an explicit conversion leaf; none is planned.

## 1. Scripts see typed facts only

The script machine has **no character memory**. Direct or computed, an address in the character window goes through `CharBinding` to `Rules.Fact`/`SetFact`, which have no side effects. An unbound address is a run-time error, never a byte. Consequences are `Action`s through `Resolve`.

| Today | Replacement |
|---|---|
| COTAB `cvActive` read clears `justJoined` (eclvars.go:156-159, set at eclcmds.go:171-173) | The flag only means SelectMember ran past the party's end. That selects `NoChar`; `FactActive` of `NoChar` reads inactive, purely. A second read then differs: old returns the previous member's state. Confirm no script relies on it |
| POR off 0 write 0 sets `memberGone`, 0x100 write 0 `memberDied`, then SelectMember's drop bit (eclops.go:178-194); COTAB `nameCleared`, `activeCleared` | Script patch at import: the sequence becomes one `RemoveMember` action |
| POR status write 0x87 stones | `SetStatusAction` from the patch; `SetFact` only stores |
| BR 76 reads 128 if the skill byte > 0; 268 packs r[214], r[76], r[215] (ecl.go:98, :109, :193) | `KnownFact` with `Values{128: true}`; 268 is a `CharBinding.Bits`: `FactSide` bit 0, `KnownFact` bit 6, `FactSideLock` bit 7 (r[215]'s name unverified) |
| BR 16..185 raw and party-memory double write; POD `charRecordBytes()[a-charVarFirst:]` | Gone. Each used offset gets a binding (survey below); others error |
| POR `effectTable` over 0x4900 (eclops.go:427+) | Patch to `HasCondition{CondFeatherFall}` and its Textile House case |

**Scan** (all decoded blocks; tool in the scratchpad, not committed). The old count saw only variable operands, so it missed every GetTable/SaveTable base and immediate operands:

| Game | Table sites reaching the window | Targets and index source |
|---|---|---|
| POR | 5 | block 9: rel 0, 1, index 187..195 plus AND and Random(9): the coin words, so `CoinFact`. Block 14: rel 32, clears memo |
| COTAB | 2 | blocks 33, 34: rel 32 memo, rel 0 name; index loops |
| SotSB | 1 + 8 | block 34 rel 27. Eight more with base below the window (rel -112, -62, index member×2, +128 steps seen) |
| POD | 12 + 42 | rel 30 (memo, write 255, index up to Random 99 + more); 42 below it (rel -124, -24: needs index ≥ 124 or 24) |
| BR1, BR2 | 0 | |

Also one immediate: POD op 0x8E arg 0, rel 75. Other bases are the code image (POR 260, COTAB 148, SotSB 190, POD 332, BR 102, 122 sites) or plain memory.

Unverified: this is static, listing each index's writers. It does not bound them all and I did not run the games. The real proof is the machine's error: run it first on the current VM as a logging trace through each game's script playthrough (cheap, targeted), and bind what appears.

## 2. One script machine: merge ecl-br into gobox/ecl

Merge: meaningful. Both run `script.Instr` with `Arg.N int`; the loops are almost the same code.

| Command | gobox/ecl | ecl-br | Merged |
|---|---|---|---|
| value type | uint16 | int32 | int32; POR-line hosts zero-extend, store truncates. Add, sub, mul, div keep the low 16 bits' results |
| operands | `GetVar(a)` | `getVar(a, t)`, width from type: 2,3 = 2 bytes; 4,5 = 4; else 1 | `GetVar(a, width)`; types 4, 5 never occur in POR-line |
| COMPARE | unsigned | signed int32 | signed over zero-extended: same results. **BR word sign extension unverified** |
| ADD SUB MUL RANDOM SAVE, IFs, ONGOTO, ONGOSUB, GOTO, GOSUB, RETURN, AND, OR | same | same | one |
| DIV | remainder to `RemainderAddr` by Layout, POR none (`ecl.NoRemainder`) | Party[831:] width 4 | Layout says where, or none: data |
| RANDOM | byte n, ++ below 255 | same; PCG | one, `core.RNG` |
| GETTABLE SAVETABLE | word memory | byte-addressed | `GetVar` with width; bindings catch the window |
| text flags (COMPARE on text) | same | same | one |
| skip tables | `vm.Args` | `e.skip` switch | `vm.Args` by dialect |
| shared command numbers | lowest (Raw) | highest; 59, 60, 68 "pc stays" | Version's table, data |
| FOR, NEXT 70, 71 (Area[502/503], LoopAt) | none | yes | default op, dialect table enables |
| blocks and waits | passive VM | `Frame`, `NextKind`, `BlockThen`, `EclWait` | **Compare before step 3; unread line by line** |
| memory | `Layout` struct | constants (Area 19200.., Party 31744.., Aux 31232.., Code 32768..) | `Layout` values per dialect |
| 30, 31, 34, 35, 50, 54, 57, 58, 62-69, 72, 76 | none | host ops | engine ops on `Facts`, `Check`, `Action`; `SpaceFight` (31) on `ShipRules` |

Plan: (1) widen `gobox/ecl` to int32 and width, all four Pool traces unchanged; (2) add dialect tables for V4 and FOR/NEXT; (3) move BR's ops over with its traces passing; (4) delete `ecl-br`'s VM, leaving ship, record, tables for `rulesxxvc`. POR's overrides `Compare` and `GetTable` go by patches (below), `Exit` empty by patching its scripts, `NoRemainder` by Layout. No Quirk list.

## 3. POR's native scripts become scripts

`nativeScripts` and `purseWhos` (por-go/rules/native.go, eclops.go:83, eclops2.go:553, script.go:36-45) go. Patches in the area text, using typed facts:

| Native | Script |
|---|---|
| TavernGamble | `Okay` text, InputNumber, compare `CoinFact{Gold}` against 5000 and the wager, RANDOM 10: ≥8 double, >2 back, else lose, "WILL YOU GO AGAIN?" |
| TavernTale | RANDOM 22, a journal command with `Tavern Tale n`, RANDOM 3 to DrunkenBrawl or exit |
| PhlanHistory1-6 | journal command with the entry |
| PayPlatinum, TavernGamble purse Who | Who removed; `CoinFact` is the party's whoever is selected |

Needs: a journal command in the patch vocabulary; the shown coin split (`purseCoin`, money.go:65: Gold%5, Gold/5) stays a view of the same purse. COTAB-style per-member coin writes also land in the purse.

## 4. Adventures name, rulesets number

"4 hobgoblins" is `LoadMonster` with a table: `ScriptSet.Monsters` maps `(set, index)` to `MonsterID`. Constant indices can also fold at import; variable ones (POR 24, COTAB 8, SotSB 21, POD 69, BR 9, 12 sites) use the table at run time, an index with no row is an error. `Rules.Spawn(MonsterID, *RNG)` supplies every number. Items: `ItemRef{ID, Plus, Charges, Count, Cursed}`, `Rules.Item(ItemRef)`; `FindItem` terms likewise (4, 3, 1, 7, 0, 0 distinct constants). Spells as before by `SpellTerm`.

Scan of `MON*CHA` (records / names / names with differing stat blocks): POR 172/103/12 (KOBOLD ×2, VAMPIRE ×2), COTAB 81/68/9 (RAKSHASA ×3), SotSB 70/68/2, POD 122/118/2, BR1 54/49/5, BR2 63/55/6. So one id per **distinct block**, not per name: the importer generates rosters, a person names variants. Consequence to accept: `core/dnd` and the 1e tables carry every module's named NPCs (Tyranithraxus), since the ruleset holds all stats. Across games the same name will also differ; each block gets an id, merged only if equal.

Rule-scale gates: `Compare(PartyStrength, 70)` (POR 18 sites, COTAB 2; others none) becomes `Check{PartyAtLeast{Tier}}` by patch: the ruleset decides what "strong" means. Three `Tier`s is a guess until the sites are read. `Census` over abilities is world-shared; thresholds in scripts stay scripts' own.

## 5. Asset data

**Import, not on-demand decoding, of originals.** `cmd/import` (gobox; may use dax, geo, ega, script) runs once on the user's machine, turning originals into canonical files in the adventure's folder: PNG (animations: frames plus delays), WAV, script text, walls. At run time `Adventure` serves `Picture(PictureID)`, `Sound(SoundCue)`, `Map(MapID)`, `Script(AreaID)`, on demand from those files, returning decoded types only. No file name, block or format reaches the engine. An adventure may read originals itself if it brings its own decoder (a leaf may not import `gobox/dax`).

**Rule tables likewise** (ruling 6). One generator, `exe/cmd/tablegen`, already POD, COTAB, SotSB (`tables.json`, `TestTablesMatchEXE`), extended to typed values in the ruleset: thac0, saves, exp, monsters, items, ships. Dropped from run time: BR2 `dsimg.go` `os.ReadFile(START.EXE)`, BR1's `dsImg` bytes, `tab`, `fightData()`, cotab `engine.go`'s `gobox/exe` import, MON*/ITEMS at run time, words (BR's `word()`). Generated Go is checked in; fine for private use, a question if ever published.

## 6. Saves: one-way

The importer reads an original save and CHRDATA/.GUY once and writes a canonical save. Nothing writes originals. `Rules.Save` is characters, the engine's part is script memory, board, RNG, and bindings carry the rest. POR's byte-for-byte save stops being a test.

## 7. One RNG

`core.RNG{State}`: **the engine owns it**, seeded at start and saved; every rolling Rules method takes `*RNG` as a parameter, so no callback and no hidden stream. Script RANDOM draws the same. Sequence: Turbo Pascal LCG. Testing: fixed seed, full reproduction.

Trace interaction: POR, POD, COTAB, SotSB keep original traces as long as draws stay in the same order, including for monsters' HP (now Spawn) and scripts. BR1, BR2 use PCG (`e.rng`), so they get new sequences once: re-record their traces at that step, separately from anything else. Anything else drawing outside: `por dasp.go:166` voice picks from global `rand`. See questions.

## 8. Ship fights

`core/ship.go`: data (`Ship`, `Gauge`, `ShipFight`, `ShipEncounter`), actions (fire, boost, jury-rig, ram, dodge, close, retreat, reload), `ShipRules` (5 methods). Adventure: script's SpaceFight names `ShipEncounter`. Rules: all of spacefight.go's arithmetic. Engine: loop, input, `game.SpaceFight` view; stage constants `Ss*` become its loop. Crew turns are ordinary `Allowed`/`Resolve` with `ShipFight` as situation. `Rules.Ships()` is nil for a ruleset without `FeatureSpaceCombat`, which then cannot load such an adventure.

`Board` thickened: `Height`, `Ground`, `ZeroG`, `Free`.

## 9. One 1e ruleset: every variation, for Felix

Ruling: one ruleset, flagged small variations. A flag belongs to the adventure's rules config or a table row. "Rec" is mine.

| # | Rule | POR | COTAB | SotSB | POD | Rec |
|---|---|---|---|---|---|---|
| 1 | Saving throw compare | signed, rules.go | was unsigned | was unsigned | | done: signed (RULE2 a) |
| 2 | Initiative | `nextCombatant` d100 per pick combat.go:32 | turnSetup | `buildTurnOrder` :74 | :75 sort, coin toss | decide; flag breaks no trace |
| 3 | To-hit by difficulty | none rules.go:222 | none rules.go:314 | none rules.go:351 | party ±2 table rules.go:423, 438 | flag |
| 4 | Critical hits | none | none | none | rules.go:445 | flag |
| 5 | Monster HP by difficulty | none | none | ×(level+offset)/4 eclcmds2.go:353-363 | none | flag |
| 6 | Difficulty storage | none | none | area word 500 camp.go:234 | game byte 38 alterflow.go:127 | one `WordDifficulty` |
| 7 | Hide exp | none | none | area 501 aftercombat.go:124 | game 30 :143 | `WordHideExp` |
| 8 | Cleric prime exp | aftercombat.go:86 own constants | Wis>15 :129 | Wis>=15 :136 | check | decide; SotSB looks like the original's off-by-one |
| 9 | Exp table lookup | | train.go:376, runs into the next row | train.go:295, 0 past end | train.go:467 | table; past end = cap |
| 10 | Trap damage | | base | `TrapSkipStoned`, `TrapRollEach` damage.go:36, :154 | | RULE2 b, flags |
| 11 | Shop price | fixed item price | factor word shop.go:19 | same shop.go:22 | byte, platinum shop.go:18 | RULE2 e: flag or table |
| 12 | Encumbrance | none | cut | cut | | RULE2 c |
| 13 | `armourMove` | charcalc.go:350 | :185 | :230 | :240 | undecided in RULE2; decide |
| 14 | Charisma percent | ecl.go:156 | eclvars.go:86, <=3 | eclvars.go:107, ==3 | | one; only below-3 differs |
| 15 | Spell roster, numbers | 56 | 100+ | 100+ | 126, item powers | union by id; `SpellTerm` per module |
| 16 | Memo slots | 21 | offsets 32..112 | 27..101 | 141 | Ruleset limit: flag or max, decide |
| 17 | Dual class | extra dualclass.go | | | | flag, or all |
| 18 | Knock | extra slot ecl.go:201 | door choice adventure.go:233 | :256 | | check |
| 19 | Rule tables (thac0, saves, exp, str, dex, thief) | 8 classes, saveTable charcalc.go:94 | 480 bytes | 635 | 740 | table per game or merge: decide. Means numeric parity vs. one table: real cost |
| 20 | Attack hook | rules.go:230 `Hits` (rules4e) | | | | ruleset-private |
| 21 | Dancing affect | | | | affects.go:140 | condition in 1e; check others have it |

Gaps left empty were not checked. The census (pool-shared-analysis.md: 238 common of 1166-1716 functions; twins SotSB, POD 85-100%) says more lie in `computeExp`, `giveExp`, `attackSlots`, `thiefSkills`, affects, and AI: a function-by-function diff at step 5 will extend this table. BR's "no 2e game" stands; XXVc has its own ruleset, BR1 and BR2 two tables.

## Scripts' data and today's code

Survey of direct character use, per game (reads through bindings):
POR 21 words; COTAB 18; SotSB 14; POD 31; BR1 11; BR2 8. Conditions queried: POR 19; COTAB 20; SotSB 3; POD 25, 56, 72, 126 and a word. Which is why `Fact` and `Condition` are small.

Where code goes: rules leaves take combat rules, affects, spells, creation, training, tables. Engine takes combat flow, map, VM, ops, presentation. Adventure takes areas, journal, lore, scripts and patches, assets. Importers take savefile, tablegen, decoders. `ecl-br` splits by section 2.

Hard spots left: combat AI one long routine (por `computerTurn`); `game.Data` mixes module, series look and rules; narration words (rules produce text: table at import); clock owned by the engine with `Elapse`; `WorldWord` list unsurveyed in part.

## Migration order

Each step keeps all six games on traces, except the named BR RNG step. 1. Settle this. 2. Machine merge (section 2), logging computed window hits. 3. Bindings and patches, POR first, then errors on. 4. Importers and asset ids. 5. POR adventure and rules leaf. 6. Engine loop. 7. SotSB, COTAB, POD on 1e, extending the variation table. 8. XXVc, BR's RNG, ship fights.

Unverified: nothing here has run on a game. Only core's build, vet, tests. No ruleset or adventure implements these interfaces yet; expect change.

## Remaining questions

1. Section 9: each row.
2. "One RNG over everything": presentation draws (voice, `dasp.go:166`) in the stream would make traces depend on frames. Recommend a second `RNG` seeded from the first, engine-owned, for presentation only.
3. Monster ids per stat block, with module NPCs in the 1e tables: accept? The alternative is an adventure-owned override, which ruling 4 seems to forbid.
4. `Tier` for gates: read 20 sites, then fix the names?
5. Do adventures reading originals need a shared decoder leaf (dax, ega outside gobox)? Recommend no: importer only.
6. Rule tables generated from original EXEs and checked in: private use only?
7. Glossary: fix "character" (includes monsters), "fact", "patch", "situation", "world"?
8. Alignment in `dnd`: confirm BR has none (unread).
