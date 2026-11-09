# Architecture: engine, rules, adventure

DRAFT for Felix's review, revision 4: his 11 rulings, the 21 variation rulings, and a check against the reconstructions `~/Projects/{por,cotab,sotsb,pod}-re/src`, the truth from now on. Migration has started: see the checklist at the end. Settled parts: ONE-PATH.md "Next: three components".

## Dependencies

```
 gobox/core        Rules, Adventure, neutral types and id types, RNG; imports nothing of ours
 core/dnd  core/xxvc   the concepts of one world: races, classes, spells, monster and item rosters
     |          |
 rules1e    rulesxxvc   adventures: POR, COTAB, SotSB, POD | BR1, BR2     (leaves: core + one world)
     \________ engine (gobox): main loop, one script machine, combat and ship loops, front, saves ________/
 cmd/<game>: main wires one Rules and one Adventure.   cmd/import: originals to canonical, once.
```

Leaves know neither gobox nor each other (`imports_test.go`, list empty until a leaf exists). A rule never calls back: all it needs arrives as a value. Sizes: `Rules` 8 + `CharacterRules` 13 + `CombatRules` 9; optional `ShipRules` 5; `Adventure` 11.

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

Other bases are the code image (POR 260, COTAB 148, SotSB 190, POD 332, BR 102, 122 sites) or plain memory.

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

Scripts carry typed ids only: "4 hobgoblins" is `LoadMonster dnd.Hobgoblin, 4`. `MonsterID`, `ItemID`, `ShipID`, `TreasureID` are typed enums with numbers assigned once (`SpellID` stays a string). The decompiler resolves `(set, index)` and treasure file and block at import, so no set, index or block survives; `TreasureRef` is gone. `Rules.Spawn(MonsterID, *RNG)` and `Rules.Item(ItemRef)` supply every number. `FindItem` and spell numbers map at import likewise.

**Computed indices:** the patch names a typed candidate list at the site, for example `LoadMonster pick v {0: Kobold, 1: Goblin}`. A value outside the list is a run-time error. Import fails on a variable site with no list unless its writers bound it automatically.

| Sites | POR | COTAB | SotSB | POD | BR1 | BR2 |
|---|---|---|---|---|---|---|
| LoadMonster, all | 330 | 276 | 349 | 979 | 227 | 457 |
| variable index | 24 | 8 | 21 | 69 | 9 | 12 |
| of those, auto-bounded (const, Random, And in the block) | 0 | 0 | 0 | 2 | 0 | 0 |
| Treasure constant / variable block | 74 / 3 | 58 / 5 | 90 / 1 | none found | none found | none found |

So **141 monster sites and 9 treasure sites need a hand-written list**, each from reading the script or a logging trace. Writers often sit in another block, so a whole-script analysis may lower this; not tried. POD has no block operand: it uses `TreasureOf` with the items inline (op 0x8E, 145 sites). BR is unchecked.

`MON*CHA` (records / names / names with differing stat blocks): POR 172/103/12, COTAB 81/68/9, SotSB 70/68/2, POD 122/118/2, BR1 54/49/5, BR2 63/55/6. One id per distinct block; a person names the variants. `core/dnd` and the 1e tables carry every module's NPCs.

Party strength stays a number: see "Source check".

## 5. Asset data

**Import, not on-demand decoding, of originals.** `cmd/import` (gobox; may use dax, geo, ega, script) runs once on the user's machine, turning originals into canonical files in the adventure's folder: PNG (animations: frames plus delays), WAV, script text, walls. At run time `Adventure` serves `Picture(PictureID)`, `Sound(SoundCue)`, `Map(MapID)`, `Script(AreaID)`, on demand from those files, returning decoded types only. No file name, block or format reaches the engine. An adventure may read originals itself if it brings its own decoder (a leaf may not import `gobox/dax`).

**Rule tables likewise** (ruling 6). One generator, `exe/cmd/tablegen`, already POD, COTAB, SotSB (`tables.json`, `TestTablesMatchEXE`), extended to typed values in the ruleset: thac0, saves, exp, monsters, items, ships. Dropped from run time: BR2 `dsimg.go` `os.ReadFile(START.EXE)`, BR1's `dsImg` bytes, `tab`, `fightData()`, cotab `engine.go`'s `gobox/exe` import, MON*/ITEMS at run time, words (BR's `word()`). Generated Go is checked in; fine for private use, a question if ever published.

## 6. Saves: one-way

The importer reads an original save and CHRDATA/.GUY once and writes a canonical save. Nothing writes originals. `Rules.Save` is characters, the engine's part is script memory, board, RNG, and bindings carry the rest. POR's byte-for-byte save stops being a test.

## 7. One RNG

`core.RNG{State}`: **the engine owns it**, seeded at start and saved; every rolling Rules method takes `*RNG` as a parameter, so no callback and no hidden stream. Script RANDOM draws the same. Sequence: Turbo Pascal LCG. Testing: fixed seed, full reproduction.

Trace interaction: Original traces hold game by game while draw order is unchanged, including monster HP (now Spawn) and scripts; see "Traces and rule changes". BR1, BR2 use PCG (`e.rng`), so they get new sequences once: re-record at that step, separately. Anything else drawing outside: `por dasp.go:166` voice picks from global `rand`. See questions.

## 8. Ship fights

`core/ship.go`: data (`Ship`, `Gauge`, `ShipFight`, `ShipEncounter`), actions (fire, boost, jury-rig, ram, dodge, close, retreat, reload), `ShipRules` (5 methods). Adventure: script's SpaceFight names `ShipEncounter`. Rules: all of spacefight.go's arithmetic. Engine: loop, input, `game.SpaceFight` view; stage constants `Ss*` become its loop. Crew turns are ordinary `Allowed`/`Resolve` with `ShipFight` as situation. `Rules.Ships()` is nil for a ruleset without `FeatureSpaceCombat`, which then cannot load such an adventure.

`Board` thickened: `Height`, `Ground`, `ZeroG`, `Free`.

## 9. One 1e ruleset: Felix's rulings

"All four" means the Pool games. BR is a separate world with its own ruleset; the last column says how each ruling touches it.

| # | Ruling | Core and notes | BR (xxvc) |
|---|---|---|---|
| 1 | Saves: signed ints, explicit clamping at the right stage, natural-roll rules and special effects kept. A bug fix | rules-internal | same principle, 8 saves |
| 2 | Initiative: each round roll all eligible, modifiers, deterministic tie-break, sort, execute in order with delayed spells | `BeginRound`, `NextActor`. POD's coin toss on ties and POR's d100 per pick go | has its own d10 + dex + tactics; not changed by this |
| 3, 5, 6 | Difficulty cut: normal values, standard attack values | removed `WordDifficulty`. Script `PartyHit`, `MonsterHit` words are not difficulty and stay (see "Source check") | none |
| 4 | Critical hits: option, off; on = POD's rule everywhere | `Options.CriticalHits` | not applicable unless asked |
| 7 | Exp display is presentation | removed `WordHideExp`. **Conflict:** POD aftercombat.go:143 and SotSB :124 also skip the *award* when set. **Corrected:** the scripts were read and the award skip is honored, `WordNoExpAward` (see "Source check") | none |
| 8 | Cleric 10% at Wis >= 16, `xp*110/100`. Other classes separate | Verified: for integer xp `xp*110/100` equals today's `g+g/10`. POD also has the off-by-one (`clericPrimeScore` 14, aftercombat.go:39), not only SotSB. Multi-class: `exp/2` or `exp/3` truncated, and **no bonus in any game**, so no order question arises today; the remainder is lost. Paladin needs Str+Wis, ranger Str+Int+Wis, all > 15 | own XP rules |
| 9 | XP for a level, campaign cap, race+class cap are separate. Progression and race caps in rules, cap from adventure. Cap stops training, keeps XP | `Campaign.MaxLevel` | same split |
| 10 | Traps as data; no universal stoned skip | `TrapDefinition`, `SpringTrap` | same type, if BR traps fit |
| 11 | Adventure gives buy and sell per shop; copper internally; gold display is presentation | `Copper`, `ShopItem`, `ServicePrice`; `Purse` map, `ShopTerms`, `WordPriceFactor` gone. Scripts set the factor word at run time (cotab word 877), so `Adventure.ShopPrice` answers when asked, static or computed; no import bound. Gems and jewelry stay items | credits convert the same way |
| 12 | One encumbrance from the default rules; option "Party inventory", off: none, one shared slotted inventory | `Options.PartyInventory`, `Burden`, `Rules.Shared`. Note: RULE2 (c) had cut encumbrance; this reverses it, and its code is still in COTAB and SotSB | own |
| 13 | armourMove: compare | **All four identical**: weight <= 150 base move, <= 399 move 9, else 6; any plus adds 3 when move <= 9. POR 0x96, 0x18F are the same numbers (por charcalc.go:350, cotab :185, sotsb :230, pod :240). Unify | own |
| 14 | Charisma: one lookup on `max(3, min(cha, maxCharisma))`; score apart from percent | rules-internal; `FactReaction` | none |
| 15 | Spell registry, stable ids, availability, sources, variants; presence is not learnable | `Spell`, `SpellSource`, `SpellEffect`, `Campaign.Spells` | effects could use the same type |
| 16 | Slots from class and level tables plus cleric Wisdom; known, memorized, remaining apart | `Known`, `MemorizedSpell`. `MemoFact{Slot}` remains only as a script's view | no slots |
| 17 | Dual class everywhere; adventure restricts where it starts | `Training.DualClass` | none |
| 18 | Doors: adventure defines state and outcome; engine offers Pick, Bash, casting; Knock is an ordinary spell | `Door`, `DoorState`; `KnockDoor` removed | same door type |
| 19 | Canonical shared tables, documented exceptions | report below: baseline COTAB/SotSB, POD as named exceptions | own tables |
| 20 | Common attack interface replaces the hook | `AttackRequest`, `AttackResult`, `AttackResolver` | may implement it |
| 21 | Dancing a shared condition | `dnd.CondDancing`. Verified POD only: AC[0], AC[1] -4 (affects.go:1606); every save fails, die still rolled (rules.go:461); from Otto's Irresistible Dance (spells2.go:1409) on a hit. Other games have no such spell | none |

### Table comparison (ruling 19, read-only)

Values taken from each game's `rules/tables.go`, normalized to class, level. Comparison script: `/private/tmp/claude-501/-Users-felix-Projects-por-go/87588ad1-e12f-483f-a8a6-62667562088c/scratchpad/tb/cmp2.py`. Longer tables in later games are **added levels**: POR stops at 10, COTAB 12, SotSB 18, POD 21; POR rows past their cap hold unrelated bytes and are ignored.

**No conflict:** thief base by level (POR-COTAB to 9, others to 12), hit die sides and max hit dice, Con HP 3..18, cleric and mage slots through 9 where both exist, cleric and mage XP, turn undead POR=COTAB=SotSB.

**Conflicts, stored value (40 = THAC0 20, higher is better):**

| Table | Values | Suggested canonical |
|---|---|---|
| Fighter THAC0, levels 2..: paladin, ranger same | POR L2 41, COTAB, SotSB 40 and +1 each level after; POD 40, 42, 42, 44, 44 (two-level steps, 42 at L3 as others, L4 42 not 43), flat 56 from L18 | COTAB-SotSB; **POD is a rule change, decide** |
| Fighter level 0 | 39 COTAB, SotSB, POD; 40 POR | unused row probably |
| Mage THAC0 | L1-5: 40 POR, POD; 39 COTAB, SotSB. L11-15: 43 COTAB, SotSB; 44 POD. L16+: 46 SotSB; 47 POD | decide |
| Paladin saves | POR table higher by 2 (L1 14,15,16,17,17; later 12,13,14,15,15), but POR's `paladinSaves` overrides with COTAB's | later games. L8 9,9,10,10,11 worse than L7 8,9,10,10,11 in all three: data oddity |
| Paladin XP from L11 | COTAB, SotSB 1050001, 1400001, 1750001, 2100001, 2460001; POD row 1050000 (gdconsts.pas:2417), then +350000 (`xpIncrement`) | original data. Fix: ..001. L15 2460001 (sotsb gdconsts.pas:1120) breaks the +350000 step, SSI typo likely: 2450001, **flag** |
| Ranger XP from L11 | 650001 .. 1950001; POD row 650000 (gdconsts.pas:2449), then +325000 | original. Fix: ..001 |
| POD druid XP | row all -1 (gdconsts.pas:2353); `XpForLevel` (classcal.pas:522) gives -1 + (L-11)*750000 past 11 | original bug, past 11 only: no progression |
| Thief Dex 10, PP | -19 in all four sources; -10 only in the POR port, a deliberate fix (por-go rules/tables.go:931) | original SSI bug. Fix: -10 |
| Thief Dex 16, Open Locks | -5 in all four sources; +5 only in the POR port (f314dce) | original SSI bug. Fix: +5 |
| First hit dice, cleric | 2 SotSB (`firstLevelDiceTbl`, gdconsts.pas:203, used chargen.pas:1767); 1 COTAB, POD | original. Likely SotSB bug, and the base games disagree: 1, **flag** |
| Class minimum, paladin | POR Str 13, Wis 12; later Str 12, Wis 13 | later |
| Class minimum, ranger | POR Str 0; later 13 | later |
| Class minimum, Cha | POD 6 for every class but paladin (17); others 0 | decide |
| Alignments | POR cleric and four multi-classes (rows 0, 8, 9, 11, 12): 8 without True Neutral; later 9 | later |
| Slots | POR has no row above 6 (zero in table): added levels, no conflict | later |

**Spells, 137 names, 91 overlap, 46 in one game.** Correction: revision 3 compared by name, but names repeat within a table, so its level differences for Animate Dead, Bestow and Remove Curse, Cure Light Wounds and Neutralize Poison were artefacts: those games have extra rows, listed first. Rows now align by name and level; 32 names differ. Byte fields; field 0 (class group, renumbered) and 13 are not compared as rules. POR/COTAB/SotSB/POD, `-` absent, `|` several rows at that level; field meanings besides f1 level unconfirmed (`TSpellRec` in the sources could name them):

| Spell | Fields |
|---|---|
| Rows at other levels | Animate Dead L3 POR only, L5 and L7 COTAB only; Bestow Curse, Remove Curse L4 not in POR; Cure Light Wounds L2 SotSB, POD only; Neutralize Poison L3 POD only |
| Barkskin L2 | f4 -/-/0/4, f5 -/-/0/1, f7 -/-/3/2 |
| Bestow Curse L4 | f2 -/0/0/3, f5 -/10/10/3, f10 -/0/36/36 |
| Bless, Haste, Invisibility 10' | f7 4/4/3/3 |
| Blink | f7 1/0/0/0 |
| Charm Person | f4, f5 0/0/60/60 |
| Cure Disease | f12 100/100/100/100\|1 |
| Detect Magic L1 | f2 3\|0, then 3\|3\|0; f4 10\|0, then 12\|10\|0; f5 0\|2, then 0\|0\|2; f12 1\|1, then 3\|1\|1: POR two rows, later games three |
| Dimension Door | f3 -/3/12/12 |
| Dispel Evil, Entangle, Faerie Fire, Fear | f10 -/145/32/32, -/136/0/0, -/7/71/71, -/142/111/111 |
| Fireball, Lightning Bolt, Slow | f3 1/1/4/4 |
| Fumble, Ice Storm | f3 -/1/4/4 |
| Globe of Invulnerability | f6 -/-/4/0 |
| Heal | f2 -/-/5/0 |
| Hold Monsters | f3 -/1/2/2 |
| Minor Globe | f3 -/1/16/16 |
| Prayer | f7 1/4/3/3 |
| Reduce | f10 13/13/12/12 |
| Restoration | f6 4/4/-/0, f11 2/2/-/0, f12 6/6/-/1 |
| Silence 15' | f8 3/3/1/1 |
| Slow Poison | f11 2/2/0/0 |

**Not compared:** race ability limits and race-class lists (orders differ, SotSB holds two limit tables, POR another layout), ages, COTAB's combined thief race and dex table, paladin, ranger and druid slots, POD turn undead, item and monster tables. Checked in the sources as original data: THAC0 (por gdconsts.pas:1762, cotab :1356, pod :2120), class minimums (por :1872, pod :2278), and the rows marked. Paladin saves, alignment and spell rows still come from the ports' tables.

**Felix's decision:** base COTAB/SotSB, POD only as named exceptions, likely bugs checked against sources and fixed. THAC0 and spell differences are original data, not extraction errors.

### Traces and rule changes

Felix's call: traces match the originals game by game during migration, recorded traces after. Rulings 1, 2, 3/5/6, 8, 9, 10, 11, 14, 15, 16, 19 alter play or draw order, so they conflict with original traces.

Handling: **port first with original behaviour and traces green; bug fixes and unifications then land after each game's port as separate commits** that re-record the traces they affect and name the ruling. During the port a game's old behaviour lives in temporary per-game values marked for deletion, never as lasting flags; only the two `Options` stay. BR re-records once at its RNG switch.

## Scripts' data and today's code

Survey of direct character use, per game (reads through bindings):
POR 21 words; COTAB 18; SotSB 14; POD 31; BR1 11; BR2 8. Conditions queried: POR 19; COTAB 20; SotSB 3; POD 25, 56, 72, 126 and a word. Which is why `Fact` and `Condition` are small.

Where code goes: rules leaves take combat rules, affects, spells, creation, training, tables. Engine takes combat flow, map, VM, ops, presentation. Adventure takes areas, journal, lore, scripts and patches, assets. Importers take savefile, tablegen, decoders. `ecl-br` splits by section 2.

Hard spots left: combat AI one long routine (por `computerTurn`); `game.Data` mixes module, series look and rules; narration words (rules produce text: table at import); clock owned by the engine with `Elapse`; `WorldWord` list unsurveyed in part.

## Migration order

Each step keeps all six games on traces, except the named BR RNG step; rule changes follow as separate commits. 1. Settle this. 2. Machine merge (section 2), logging computed window hits. 3. Bindings and patches, POR first, then errors on. 4. Importers and asset ids. 5. POR adventure and rules leaf. 6. Engine loop. 7. SotSB, COTAB, POD on 1e. Then the rule changes and canonical tables, one commit each. 8. XXVc, BR's RNG, ship fights.

Unverified: nothing here has run on a game. Only core's build, vet, tests. No ruleset or adventure implements these interfaces yet; expect change.

## POR migration checklist

Goal: POR runs as engine (gobox) + `~/Projects/rules1e` + `por-go/adventure`, the leaves importing only core, behind `Rules` and `Adventure`, with original port behaviour kept until the port is done. How: move piece by piece; every commit builds, vets and passes `TestPortTrace` (por-go `rules/testdata/porttrace.txt`, this port's stand-in for the retired traces: RNG state and party values per fight step, Quick Start, derived values by class, level, race, dex). Core's interfaces are a draft this port corrects; changes touching a ruling are reported. BR's interpreter is not merged here. Changes: CHANGES-FROM-ORIGINAL.md.

- [x] Baseline: rulings and table base stand; options only Felix's additions
- [x] Port trace pinned
- [x] rules1e module, leaf guard on, raw-byte guard
- [x] Advancement tables in rules1e: XP, hit dice, saves, thief skills, con, slots
- [ ] Creation tables: race limits and classes, class minimums, alignments, ages
- [ ] Combat rule tables and calculations: to-hit, turn undead, affects
- [ ] Spells, items, training behind `CharacterRules`
- [ ] `por-go/adventure` leaf: areas, journal, lore, script patches, assets via importers
- [ ] Scripts through bindings (sections 1, 3), errors on
- [ ] Engine loop in gobox: main loop, combat flow, presentation
- [ ] Rule fixes, one commit each, re-recorded

## Felix's later answers

- Presentation has a second, engine-owned RNG. Gameplay keeps one stream.
- Module NPCs are in the ruleset as special/unique types; their names come from the adventure.
- A glossary fixes "character", "fact", "patch", "situation" and "world".
- Alignment is in `dnd` only: BR has none (its creation sheet shows race, sex and career).
- Table base as above. Gems and jewelry stay items, converted to copper once appraised and sold; `Coin` constants for them removed. Multi-class truncated split stays. Shop prices: the adventure answers when asked. Traces first, then rule changes as separate commits.

## Source check

These supersede earlier rows where they differ.

**Party strength is a number, so `Tier` and `PartyAtLeast` are cut.** `EclPartyStrength` (por eclvm.pas:703, cotab :716) sums over the party: (4 x cleric level + hp + 5 x AC above stored 60 + 5 x THAC0 above stored 39 + 8 x mage level) div 10, in a Byte, which a source note says may overflow. COTAB adds earlier-class levels. Only POR (18 sites) and COTAB (2) use it:

| Use | Sites |
|---|---|
| Encounter size, by arithmetic | POR area 8:642, 10:83, 10:532, 10:852, 17:113, 18:294, 20:69, 20:86, 20:577, 24:664, 25:599, 26:622, 27:588, 29:566; COTAB 64 (/16), 66 (/12) |
| Which monsters | POR 0:398 (> 24, > 50) |
| Story thresholds | POR 5:419 (with reaction, < 85), 8:382 (< 19, > 36), 18:162 (against Random 50) |

Proposal: `Check PartyStrength` answers `Value`; scripts keep their arithmetic and thresholds, no patches. The scale belongs to the `dnd` world: a `dnd` ruleset answers on this scale. Vanilla formula; the Byte wrap is a bug to fix by widening, which changes encounter sizes only where it wrapped. Unverified: how often it wraps.

**No-XP word: honored, a rule input.** `zone0^.noXpAward` (sotsb postcbt.pas:160, message :336, cleared :705); same offset named `mirrorDuel` in pod (postcbt.pas:153, :308, :594), a name the two scripts below do not support. Direct writers, one each: SotSB area 16 (platinum traded for gems: word set, `ClearMonsters`, `Treasure`, `Combat`, cleared) and POD area 66 (armour stripped off a skeleton: set, `TreasureOf`, `Combat`, cleared). Intent: `Combat` with no monsters as a loot screen, whose loot must give no XP and no award line. So `WordNoExpAward` and `FightTerms.NoExpAward`. Computed writes unverified.

**`PartyHit`, `MonsterHit`: keep.** Scripted modifiers for a fight, independent of difficulty (pod's is separate, fxengine.pas:691-699), reset after combat. COTAB 49, 51, 53 ("gods will not favor you": party -6, monsters +5), 64, 66, 67 (boon or curse 2 or -2, carried), 82 (party +8, monsters -6); POR 3, 13, 16, 17, 25, 26; POD 35, 39.

## Baseline and options

Felix on the rulings and tables: "These are unifications, and we'll treat them the same way as bugfixes: They are allowed and become our baseline." So the rulings and the table base stand. `Options` hold only his additions: critical hits and party inventory. Bug fixes and unifications land as separate commits after each game's port, with re-recorded traces.

## Remaining questions

1. Fix or keep: SotSB first cleric die 2, paladin L15 2460001, party strength wrap.
2. Rule tables generated from EXEs checked in: private only? Adventures with own decoders: none, importer only?
