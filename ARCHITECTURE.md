# Architecture: engine, rules, adventure

DRAFT for Felix's review. Nothing uses `core/` yet. Settled parts: ONE-PATH.md "Next: three components". Code: `core/`.

## Dependencies

```
            gobox/core            types, Rules, Adventure, neutral ids; imports nothing of ours
           /    |     \
       rules  engine  adventure   rules and adventure: leaves, import only core
       (1e,    |      (POR, COTAB, ...)
       XXVc)   |
   gobox: main loop, script machine, combat loop and board, front, saves, dax, geo
   cmd/<game>: main wires one Rules and one Adventure into the engine
```

The engine knows only the two interfaces. A rule never calls back: everything it needs from the engine comes in as a value (`Situation`, `Fight`). Guard: `core/imports_test.go` runs `go list -deps` over core and each module in `leaves`, and fails on anything but stdlib, core and own packages (planted once in `TestGuardCatches`).

## Interfaces

`Rules` has 26 methods (7 + `CharacterRules` 10 + `CombatRules` 9). `Adventure` has 11. Each method's doc cites its call in today's code.

**One fight**, engine's loop (methods all on `CombatRules`):
```
BeginFight(f)
for Verdict(f) == FightOn:
    BeginRound(f)
    for id, ok := NextActor(f); ok:
        t := BeginTurn(f, id)            // Skipped, Computer, Effects
        loop: Allowed(f) + input, or Decide(f, id)
              Resolve(f, action)         // Outcome: Effects in order; engine applies board ones, plays all
        EndTurn(f, id)
    EndRound(f)
EndFight(f)                              // Exp, Items, Money, Beaten; engine writes the script words
```
`Outcome.Confirm` is a yes/no with nothing done; a yes sends `Confirmed{action}` (por's "sure" call).

**Everything else** is `Allowed(situation)` and `Resolve(situation, action)`, one path as `play.Game.Do` is: `Assembling`, `Creating`, `Adventuring`, `Camping`, `Shopping`, `AtTemple`, `Training`, `Looting`, `AtDoor`, `Scripted`. A price is `Offer.Price`; goods and script words come in the situation. 62 action types, the rule verbs of `play`; at migration play's become aliases, never a copy.

**Characters**: the ruleset owns them, the engine stores `CharID` and owns party order and selection. `Spawn(Creature)` joins the two leaves: the engine takes a `Creature` from the adventure and hands it to the rules. `Random` keeps one dice stream for scripts and rules, so replays keep their order.

**Adventure**: `Scripts()` (dialect, quirks, exit rows, names, bindings), `ScriptBlock`, `ScriptFile` (names, patches), `TreasureNote`, `Area`, `Map`, `Creature`, `Items`, `Asset`, `Identity`, `Needs`. At load the engine runs `Mismatches(adventure.Needs(), rules.Catalog())` and refuses with the list.

## Where today's code goes

### POR (por-go/rules, 38.9k lines with tests; rules4e 2.5k)

| Files | Goes to |
|---|---|
| combat.go, combatturn1/2.go, combatrules.go, combatai.go, aoe.go, affects1/2.go, affecttypes.go, spells1/2.go, spellrules.go, charcalc.go, create.go, dualclass.go, train.go, temple.go, shop.go (rules part), money.go, camp.go (rest, memorize), powers.go, powercalls.go, aftercombat.go, items.go, itemtypes.go, randitem.go, tables.go (rule tables) | rules (1e) |
| rules4e/ | second ruleset, embedding 1e; `CombatRules` hooks become private to it |
| combatflow.go, combatsetup.go, combatmap.go, cmap.go, anim.go | engine, combat loop and board; the rule calls inside go to rules |
| ecl.go, eclops.go, eclops2.go, eclask.go, script.go, vars.go | engine, script machine; `ReadVar`/`WriteVar` become bindings |
| files.go (LoadMonster, item reads), geo.go, scriptvalues.go, areamap.go, gamedata.go, pic.go, set.go, lore.go, journal.go, quests.go, cluebook.go, bestiary.go, beastpics.go | adventure, decoding to neutral |
| native.go | adventure patches where they can be; see hard spot 3 |
| savefile.go, slots.go, files.go (saveSlot, loadSlot) | split: hard spot 5 |
| play.go, playask.go, playfacts.go, pending.go, cont.go, statelive.go, view.go, page.go, menu.go, nav.go, explore.go, party.go, and the rest of the presentation | engine, then deleted where gobox has it. Unread line by line: expect rule fragments in these |

### Contrast: BR1 (buckrogers-go/rules 19.8k, ecl-br 1.9k)

| Files | Goes to |
|---|---|
| combat.go, combatturn.go, ai.go, effects.go, skills.go, create.go, rest.go, loot.go (medics), shop.go, items.go, pack.go, throw.go, eclbr: fighter.go, item.go, record.go, tables.go, draft.go | rules (XXVc) |
| cmap.go, eclbr combatmap.go | engine board |
| ecl.go, eclops.go, eclbr vm.go, ecl.go | engine: a second machine, hard spot 2 |
| monster.go, treasure.go, adv.go, journal*.go, *_data.go, dsimg.go | adventure; dsimg is hard spot 9 |
| ship.go, space*.go, eclbr ship, space* (2.6k+) | **unplaced**: hard spot 8 |

ecl-br is the XXVc ruleset and half a VM in one module: it splits rather than moves. BR2 keeps its difference as data, so BR1 and BR2 could be one ruleset with two tables.

## Neutral ids and the script mapping

Ids (`core/ids.go`, `spells.go`, `facts.go`): `Ability`, `Race`, `Class`/`ClassSet`, `Sex`, `Alignment`, `Status`, `Side`, `Skill`, `Save`, `Condition`, `SpellID` (127), `Coin`, `Trait`, `Feature`, typed string constants; `Fact` is a closed set of typed values. Handles: `CharID`, `AreaID`, `CreatureRef`, `ItemSetRef`, `MapRef`.

Raw numbers stop at the adventure, as tables in `ScriptSet`:
- `CharBinding{Addr, Fact, Values}`: the engine's `GetVar`/`SetVar` looks up the address; a hit goes to `Rules.Fact`/`SetFact`, a miss to its flat memory. `Values` maps numbers: `{0x87, StatusStoned}`.
- `WorldBinding{Addr, Word}`: script words rules depend on; passed in `FightTerms`, `ShopTerms`.
- `ConditionTerm`, `SpellTerm`, `ItemTermRow`: operands of FindAffect, CheckParty, FindItem and memory slots.

**The scripts use little.** Survey: every ECL block of every game decoded, every variable operand in the character window counted (reads/writes, static):

| Game | Commands | Character words | Used |
|---|---|---|---|
| POR | 16,025 | 21 | 0 name 15/4, 24 con, 114 12/0, 155 save, 165-171 thief, 184 control 7/4, 187-195 coins, 216 align, 256 status 24/5, 268 side 0/7, 283 move; flat in the port: 1, 32, 50, 114, 242 |
| COTAB | 14,183 | 18 | 0, 21 int, 32, 114 race, 115 class, 165-171, 184, 195-196, 201 mage level, 214 sex, 216, 228 paladin, 235, 256, 268 |
| SotSB | 15,322 | 14 | 0, 25 cha, 27, 114, 118 max HP, 158, 167, 195-198, 208 ranger level, 256, 281 HP |
| POD | 33,721 | 31 | raw record, window 1025: 0, 16-27 stats, 30 memo, 173 race, 174 class, 178 max HP, 316-321 thief, 331-333 coins, 337-349 levels, 358 sex, 493 status 46/2, 508 HP |
| BR1 | 17,819 | 11 | 0, 38 gender, 40 career, 69 max HP, 76, 161-163, 256, 268, 281 |
| BR2 | 24,950 | 8 | 0, 19 int, 39 race, 69, 76, 127 skill 51, 256, 268, 281 |

Conditions queried at all: POR 19 find traps; COTAB 20 resist fire; SotSB 3 sticks to snakes; POD 25, 56, 72, 126, and detect magic through a word. That is why `Condition` has 8 members while the games name 125-142 affects each. Skills: BR1 47 and BR2 119 skill commands.

Spell numbers are nearly common and not the same: 36 is Animate Dead (POR, COTAB), Heal (SotSB, POD); 56 Restoration against Harm. Item powers sit in the same table in POD only (57 Potion of Speed). Hence per-module `SpellTerm`.

## Ruleset differences that stress the interface

There is **no 2nd edition game here**: the four Forgotten Realms games are the 1e line, BR is XXVc, and POR's 4e plug-in is the only other ruleset. 2e stays untested by this design.

| | POR | COTAB, SotSB, POD | BR1, BR2 (XXVc) |
|---|---|---|---|
| Abilities | 6 | 6 | 7 (Tech) |
| Saves | 5 | 5 | 8 |
| Classes | 8 + multi | same, dual class | careers, 85 skills, one level |
| Magic | 56 spells, 21 slots | 100-126 rows, POD 141 slots | none: about 32 effects, no slots, no `Cast` |
| Initiative | d6 + dex; `nextCombatant` rolls d100 per unit per pick | SotSB, POD sort once, coin toss per insert | d10 + dex + 1, tie roll stored, side tactics roll |
| Script operands | 16-bit | 16-bit, POD sized and raw record | longint, sized |
| Money | one purse, gold | purse, platinum in POD | credits, on the record, paid across members |
| Turn verbs | classic | classic | Aid, Sprint, Dodge, Intimidate, Jetpack, zero-G |
| Prices | fixed | factor word; POD platinum | clips, repair |
| Fights | board | board | board and ship fights |

`NextActor` hides the initiative split; a sequence of `NextActor` calls keeps each game's roll order only if the rules roll at the same moments.

## Hard spots: where the split does not fit

1. **Scripts writing raw character bytes.** Few, per the table, but not clean. POD and BR expose the whole record: BR writes the party memory and the record both, `o` 16..185 passes through (ecl.go setChar), so `Fact` needs a raw mirror or a proof that nothing beyond the table is touched. The count is static operands; a table-driven address was not looked for. Reads have side effects: COTAB `cvActive` clears `justJoined`. Writes set engine flags: POR off 0 `memberGone`. BR 76 is both a skill byte and a side bit. POR leaves 114 flat where COTAB binds race: bug or original, unverified.
2. **Two machines.** "The interpreter is the engine's", but ecl-br has its own for longint operands. The opcode handlers (cotab eclcmds 1.7k lines, por ecl and eclops 2.4k, BR eclops 0.8k) mix rule calls with text and screens. This is the largest piece of the migration, and core does not reduce it.
3. **Overrides that are code.** POR's `nativeScripts` (tavern, Phlan history) are Go on the Engine, written for one purse. An adventure holds no engine code, so they become script patches, or quirks in the engine by name. `Quirk` is a typed list of today's four; it is per-game behaviour in gobox, against the "game-agnostic" rule unless it counts as dialect.
4. **Spell effects scripts depend on.** Knock in POR area 14 writes memory slot directly (the `0x1F` extra); a ruleset without slots fails `Mismatches`, correct but it means POR's adventure does not run on 4e without a patch. "Any adventure on any ruleset" holds only down to these facts. Worse is silent: scripts compare numbers such as `PartyStrength` and thief scores on the 1e scale. A ruleset can pass the load check and misbalance every gate.
5. **Save files.** POR's save is the original's byte for byte: `saveFile` is area + game + aux words + ECL + position, with CHRDATA files; walls overlaid in the area record. Script memory is the engine's, characters the rules'. So `Rules.Save` is new, and old saves need a converter that knows both (por `SaveEditor`, fixtures, BR savegame.go, POD imports from SotSB). Trace fixtures depend on this format. Not resolved here.
6. **Combat AI.** It goes to rules, since it needs spell tables, and por's rules4e ai.go overrides it. But `computerTurn` is one long routine mixing rolls, moves and attacks; `Decide` returning one action per call needs state kept across calls and the same RNG order. Reimplementing it risks every combat trace. It reads the board by value, costing a copy per decision.
7. **Board facts.** `Board` has `Blocked` only. `game.Ground.Height`, BR ground kinds (`bareGround`), zero-G and POR's `FreeSteps` encounter distance are unmapped; `Board` will grow at migration. Setup (combatsetup.go, 773 lines) mixes rule rolls with placement, unsplit.
8. **BR ship fights** (1.6k lines) are not in this interface: only `FeatureSpaceCombat`, so they cannot yet run through it. Their view types are in `game`.
9. **Rule data from the game's files.** BR reads rule tables from START.EXE at run time (dsimg.go, `tab`); POR's are checked-in Go. A leaf ruleset may not know the adventure, so who loads them? Only the `cmd` wiring can pass them to the ruleset's constructor.
10. **Items.** Records carry rule fields; the base table (por DS:5582, BR ITEMS) is rule data in a module's file. `Item` here is an untested proposal, `Trait` is a bare string with no list, and monster SPC blocks (raw affect numbers) need a row each. `randomItemType` rolls treasure in the rules but is fed by a script.
11. **Script words as rule inputs.** `WorldWord` lists 11 found. More hide in the games: not surveyed completely (POR's `RestEvery`, `Duel`, POD's vault).
12. **`game.Data`.** Mixes module facts (art, music, credits), series look (verb keys, party menu) and rules (`Rests`, `SpellRules`). No `Manifest` is drafted because splitting it was not traced; until then the engine would still get it per game.
13. **Narration.** Rules produce words ("flees in panic", BR's `word()` table from the game's files). Same issue as 9.

## Migration order (proposal)

Each step builds and keeps all six games on their traces.

1. Review this; settle the open questions; core stays unused.
2. **POR, read only**: the bindings as data, checked against today's `ReadVar`/`WriteVar` by a test. Cheap, and tests hard spot 1 before anything moves.
3. **POR adventure**: `Adventure` over dax and scripts, with `Creature` and `Item` decoded and compared to today's records.
4. **The VM's opcodes to the engine**, for POR: they call a Rules-shaped private interface in por-go.
5. **POR rules leaf**: the 1e package, combat last; rules4e on it as the proof of a second ruleset.
6. **POR on the engine loop**: delete the por-go presentation.
7. **SotSB, COTAB, POD** on 1e, one at a time: their differences decide how many are per-game data. `RULE2.md`'s per-game shop and XP rules test this.
8. **BR1, BR2**: XXVc leaf, then the second machine, ship fights last.

Unverified: nothing here has run against a game; only core's build, vet, tests. No method has been tried by an implementer, so expect changes in step 5.

## Open questions for Felix

1. **One ruleset or six?** Do COTAB, SotSB and POD share a 1e ruleset (parameters), given RULE2 keeps shop and XP per game? Recommend one, per-game differences as constructor data.
2. **"Imports only core"** literally forbids `gobox/rules` (dice, stats, purse, grid) in a ruleset, and `gobox/dax`, `geo` in an adventure. Recommend kits that import only core, allowed by the guard. Otherwise each leaf duplicates them.
3. **Old saves**: keep byte compatibility with originals, or a one-way import? Recommend import; it frees the formats.
4. **Numeric parity**: must every game keep byte-exact RNG traces through the migration? It decides whether AI and setup can be restructured or must be ported call for call. Recommend yes until each game has moved.
5. **Quirks** in the engine, by name, acceptable as dialect? Or must POR's scripts be patched until none remain?
6. **Ids as typed strings** against numbered enums? Rule 13 reads either way. I chose strings: stable in saves, clear at load errors.
7. **Who loads rule tables** from game binaries (hard spot 9)? Recommend the ruleset constructor, called from cmd.
8. **Ship fights**: part of `Rules`, or a separate optional interface? Felix's "two literal interfaces" suggests inside; recommend deferring until step 8.
9. **Is "any adventure on any ruleset" a load-level promise only?** Numbers in scripts (hard spot 4) cannot be checked.
10. **Clock**: drafted as engine-owned since scripts hold it (`Clock` words), with `Elapse` for rules. Agree?
11. **Domain language**: fix "character" (includes monsters), "fact", "term", "situation" in a glossary?
