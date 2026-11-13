# Architecture: engine, rules, adventure

DRAFT for Felix's review, revision 4: his 11 rulings, the 21 variation rulings, and a check against the reconstructions `~/Projects/{por,cotab,sotsb,pod}-re/src`, the truth from now on. Migration has started: see the checklist at the end. Settled parts: ONE-PATH.md "Next: three components".

## Dependencies

```
 gobox/core        Rules, Adventure, neutral types and id types, RNG; imports nothing of ours
 rules1e     core/xxvc   the concepts of one world: races, classes, spells, monster and item rosters
     |          |
 rules1e    rulesxxvc   adventures: POR, COTAB, SotSB, POD | BR1, BR2     (leaves: core + one world)
     \________ engine (gobox): main loop, one script machine, combat and ship loops, front, saves ________/
 cmd/gobox: the one binary; starts any adventure from its data folder.   gobox import: originals to canonical, once.
```

**One binary, adventures are data (Felix 2026-11-11).** There is no per-game executable and no per-game repo at the end: `por-go`, `cotab-go`, `sotsb-go`, `pod-go`, `buckrogers*-go` are wound down. One `gobox` binary lists the adventure folders under its data directory (`adventures/<id>/`) and plays any of them. An adventure folder holds only imported, canonical data: a manifest (id, title, ruleset, start sequence, declared custom fields), scripts, the PNG atlases, map `.bin`+`.txt`, JSON records and text, sound as MIDI and JSON synth data. The binary needs nothing else at run time: no original file, no POOL.CFG, no check that one exists; a folder without a valid manifest is not listed. Consequences:
- The adventure "leaf" is no Go code per game: `core.Adventure` has one implementation in gobox that reads an adventure folder. What a game used to do in Go becomes data (scripts, declared fields, the start/end sequences) or a shared engine feature (wilderness travel, the pager, WhoPays, Quest/Journal).
- Rulesets (`rules1e`, `rulesxxvc`, `rules4e`) are compiled into the binary; the manifest names which one, `Mismatches(Needs, Catalog)` checks it at load.
- Everything per game lives on the import path only: `gobox import <game> <originals>` runs that game's decompiler, its commented special cases (the former script patches and journal/quest sites), the DAX/GEO/MON readers and the tablegen, and writes the folder. Importers may know their game; the run-time binary may not.
- Each game's tests move with it: import-equality tests (art, data, sound, scripts) into the importer, rule vectors (TestAgainstPascal) into the ruleset.

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

Unverified: this is static, listing each index's writers. It does not bound them all and I did not run the games. The real proof is the machine's error: run it first on the current VM as a logging run through each game's script playthrough (cheap, targeted), and bind what appears.

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

Plan: (1) widen `gobox/ecl` to int32 and width, all four Pool games unchanged; (2) add dialect tables for V4 and FOR/NEXT; (3) move BR's ops over; (4) delete `ecl-br`'s VM, leaving ship, record, tables for `rulesxxvc`. POR's overrides `Compare` and `GetTable` go by patches (below), `Exit` empty by patching its scripts, `NoRemainder` by Layout. No Quirk list.

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

So **141 monster sites and 9 treasure sites need a hand-written list**, each from reading the script or a logging run. Writers often sit in another block, so a whole-script analysis may lower this; not tried. POD has no block operand: it uses `TreasureOf` with the items inline (op 0x8E, 145 sites). BR is unchecked.

`MON*CHA` (records / names / names with differing stat blocks): POR 172/103/12, COTAB 81/68/9, SotSB 70/68/2, POD 122/118/2, BR1 54/49/5, BR2 63/55/6. One id per distinct block; a person names the variants. `rules1e` (dnd*.go) and the 1e tables carry every module's NPCs.

Party strength stays a number: see "Source check".

## 5. Asset data

**Import, not on-demand decoding, of originals.** `gobox import` (formerly `cmd/import`; may use dax, geo, ega, script) runs once on the user's machine, turning originals into canonical files in the adventure's folder: PNG (animations: frames plus delays), WAV, script text, walls. At run time `Adventure` serves `Picture(PictureID)`, `Sound(SoundCue)`, `Map(MapID)`, `Script(AreaID)`, on demand from those files, returning decoded types only. No file name, block or format reaches the engine. An adventure may read originals itself if it brings its own decoder (a leaf may not import `gobox/dax`).

**Rule tables likewise** (ruling 6). One generator, `exe/cmd/tablegen`, already POD, COTAB, SotSB (`tables.json`, `TestTablesMatchEXE`), extended to typed values in the ruleset: thac0, saves, exp, monsters, items, ships. Dropped from run time: BR2 `dsimg.go` `os.ReadFile(START.EXE)`, BR1's `dsImg` bytes, `tab`, `fightData()`, cotab `engine.go`'s `gobox/exe` import, MON*/ITEMS at run time, words (BR's `word()`). Generated Go is checked in; fine for private use, a question if ever published.

## 6. Saves: one-way

The importer reads an original save and CHRDATA/.GUY once and writes a canonical save. Nothing writes originals. `Rules.Save` is characters, the engine's part is script memory, board, RNG, and bindings carry the rest. POR's byte-for-byte save stops being a test.

## 7. One RNG

`core.RNG{State}`: **the engine owns it**, seeded at start and saved; every rolling Rules method takes `*RNG` as a parameter, so no callback and no hidden stream. Script RANDOM draws the same. Sequence: Turbo Pascal LCG. Testing: fixed seed, full reproduction.

What is left of RNG compatibility: the differential harness seeds Turbo Pascal's `System.RandSeed` and `core.RNG` alike, so a rule compared against the pure Pascal must draw in the Pascal's order and count, unless CHANGES-FROM-ORIGINAL.md lists the difference. Nothing else binds the sequence: no recorded run exists. BR1, BR2 use PCG (`e.rng`); moving them to `core.RNG` is needed only if they get a pure reference. Anything else drawing outside: `por dasp.go:166` voice picks from global `rand`. See questions.

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

### The reference and allowed differences

Felix: "We'll only ever compare rules and adventure behaviour against the originals. Both need to come from the original." **The reference is the pure Pascal** (`~/Projects/por-re` branch `bugfixed-rules`, `src/pure`). A port difference from it is a bug, or a documented entry in CHANGES-FROM-ORIGINAL.md, which is the list of allowed differences. A Pascal driver writes vectors from pure functions over generated inputs; the Go side runs the leaf on the same inputs and compares (`rules1e` `TestAgainstPascal`).

The reference includes the 18 fixes of `4a28cb6` (fix 17 corrected on `bugfixed-rules`) (por-re `BUGS.md`): `bugfixed-rules` contains that commit, so it is SSI's rules plus those fixes, not the shipped bytes.

The engine owns the combat loop and the script interpreter (Felix's rulings), so `CombatMain`, `EndOfRound` and `eclvm` split at that line, and those vectors compare engine plus leaf.

Vectors are stored as test data with the por-re commit in their header, regenerated by `vectors-pure.sh`: tests then need no DOSBox-X and any change of a vector shows in review. COTAB, SotSB, POD and BR have no pure branch, so those games have no reference to compare against until theirs exist.

Bug fixes and unifications land after each game's port as separate commits naming the ruling and their CHANGES-FROM-ORIGINAL.md row. During the port a game's old behaviour lives in temporary per-game values marked for deletion, never as lasting flags; only the two `Options` stay.

## Scripts' data and today's code

Survey of direct character use, per game (reads through bindings):
POR 21 words; COTAB 18; SotSB 14; POD 31; BR1 11; BR2 8. Conditions queried: POR 19; COTAB 20; SotSB 3; POD 25, 56, 72, 126 and a word. Which is why `Fact` and `Condition` are small.

Where code goes: rules leaves take combat rules, affects, spells, creation, training, tables. Engine takes combat flow, map, VM, ops, presentation. Adventure takes areas, journal, lore, scripts and patches, assets. Importers take savefile, tablegen, decoders. `ecl-br` splits by section 2.

Hard spots left: combat AI one long routine (por `computerTurn`); `game.Data` mixes module, series look and rules; narration words (rules produce text: table at import); clock owned by the engine with `Elapse`; `WorldWord` list unsurveyed in part.

## Migration order

Each step compares against the pure reference where one exists; rule changes follow as separate commits. 1. Settle this. 2. Machine merge (section 2), logging computed window hits. 3. Bindings and patches, POR first, then errors on. 4. Importers and asset ids. 5. POR adventure and rules leaf. 6. Engine loop. 7. SotSB, COTAB, POD on 1e. Then the rule changes and canonical tables, one commit each. 8. XXVc, BR's RNG, ship fights.

Unverified: nothing here has run on a game. Only core's build, vet, tests. No ruleset or adventure implements these interfaces yet; expect change.

## POR migration checklist

Goal: POR runs as engine (gobox) + `~/Projects/rules1e` + its imported folder `adventures/por/` (see "One binary, adventures are data"), the ruleset importing only core, behind `Rules` and `Adventure`, compared against the pure Pascal. How: move piece by piece; every commit builds, vets and passes `TestAgainstPascal`, which grows with each piece. Core's interfaces are a draft this port corrects; changes touching a ruling are reported. BR's interpreter is not merged here. Changes: CHANGES-FROM-ORIGINAL.md.

- [x] Baseline: rulings and table base stand; options only Felix's additions
- [x] Differential harness: `vecadv`, `vecstat`, `vectab`, `vecdice`, `veccmb`, `vectors-pure.sh`, `TestAgainstPascal`, 8307 cases
- [x] Class stats with Wisdom bonus in rules1e (`ClassStats`); seeded dice agree: `core.RNG` is Turbo Pascal's generator (605 cases)
- [x] rules1e module, leaf guard on, raw-byte guard
- [x] Advancement tables in rules1e: XP, hit dice, saves, thief skills, con, slots
- [x] Creation tables as data. Uncompared logic, not in `src/pure`: chargen, shop, temple, camp
- [x] Combat rolls: attack, trap, save, strength adjustments, turn table, attack rate, seeded
- [x] One strength and attack rate: por-go on rules1e. grules' copies stay for cotab, pod, sotsb until they migrate
- [x] Table reads past their end: `vecrange` (levels -1..14), `vecturn` (TurnColumn through CombatTurnUndead), records checked by `TestMonsterRecordsStayInTables`
- [x] Gate: `gate.sh`, staged trees exported clean; other games when gobox-engine outside core changes
- [x] Effects on rolls: `veceff`, `TestEffectsAgainstPascal` in por-go while the handlers live there
- [x] Weapon and plain damage, movement with effects: `vecdmg`, same test. Backstab multiple not yet compared
- [x] Whole computer turn, melee only: `vecai`, `TestDecideAgainstPascal`; backstab multiple compared
- [x] Whole computer turn, morale, flight, walls: `aimorale`, `aiwall` in `vecai`
- [x] Whole computer turn, spells and items: `aicast` in `vecai`; item cast and combat aim are gaps of the reference
- [ ] Combat: `Decide` through the seam
- [x] src/pure additions: temple (`Temple`, `vectmpl`, TestTempleAgainstPascal)
- [ ] src/pure additions: shop, camp, chargen, item spell casting, combat aiming
- [ ] Spells, items, training behind `CharacterRules`
- [ ] adventure folder `adventures/por/` read by gobox: areas, journal, lore, scripts with fixes applied at import, assets
- [ ] Scripts through bindings (sections 1, 3), errors on
- [ ] Engine loop in gobox: main loop, combat flow, presentation
- [ ] Rule fixes, one commit each, listed in CHANGES-FROM-ORIGINAL.md

### POR onto the architecture (handover, Felix 2026-11-10: "make POR adhere to our architecture first")

Goal (revised 2026-11-11, "One binary, adventures are data"): POR plays in the one `gobox` binary from `adventures/por/`, a folder `gobox import por` writes; `rules1e` implements `core.Rules` and imports only `gobox/core` and `rules1e` (dnd*.go); gobox's one folder reader implements `core.Adventure`; `por-go/rules` and `cmd/por` are gone, POR's decompiler fixes and readers live on gobox's import path. How: two agents in parallel, each step gated and committed, behaviour equal to the pure Pascal where vectors exist (any difference a CHANGES row). Lock: stage only your own paths, and hold `/private/tmp/por-gate.lock` (mkdir; rmdir after) from staging through gate and commit. Core interface changes are reported to Felix, not decided silently. Other games' 1b-e and the BR layouts wait until this is done.

**Felix's order (2026-11-10):** "Separate the drawing from the rest first. We need pure functions, like the pascal code. Make sure we only have rules in the engine not in POR. Then do one big refactor and write everything as it should be and then fix the bugs." `Decide` + `Resolve` stay (core). An adventure is data only, read by one small shared piece of code that answers the engine. State the games share becomes a typed engine state variable. A game's own variables become an engine feature (the adventure declares them by name and type, and scripts read and write them). What cannot trivially stay data is to be discussed with Felix before it is built.

**Felix's decisions on state, routines and flow (2026-11-11):**
- **Shared state:** the state the games share (position, facing, clock, previous position, leaving the map, rest interrupt, surprise, morale, search, sky, temple, shop, party modifiers, combat result, ...) becomes ordinary typed Go variables in the engine's shared state, one name each.
- **Per-game state** becomes custom fields, an engine feature: the adventure declares each by name and type, and scripts reach it by name. Access is validated: a name must refer to a value declared before it is used, or the import fails.
- **Engine routines:** every routine a script calls (CALL routines, native scripts, per-dialect commands) is part of one shared collection of engine functions callable from scripts.
- **Side effects of writes:** every side effect of writing a variable becomes an explicit action; a write only stores.
- **Game flow:** the hard-coded flows of the four games are replaced by one shared flow that every game uses. Differences are analysed first, and the places where one flow is problematic are reported to Felix.
- **One function vocabulary, no dialects at run time:** there is one list of identifiers for the shared script functions. Each game's importer or decompiler rewrites its dialect (opcode numbers, CALL routine numbers, operand encodings) onto that list, so the engine knows no dialect. Where one name behaves differently per game, that is either unified by a ruling or becomes two shared functions.
- **Side effects rewritten at import:** the decompiler reads each game's original write side effects from its Pascal (ReadGbvm/WriteGbvm and their specials) into a per-game table of (field, value condition → action). It rewrites every matching write into a plain write plus that action. A write whose target or value is only known at run time (computed table writes, a variable value) gets the guarded form (`if v == 0 RemoveMember`) or fails the import, never a silent write. No hand patching of scripts for side effects.
- **Mapping meanings:** only fields the original engine itself reads or writes need mapping to shared state. Those are found from the engine procedures that use them in each game's Pascal and matched across games by that use; a person confirms each match. Script-only fields need no mapping: they become custom fields as they are.
- **Wilderness travel** is in all six games (BuildWildernessArena; MovePartyWilderness in POR, POD, BR1, BR2; COTAB's mapcurs): one shared engine feature.
- **Start and end sequences** (title, intro, import offer, ending) are declarative script sequences: timings, images, page layouts, animation hints.
- **References:** every game has a Pascal reconstruction (`~/Projects/*-re`); only POR has a pure branch so far. Each game's quirks dropped by the shared flow are judged against its own reconstruction, and a pure branch per game is the way to compare behaviour.
- **Decompiler and DAX at import only (Felix, 2026-11-11):** nothing at run time decompiles or reads DAX. `cmd/import` runs the decompiler once and writes the finished scripts into the adventure folder; `Library` only reads text, and a missing or stale script is an import error at load.
- **The patch files go:** each fix becomes a hard-coded special case in that game's decompiler, commented with what the original did, why it changes and its CHANGES row, so the decompiler writes the patched scripts directly. `patches/*.patch` and `ApplyPatch` are deleted. A special case whose site no longer matches the code it expects fails the import, never silently skips.

**Felix's decisions for A3 and A5 (2026-11-11):**
- **Long text:** the scripts' text commands stay. A pager command shows long passages (journal entries, the Phlan histories, tavern tales) from the adventure's text data; one exists or gets added, one for every game.
- **Script cells are 64-bit:** this supersedes section 2's int32 widening. Gold and every other value fit, and the purse is read as an ordinary value.
- **Shared Purse, a gameplay option:**
  - On: one party purse; a "who pays" prompt is skipped and the purse pays.
  - Off (original mode): each member holds their own coins; the engine asks who pays, as the original did.
  - So a script asks for payment through a dedicated command (or at least marks the prompt as "who pays"), never by plain text.
  - This also answers open question 2 (coin weight): with the option off, coins and their weight are per member again. `CharacterRules.Purse()` (one `Copper`) must serve both modes: a core change, reported when made.
- **Formats** (canonical, rule 9):
  - every graphic, animations included, in a PNG pixel atlas;
  - sound as MIDI data or Ogg Vorbis samples;
  - map data as a very simple binary format plus a metadata text file;
  - JSON for structured data meant to be readable as plain text (items, text entries).

- [ ] S. Separation in place (agent R): every drawing, sound, message and wait in `por-go/rules` logic becomes a returned `core.Effect`. The logic becomes pure functions on values (as `src/pure`), with no `Engine` receiver where the Pascal has none. Behaviour is unchanged: the Pascal vectors stay equal. Order: combat AI turn, attacks, spells and effects, items, characters (chargen, training, shop, temple, camp). No move to rules1e yet
  - [x] combat AI turn (por-go a132998): `computerTurn` returns its `core.Outcome`, `present` applies it in order; combat text `Narrated` (core: `Narrated.More`, a line going on), `Sounded`, standing `Turned`. Still drawing inside the turn's callees: attacks, spells, items (next steps); the turn still has an `Engine` receiver (the rule state is not yet apart from presentation)
  - [ ] attacks: [x] facts `Attacked`, `Struck`, `Fell`, `Walked`, `Sounded` as effects, `Delay`/`MsgDelay` gone (por-go 0138853); [x] poses, hit flashes, swing and step camera gone, a turn is `Turned` (por-go a1db765); [x] missiles: `Shot` carries what flies (`Start`, `Item`, `Spell`), `present` picks the art (`missileArt`); spell casts are `Casting`; every rule `dasp` is a `Sounded` (por-go 3f73d52)
  - [x] rule state: `RuleState` (rulestate.go), the gdvars apart from the Engine's presentation, script machine and map (por-go b56ed51)
  - [x] rule functions on it: 236 Engine methods that touch only the rule state take `*RuleState`, bodies unchanged (por-go 2728556)
  - [x] computer turn: `computerTurn` and its callees (attacks, spells, items, affects; 467 functions) on `*RuleState`; `CombatRules` and rules4e take it (por-go b4e5abe)
  - [x] spells, effects and items: spell handlers narrate and sound, `combatCentre`, `soundWait`, `combatRedrawChar` and `quietReady` gone from the rules; the player's Ready makes the wield (por-go 7182dfc)
  - [x] characters: chargen, camp (clock, study, rest), training, dual class, purse and buying on `*RuleState` (527 functions); shop Identify, temple Cure and Train stay Engine commands over them (por-go 03013f4). Still open: `Engine` commands over rule functions remain, no move to rules1e (B)
  - **Felix's decisions (2026-11-11):** (1) ids: typed spell and item ids in `rules1e` (dnd*.go) from the tables, one generator (`rules1e/idgen`, spec per game: por-go `rules/ids.txt`), checked in, numbers assigned once; `SpellID` stays a string (section 4), monster powers that fly get one too (`StoneGaze`, `CharmGaze`, `SpitAcid`, `BreathWeapon`); A4 uses the same ids. (2) rule state: one value of its own, the Go counterpart of gdvars (combatants, positions, party, effects); rule functions become plain functions taking it and `*core.RNG` and returning effects; board positions split off in B
- [ ] B. One big refactor into the target layout (after S, and after the adventure-data discussion): rules to rules1e behind `core.Rules`, the adventure as data plus the shared reader, the engine loop in gobox
  - **B's partition of `RuleState`** (por-go rulestate.go at a317cc0; the spec of step B, no code moved yet):
    - rules1e state: `RandSeed`, `party`, `pool`, `items`, `loot`, `purse`, `treasure`, `usingItem`, `usedScroll`, `rv`, `thrownCopy`, `spellTarget`, `spellTargets`, `nSpellTargets`, `spellCaster`, `curSpell`, `spellOwnTargets`, `spellValue`, `spellValue2`, `turnValue`, `modMoves`, `affectVeto`, `fleeRoll`, `slotHits`, `slotShots`, `attackedAlly`, `tiedTarget`, `foesHealth`, `magicOn`, `combatRound`, `fight`, `loadouts`, `aiLastDir`, `aiBlocked`, `aiState3`, `aiOpen`, `beginFlag`, `game`, `area`, `camp` (minus `resting.shown`), `effTick`, `training`, `learning`, `minStats`, `debug`, `fx` (the effects out)
    - combat board (split off in B): `near`, `nearList`, `combatants`, `targets`, `cmap`, `cpos`, `ncomb`, `occ`, `bodies`, `nbodies`, `bodyIndex`, `place`, `terrain`, `clouds`, `tgtX`, `tgtY`, `posX`, `posY`, `facing` (the party's map place, the board's start)
    - UI or prompt leftovers, back to the presenter: `partySel` (Felix 2026-11-11, option b: the actor is an argument at the core seam only; `CharacterRules` and `CombatRules` calls take who (`CharID`), inside the rules a call-scoped actor is set from it or from the fight's turn, as the reference does with DS:5D92; the presenter's selection only remembers whom to pass and reads the rules' actor back after the calls that move it: joining, leaving, after a fight; no two variables kept in sync), `tempSel`, `temps` (selectTemp's view of a non-member), `creation` (the creation sheet), `aiKey` (the key polled in a computer turn), `castYes`, `castOn`, `castCancelled` (castAsk's answers: a prompt's reply, passed in), `resting` (DS:6DDA, a message-border flag), `camp.resting.shown`, `healed`, `interrupted` (RestStep's drawing signals: become effects)
    - to delete: `spellFx` (DS:6B84, the original's table of spell procedure pointers, filled by spellTableInit: a switch on the typed spell id does it); `combatTargets` (DS:6B84 repointed by Combat: the fight's targeting chosen by a flag, becomes a parameter of the cast); `combatRules` as `atomic.Pointer` and `PathSearch` as `atomic.Bool` (atomics because the frontend's options goroutine writes them while the engine runs, cmd/por/main.go: become plain settings handed in at a decision point)
  - **`RuleState` methods still holding presentation state** (15; what each becomes): `healTick`, `stepRest` (set `camp.resting` signals: return `Narrated`/a rest fact), `StopRest` (a command, back to Engine), `aiKeys` (reads `aiKey`: the poll becomes an input to `computerTurn`), `fullAttack`, `AddToParty`, `selectMember`, `selected` (`partySel`: the actor is an argument), `selectTemp`, `unselectTemp` (presenter's), `castSpell`, `CastSpell`, `useItem` (`castCancelled`/`castYes`: the reply is an argument, the cancel an outcome), `spellTargetSetup` (`castOn`: an argument), `spellTargetHook` (`combatTargets`: a parameter). Found by: RuleState method bodies naming those fields (no RuleState method reaches the VM or drawing by type: RuleState holds no `*Engine`)
  - [x] to delete: `spellFx` (spellEffect switches on the spell; TestSpellEffects: one case per spell, por-go 8338eb0); `combatTargets`, `castYes`, `castOn`, `castCancelled` (a cast is given a `castReply`, `CastSpell` returns the cancel, `castAsk` returns the reply, por-go 6b63641); the atomics (`Settings` handed in by `Engine.HandIn`, taken by `Do` and `CombatStep`, por-go 7706c5e)
  - [x] combat board: `Board` its own value in `RuleState` (embedded; por-go 6b5e3fb)
  - [x] presentation leftovers out: `aiKey` (Engine's, `computerTurn(c, key)`), the rest's signals (`stepRest` returns `restStep`, `healTick` bool, `restLine`/`restShown` the Engine's, `StopRest` an Engine command), `creation`, `tempSel`, `temps` (Engine's; `setField` replaces, `SetField` swaps the selection) (por-go 2ccf3bd, 4eb1009, 0fb4bfa)
  - [x] the actor: `RuleState.who` (DS:5D92) set per call from the caller (`Engine.as`) or the fight's turn; `Engine.sel` the presenter's selection, `follow` after `AddToParty` (`join`); TestActorSeam: every Engine call into a rule reading the actor goes through `as` (por-go 7900727)
  - [ ] still open: the move into rules1e behind `core.Rules`/`CharacterRules`/`CombatRules`; the adventure leaf; the engine loop
- [ ] F. Then the bugs it shows, one commit each
  - [ ] in-fight sounds passed by cue name may be silent (front/classic resolves only numeric cues)
  - [ ] no PC/AdLib fallback for the fall sound outside fights
  - [ ] toasts replace pane lines outside fights
  - [ ] the rest-time line refreshes only on the 5-step tick (RestStep)

Track R (rules seam): superseded by S and B above, kept as the list of what B must cover.
- [ ] R1. Combat `Decide` through the seam: por-go's computer turn calls `CombatRules.Decide` of rules1e, `TestDecideAgainstPascal` equal
- [ ] R2. Combat handlers into rules1e behind `CombatRules` (BeginFight .. EndFight; spell and effect handlers); engine keeps the loop and presentation (item 7 below)
- [ ] R3. Characters behind `CharacterRules`: spells, items, training, chargen (`Draft`), shop, temple, camp, purse (`Purse() Copper`); `Fact`/`SetFact` pure; `Check`
- [ ] R4. `Rules.Allowed`/`Resolve`/`Elapse`/`Save`/`Load` complete in rules1e; por-go calls only the interface

Track A (adventure seam, agent A): scripts and data move into the adventure leaf.
- [x] A1 (por-go 5ccd712; cmd/fixture still edits original saves, which load through the import). 1b-f: saves by field (typed state by id; canonical save, section 6); old saves through saveimport.go
- [ ] A2. Section 1: the character window through `Rules.Fact`/`SetFact` bindings (newEngineFields' `sel()` reads go); side-effect writes (member gone, died, stoned) become `Action`s by patch; an unbound field is an error
- [ ] A3. Section 3: native scripts become script patches (tavern, histories, PayPlatinum)
  - [x] 64-bit cells: `ecl.Host` Get/Set and `VM.Arg` are int64; each game's host stores a word as the original (truncation at the store, so arithmetic, compare and random stay equal; no script found depending on a wrap before the store)
  - [x] pager: `Journal(entry)` (POD's command 31) in every dialect (V1 at the same unused number); entries are the adventure's `names/entries.json` (`script.Entry`: id, name, book, title, box, voiced), their names the scripts' `Entry` values
- [ ] A4. Section 4: typed ids in scripts (LoadMonster, treasure, FindItem; candidate lists at POR's 24 + 3 variable sites)
- [ ] A5. Section 5: `cmd/import` originals to canonical files; no `gobox/dax` in POR at run time
  - [x] decompiler and DAX at import only: `script.Import`/`ImportIfStale` (with `DAXSource`) are the import path; `Library` has no Source and only reads texts (missing or stale: `ErrNotImported`). POR: `cmd/import` (was cmd/eclscript), `por -setup` and the start run `rules.ImportScripts`; the other five games' commands run their `rules.ImportScripts` into the cache before `New`, their tests in TestMain
  - [x] sound (Felix: keep synth data): `audio/sounddata.Decode` (was cmd/sounddata's body) decodes START.EXE's speaker and Tandy tables at import into data/assets/data/sound.json (por `ImportData`, `soundLayout`; START.EXE now copied by import.toml); the run reads it (`Engine.SoundData`); rules/sounddata.json gone; TestSoundEqualsEXE. app.WriteCover's dax.ParseImage runs only in -setup (por, pod): import path. Still raw at run time: the Amiga SOUND.BIN (amiga.Player) and the other games' embedded sounddata.json and AdLib DAX
  - [x] patches into the decompiler: POR's 14 patch files (58 hunks; COTAB, SotSB, POD had none) are `script.Fix` special cases in por-go rules/scriptfixes.go, each with original, why and its CHANGES row (29 "POR script, ..." rows under Port's own); a site not there exactly once fails the import. Proof: every area's import equals the patched text byte for byte apart from the key header (29 areas). patches/, `ApplyPatch` and the patch part of the key are gone
  - [x] journal references: 47 special cases (31 prints in 14 areas, area 13's JournalEntry routine and its 5 callers, City Hall's proclamations) name their entries through `Journal` (JournalEntry1-58, 18 proclamations in names/entries.json); the text keeps what the text window showed. journalRefs, handoutPhrase, journalPrinted go; Journal no longer clears the text window, notes a journal text, and one script's entries share one pager. TestJournalSitesNamed detects a printed reference. Same class closed (A4): POR quests by `Quest(quest)` (gobox ecl V1 66, names/quests.txt, 30 sites; TestQuestSitesNamed), COTAB and SotSB journal references by `Journal` (journalfixes.go; TestJournalSitesNamed each); gobox `journal.Ref`/`Printed` deleted
- [ ] A6. gobox's adventure-folder reader implements `core.Adventure` (manifest, areas, journal, lore, scripts, assets); POR's importer writes `adventures/por/`, with no run-time check of originals (POOL.CFG, TITLE, 8X8D1 go)

Then (one agent, after both tracks):
- [ ] E1. Engine loop in gobox (main loop, combat flow, presentation); `cmd/gobox` lists adventure folders and starts one with the manifest's ruleset; `por-go/rules` and `cmd/por` gone
- [ ] E2. Leaf guard on for both leaves; each game's full suite once; one scripted, windowed POR playthrough

### Reference queue (handover)

Goal: the bug-fixed Pascal (`por-re` branch `bugfixed-rules`, `src/pure`) is the reference; por-go `rules` and rules1e equal it with gameplay options off. How, per item: Pascal fix (BUGS.md) → vector driver (`vectors-pure.sh`, `build-pure.sh`) → regenerate rules1e `testdata/por-*.txt` → Go `*_pascal_test.go` equal with no exception → break once, see it fail. Commit order por-re, por-go, rules1e; `gate.sh` before every commit; nothing pushed. Allowed differences are CHANGES-FROM-ORIGINAL.md rows.

- [x] 1. EclSpell (area 14 Knock; no Pascal fix: the Spell command is gone, d3): `found` uninitialised, slot 0 never scanned, byte-truncated addresses; check Diogenes' `Save 1 [6B32]` before moving the window base
- [x] 2. U1 ROB: `EclRob` reads var 3 for the per-item chance
- [ ] 3. Group C: [x] U2 option, [x] U3 option, [x] U4 kept (whenCast; every listed spell has a combat handler), [ ] U5 Pascal display fix (with viewchar, item 6), [x] U6-U8 rows, [x] U9, [x] U10 (saveimport.go)
- [ ] 1b. Typed script references (Felix d1-d4): no address ranges in gobox ecl or any game; each gated step builds and passes
  - [x] a. Layout table format in gobox ecl/script: per dialect, data; field, member-field, array (typed index, bounds), local, const-table, engine fact; unknown address fails import
  - [x] b. (data in por-go c0e080d; coverage test with c) POR layout table from por-re's Pascal record types (TCharRecord, zones, specials)
  - [x] c. Importer: every operand resolved to a typed `Ref`; computed addresses to arrays; code-space tables to const tables; text format names refs
    - [x] `ResolveRefs`, `ConstLayout`, `replaces` (an area field over a game field); por-go `TestLayoutCoversScripts` resolves every patched script; area 9 coins are `SelCoins[i]`; area 26 `RiverState` replaces CouncilMeeting (patch and CHANGES row go with d); $6BF2 stays unmapped (`SelWindowF2`)
    - [x] `Library.Area` resolves every script in its layout (`ReadLayout`); a scripts folder without layout/global.txt is the 1b-i games, made mandatory there
    - [x] text format names refs: the layout is the one name source (ReadNames; array elements Name_i), value names scoped per area (Set aN); POR's names files keep only tables and labels
  - [x] d. VM `Host` on `Ref` (Get/Set/Text/SetText); `VarAddr`, `Addr`, `RemainderAddr` go
    - [x] four games: hosts map a Ref back with `Layout.Address` (their state is still the original memory); `vm.Remainder` is the layout's field; table bases resolve whatever their type (`Ref.Plus`); a variable operand without a Ref panics; tests resolve their own scripts (`Assemble(v, layout, code)`, por `resolved`); POR's Scratch6 and Hours by `Layout.Named`
    - [x] VM `Arg` of a text (0x80) or text variable (0x81) operand is 0: checked against the four -re Pascals and every shipped script's text operands, none reaches a numeric GetEclVar
  - [ ] e. host on typed fields, zones gone, per game: [x] POR (por-go 1c766dc), [ ] then COTAB, SotSB, POD (Felix: no flat memory behind Ref). After POR's: BR1/BR2 layouts. POR: (then drop patches/area26.patch, CHANGES row: RiverState replaces CouncilMeeting); `ReadVar`/`WriteVar`, `addrKind`, zones, `SaveEditor.SetVar(addr)` go
  - [x] f. Saves: typed state by id; old saves through saveimport.go (por-go 5ccd712, A1)
  - [x] g. Area 14: Spell/SaveTable pair becomes one typed CastKnock call (native caster choice, slot spent); Spell opcode goes; cabinets stay script events
  - [x] h. Diogenes' camp write read as `RestChance = 1` (d4), CHANGES row with bytes
  - [ ] i. Layout tables for COTAB, SotSB, POD, BR1, BR2 from their -re Pascal; their builds and tests
    - [x] COTAB, SotSB, POD: rules/scripts/layout/global.txt from cotab-re/sotsb-re/pod-re (zones, specials, TZone arms, TGbvmBank0/TAreaState/TCharRecord), embedded as Library.Data; each game's TestLayoutCoversScripts resolves every DAX script; ReadNames/ReadLayout take an fs.FS
    - [ ] layout required: BR1/BR2 have none yet, so ReadLayout gives nil for them (their own VM reads addresses); the exception goes with their layouts. Done: Version.Vars and v1Vars gone (the layout is the one name source; Names.Calls names CALL routines)
    - [x] CALL routines are their own space in a Layout (ResolveCall), named from each -re's EclCall; COTAB/SotSB Word cells the -re units name by constant (Shop, Temple, PriceScale, ...) named; POD 0014 is Temple (healerRequest)
    - [x] library texts carry a key of names/, layout/ and the special cases (script.Key); another key is ErrNotImported at load (A5)
    - [ ] unnamed script-only cells (WordXXXX; no -re field or constant names them): COTAB 171, SotSB 156, POD 360; meaning to be read from the scripts
    - [ ] BR1, BR2 layouts from buckroger-re / buckrogers2-re (five zones, eclmenu.pas ReadGbvm/WriteGbvm); then drop ReadLayout's nil
    - [ ] after d: name the unnamed cells from their script use, per game (COTAB, SotSB, POD)
    - [ ] after d: `WriteGbvm($2CB, …)` in cotab-re and sotsb-re EclCombat writes outside every zone (likely a Pascal bug): investigate
    - [x] bug class (cotab 8ee0677): a layout name overwrote a name the fold depends on; detection: gate.sh runs each game's Layout/Fold/Gate tests on every gobox change, and a macro refuses a word the layout does not name
    - [x] bug class (por-go 6f9b20a): a behaviour commit left tests stale in packages the gate did not run; detection: gate.sh tests every package whose .go files the staged change touches, in every repo, from the staged tree (slow tests kept out by name in `slow`); stage in all repos, run the gate, then commit
- [x] 4. aihigh: spell ids >= 0x80 fixed in the Pascal, tolerance dropped
- [x] 5. Review `fix:` / `fixed:` comments in `rules/*.go`
- [ ] 6. src/pure additions: shop, camp, chargen (conAdj clamp), item spell casting, combat aiming, viewchar, money, postcbt, saveload (B6), manual combat menus, SpellSelection, CombatMain round (B13)
- [ ] 7. Handler move into rules1e behind CombatRules; por-re refactoring template

**por-go hashes rewritten.** The por-go history from 1682aff on was rewritten (the `rules4e/powers.go` rename folded in; trees equal). Commit messages in rules1e, por-re and gobox-engine still name the old por-go hashes; this maps them, oldest first (old → new):

- 1682aff → f371700 Character.Thac0Base: was misnamed HitDice, renamed with gopls
- 459f173 → 0b242af rules: class stats and creation tables come from rules1e
- 034af50 → e60246c rules: attack, trap and save rolls and turn undead from rules1e; rules
- fe905fc → 9ba81fc rules: strength and attack rate from rules1e; records stay in the tabl
- 1c6278d → e443f04 rules: records in the tables test over every MON file and saved charac
- 6bfbc24 → 1d4855e rules: effects on rolls against the Pascal; 4680 vectors
- 9d43963 → b28a1ef rules: damage and movement against the Pascal; 13296 vectors
- 1943e2f → 716eac6 rules: a whole melee computer turn against the Pascal; 272 vectors
- 9341db1 → ed6637d rules: morale, flight and walls in a computer turn against the Pascal;
- 0ab9c1c → 0d7c647 rules: spells and items in a computer turn against the Pascal; 5144 ve
- 0f2091c → b104868 rules: fixture carry-over fixed, name exception gone; 5530 vectors aga
- f0e9159 → 565bd1c rules: surrender and team count exceptions gone with the reference fix
- 534ebb7 → a59c626 rules: computer turn offers the item before a readied spell, as the or
- f41f032 → 91a363d rules: aibreath cases against the Pascal (breath weapon aim and bolt);
- 4f9802b → 1aa4542 rules: item cases on quick fight against the Pascal, item exception go
- 7187093 → 3972bbf decomp: disassembly listings, their tools and the ghidra skill gone; t
- e166e26 → 85ffb7d decomp/emu and the unpacked binaries gone: nothing used them; the byte
- 8a0e996 → 9a34ddf rules: terrain kinds of the unclassed squares as the reference (por-re
- 089cc1c → d1b5e49 rules: NPC share cuts the coins as one copper pile; tested against the
- f5f1706 → 6fe3c8c rules: aiout cases against the Pascal (taken out of play, the cell emp
- c2c651a → 6a30ffb rules: AddToParty orders against the Pascal's AddCharToParty (por-re 5
- ac3cd2c → 74643ba rules: the locked door against the Pascal's ResolveMovement (por-re a2
- 0398203 → 77ca7ad rules: the round's end and Continue Battle against the Pascal's EndOfR
- 0d3574c → 33298b5 rules: bashDoor and pickLock against the Pascal on a shared seed (por-
- 2fc077f → 2d7ce55 rules: TREASURE's item block as addTreasureBlock, tested against the P
- 6c3a4e9 → f4c439c rules: bashDoor matches the fixed Pascal BashDoor on every seeded case
- 3e13899 → a3059ee rules: TREASURE's random items as addRandomTreasure, equal to the fixe
- 262a1b1 → cb70b13 rules: opPicture's flags against the Pascal's EclPicture (por-re 42f41
- af68a79 → 282f548 rules: the camp rest against the Pascal's CampRest on a shared seed (p
- bba11c7 → 22d0672 rules: the rest tick against the Pascal's CampRest (por-re fa0a199), a
- b5cb20f → 4dace42 rules: WHO's answer against the Pascal's EclWho (por-re 81660dd), all 
- d4c1df1 → 33a6f99 rules: aiKeys as the original's poll on a pending aiKey, against the P
- 3b73456 → a8e592e rules: the ten defect rows now match the fixed Pascal with no exceptio
- 4288ea6 → 745f0fa rules: documentedMorale's comment: only the walk differs

## Felix's later answers

- Presentation has a second, engine-owned RNG. Gameplay keeps one stream.
- Module NPCs are in the ruleset as special/unique types; their names come from the adventure.
- A glossary fixes "character", "fact", "patch", "situation" and "world".
- Alignment is in `dnd` only: BR has none (its creation sheet shows race, sex and career).
- Table base as above. Gems and jewelry stay items, converted to copper once appraised and sold; `Coin` constants for them removed. Multi-class truncated split stays. Shop prices: the adventure answers when asked. Rule changes as separate commits.

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

## Pascal seams against core

Read: the interfaces of `classcal`, `charhlp`, `fxengine`, `cbtacts`, `combat`, `combatai`, `explore`, `postcbt`, `money`, `gametime`, `eclvm`; bodies only where cited. No core signature changed: nothing read so far is better than what core has.

| Pascal | Core | Result |
|---|---|---|
| `allowAlly` consent parameter | `Attack`, then `Outcome.Confirm` and `Confirmed` | same choice, kept |
| `aimX, aimY, aimAt`, `onWhom` | `Cast.Target`, `Cast.Ats` | fit |
| `VMenuChoice`, `EncounterChoice`, `InputNumber` ... seam variables | none: the interpreter is the engine's | engine side |
| `CombatMain`, `EndOfRound`, `eclvm` | engine loop, `EndRound`, `Verdict`, `Fact` | split at that line, vectors compare engine plus leaf |
| `CombatAiControl` runs a whole turn, drawing as it goes | `Decide` returns one action at a time | fits only if draw order survives: compare at whole turns |
| `TimeOutEffects(count, unit)`, steps of at most 10 minutes | `Elapse(minutes)` | fits; the leaf takes the steps |
| `RecalcClassStats` on a record pointer | values; `SetFact` has no side effect | the Pascal writes character bytes without recalculating except at `eclvm.pas:1074`, so purity agrees |
| Coins per character: `PoolMoney`, `ShareMoney`, `CarryingCapacity`, `EncumbranceAdjust` count coin weight | `Purse()` one purse in copper | **needs a ruling**: see Remaining questions |

Not read: spell and effect handler bodies, `combatst`, `mapdata`.

## Baseline and options

Felix on the rulings and tables: "These are unifications, and we'll treat them the same way as bugfixes: They are allowed and become our baseline." So the rulings and the table base stand. `Options` hold only his additions: critical hits and party inventory. Bug fixes and unifications land as separate commits after each game's port.

## Remaining questions

1. Fix or keep: SotSB first cleric die 2, paladin L15 2460001, party strength wrap.
2. Coin weight: the one party purse has no carrier, while the Pascal slows a character for coins carried. Does coin weight go, as the port has it? It belongs under ruling 12.
3. Rule tables generated from EXEs checked in: private only? Adventures with own decoders: none, importer only?
