# One binary: every adventure as imported data (checklist and handover)

**Goal (Felix 2026-11-11).** One `gobox` binary plays all six games (POR, COTAB, SotSB, POD, BR1, BR2) from imported data folders `adventures/<id>/` only; no per-game executable, repo or run-time read of originals. The engine has every feature and flow por-go had at the baseline **c70cbee (2026-10-31)**, the last commit before the engine redesign, plus the decisions since (ARCHITECTURE.md). The title sequence changes already done per game stay. Character creation, portrait and combat-sprite selection are one flow for all games: POR's baseline flow, over gobox's one shared, deduplicated portrait and icon collection.

**How.** ARCHITECTURE.md "One binary, adventures are data" is the design. Four agents at most, each on one track below, each commit gated (`gate.sh` under `/private/tmp/por-gate.lock`), staged with `git add -p` only for its own hunks, `git diff --cached --stat` before commit. Targeted tests only during work; full suites and a scripted windowed run per game once at the end. Progress lines in the scratchpad's agentR.log / agentA.log / agentC.log / agentP.log. Every agent ticks its lines here in its own commit.

## Track R: rules and engine loop (agent B3, then successors)
- [x] `Purse(CharID) Copper` in core (gobox d995a8a)
- [x] actor at the core seam (Felix: option b; por-go 7900727, TestActorSeam)
- [ ] POR's rules into rules1e behind core.Rules / CharacterRules / CombatRules; TestAgainstPascal equal
- [ ] gobox's one adventure-folder reader implements core.Adventure; the manifest (id, title, ruleset, start/end sequence, custom fields)
- [ ] engine loop in gobox; `cmd/gobox` lists folders, starts one with its ruleset; `cmd/por`, `por-go/rules` gone
- [ ] rulesxxvc for BR1/BR2 (from buckrogers-core and the BR rules), behind the same interfaces
- [ ] the per-game flow differences (vault, ending, wilderness, doors, stack split/repair, level titles, import offers, difficulty, camp options) as shared features or manifest data
  - [x] R1 one receiver per file: por-go 6b97aff, the Engine methods of the mixed files in `<file>_engine.go`; powercalls.go no longer imports gobox/game. create.go stays mixed (agent C's).

### R handover (agent B4, 2026-11-11)
Measured at 6b97aff: RuleState methods sit in 49 files (13.2k lines); Engine embeds RuleState and its 90 files use ~300 promoted names. Step R2 (move into rules1e) is blocked by the import rule (rules1e: gobox/core, core/dnd only). The RuleState files still import:
- `gobox/rules` (GoldBoxRates, Purse, MakeFileName, Sign, Abs, RollDice, DexMissile) and `gobox/rules/grid` (LineOfSight, Sight, InArc, NearSort: cmap.go, spells1.go). Needs Felix: allow them in rules1e, or move them under core.
- `gobox/ecl`, `gobox/ecl/script` (ecl.go, eclops*.go, scriptvars.go, saveimport.go, engine.go): the script VM's state sits in RuleState. It is the shared engine's, not the ruleset's: cut it out of RuleState first (the script calls the rules through core.Rules).
- `gobox/game`, `game/play` (gamestate2.go, extras.go, poolid.go, create.go, engine.go): presenter types (Entry, SheetPage, Hold, Window aliases in types.go) go to the Engine side.
- `gobox/asset`, `ega`, `por/pal`, `geo`, `image` (battle.go, train.go, engine.go, geo.go, 12 files with image.Point): art and palette leave RuleState; image.Point needs a core position type (none in core: interface change, needs Felix).
Next order: (a) per file, move the free funcs and types that only the Engine uses into the `_engine.go` side (pure moves); (b) cut ECL and art state out of RuleState; (c) Felix's ruling on gobox/rules, grid and a core point type; (d) then the package move with gopls rename for the promoted names (`gopls rename` per identifier, scripted from the Engine files' `e.<name>` uses). R3 to R6 not started.

## Track A: import and run-time data (agent A5, then successors)
- [x] art as PNG atlases, all six games (TestArtEqualsDAX, 46023 frames)
- [x] POR maps, monsters, items as bin+txt / JSON (TestDataEqualsDAX)
- [x] quests and journals as script commands (POR, COTAB, SotSB)
- [ ] sound: JSON synth data, MIDI music (POR started: gobox 186dbf5, por-go c383aa1)
- [ ] no run-time originals: no POOL.CFG / TITLE / 8X8D1 check
- [ ] Shared Purse option (Felix's OFF-mode calls in ARCHITECTURE.md), WhoPays, CastKnock, nativeScripts gone
- [ ] `gobox import <game>`: one entry, per-game importers on the import path only, writing `adventures/<id>/`
- [ ] importers for COTAB, SotSB, POD, BR1, BR2: scripts (with their fixes), art, maps, records, sound, tables, title/end sequences as declarative data

## Track C: character creation (one flow)
- [x] inventory of chargen in POR baseline c70cbee vs each game's current one (steps, picks, portrait and combat-sprite selection), and gobox's picker
- [ ] one shared flow in gobox on POR's baseline: rules from CharacterRules, presentation in gobox only
- [ ] portraits and combat icons from the one shared, deduplicated collection every game contributes to
- [ ] every game's chargen runs through it (keyboard, mouse, pad; latch)

### C1 inventory (agent C, 2026-11-11)
POR baseline c70cbee (engine/create.go, pickgrid.go, icondesign.go, quickstart.go): creation is ONE screen, the sample character's sheet. Start: human male fighter, rolled. Name typed in place (inlineInput, 0x20..0x7A, 15). Sex, race (pickRaces: no half-orc), alignment, class are drop-downs over the sheet (current value first, grey = illegal for the rest; a grey pick forces dependants via kind.enforce, new race/class/alignment = fresh rolled character keeping name, pictures, items). Stats typed in place ("18/50"). Head, body: grid pickers (pickGrid). Combat icon: iconEditor (parts, colours, size; Amiga order). Bar: Roll, Minimal (roll to class minimums), Keep (SaveCharacter); Esc asks "Abandon this character?". Quick Start: the empty party menu creates the quickParty templates (kind, name, head, body), rerolled until prime abilities are high.

gobox today: the presentation is already one flow for all six. front/creation.go (CreateAct: Create, Modify, Roll, Name, Stat, SetField, Keep, Close), front/classic/creation.go (sheet layout from game.Creation: Fields, Stats, Derived, Minimums; one field per row when there is no alignment field = BR), front/picker.go + front/classic/pick.go (grid picker, icon editor, defaults by name: c4eaf71, 35b573e), front/present/poolpick.go (head/body/icon-head/icon-body pickers over sprites.Names: the game stores only the pool name via SetChoice), front/present/pointbuy.go, front/latch.go. Quick Start reaches every game through play (front/classic/partymenu.go).

Shared collection today: gobox/sprites, built by sprites/cmd/poolsprites, deduplicated by exact pixels, origins.json per game. TWO series: `pool` (por, cotab, sotsb, pod, + <game>-amiga sets, never deduplicated against DOS) and `buck-rogers` (BR1, BR2). Combat icons in-game stay on DAX (recoloured by palette index); only the pickers show pool PNGs.

Where the six differ (all on the rules side, not the drawing):
- the draft state machine (fields, legality, enforce, roll, minimums, rekind, keep) is written six times: por rules/create.go+play.go, cotab rules/creation.go+flow.go, sotsb rules/creation.go+partyflow.go, pod rules/draft.go+facts.go, BR1/BR2 rules/create.go (brcore DraftStep; careers, skills, no alignment). quickstart.go six times. cotab keeps its own rules/iconedit.go.
- core.Draft (core/character.go) carries only Sheet, Step string, Lows, MinHP/MaxHP; the choices go through Allowed(Creating) / play.SetChoice with Field names as strings ("portraithead", "iconbody", "race" ...): logic keyed on strings (rule 13).

Known bugs / smells: Step and SetChoice.Field are strings; BR pool picks were refused before (handover.md, fixed in BR); pool and buck-rogers are separate collections, against "one shared collection"; per-game duplicated draft logic violates "one canonical implementation".

C2 blocker (needs Felix, core interface): moving the draft flow into gobox needs CharacterRules to state the creation's data and legality, e.g. `Creation() CreationRules` with typed `Fields() []FieldDef` (id, options), `Fits(Draft, FieldID, option) bool`, `Enforce(Draft, FieldID) Draft`, `Roll(Draft, RollKind, *RNG) Draft`, `Minimums(Draft)`, `QuickParty() []Template`, and typed FieldID instead of Step/Field strings. BR careers and skills fit as fields plus a skills step. APPROVED by Felix 2026-11-11: Creation() CreationRules, typed FieldID, the string Step/Field names go; BR careers/skills as fields plus a skills step.
C3 decision (needs Felix): merge `buck-rogers` into the one `pool` series (BR art is VGA, Pool EGA; dedup would never match across them, so the merge is only one listing), or keep two series per art family. DECIDED by Felix 2026-11-11: two families, Pool games pick from `pool`, BR games from `buck-rogers`, each deduplicated within itself. In-game combat icons come from the collection, not DAX, where the pickers already use pool PNGs.

Handover (agent C, stopped at 527b2ee): core/creation.go holds CreationRules, FieldID, FieldDef, RollKind, Template, standalone (not embedded in CharacterRules, so every ruleset still builds). Next, in order, one gated commit each:
1. gobox flow over CreationRules: a draft loop in game/play (or front/classic/rules.go) doing what por c70cbee editSheet does (rekind on race/class/alignment, Roll/Minimal/Keep, Esc abandon), feeding the existing game.Creation so front/classic/creation.go stays the drawing; Quick Start from QuickParty.
2. Port por first (rules/create.go, quickstart.go are C's; B3 moves the rest into rules1e: git add -p only), then cotab (delete rules/iconedit.go too), sotsb, pod, BR1, BR2 (careers, skills step); GATE_ALL=1.
3. Embed Creation() CreationRules in CharacterRules; delete Draft.Step and the SetChoice field strings.
4. In-game combat icons from the series PNGs (front/present/poolpic.go: icons stay on DAX for palette recolour; the recolour must move to the PNG, check what each game recolours first).

C2 stop (agent C2, 2026-11-11, no code commits): step 1 cannot be written over CreationRules as approved; needs Felix (core interface). The loop that gobox owns needs, and CreationRules lacks:
- the current picks: core.Draft has Sheet, Step, Lows, HP only, no value per FieldID, so the loop cannot show a field's pick (game.Field.Pick) nor call Fits for dependants. Proposal: `Draft.Picks map[FieldID]int`, and Draft.Name, Head, Body, IconHead, IconBody (pool names) so pictures and name are the loop's, not SetChoice strings.
- finishing and starting: Keep (POR SaveCharacter) and Modify (a Draft from a member). Proposal: `Make(Draft) (CharID, error)` and `Draft(CharID) (Draft, error)` on CreationRules; `CharacterRules.Draft() *Draft` and core.KeepDraft/DropDraft go, since the draft lives in gobox.
- stats typed in place (POR setStatTyped, limitStat, "18/50"): Proposal: `SetScore(Draft, Ability, int32) (Draft, error)` (error: out of the race/class range), the 18/xx parse in gobox.
- BR's skills step: a FieldDef kind (`FieldSkills`, points to spend) or a separate `Skills(Draft) []SkillDef`; which?
Also: nothing implements core.Rules yet (rules1e has creation tables only, R3 chargen open), and the six games drive creation through front.Game (POR) or play (others) with their own draft state; the gobox loop would first sit behind front.Game as an adapter (front/creation.go), each game supplying CreationRules. Answer these and C2 resumes at step 1.

## Track P: parity with the baseline and the title sequences
- [x] feature and flow list of por-go c70cbee (PORT_CHECKLIST "Extras", MODERNIZED.md, git log) against today's engine: what is lost
- [x] the title-sequence changes per game: list the commits, keep them in the declarative sequences
- [ ] restore every lost feature in the shared engine
- [ ] per game: plays from its folder only; scripted windowed run to the first fight and a save/load

## Open, unverified
- B's one uncaptured test panic (not reproduced, count=10 and -race)
- in-fight sounds by cue name may be silent; fall-sound fallback; toasts vs pane lines; rest-time line refresh (F list)

## Track X: rulesxxvc (agent X, 2026-11-11)
Goal: `~/Projects/rulesxxvc` (module `rulesxxvc`, imports core and core/xxvc only, `imports_test.go`) implements Rules, CharacterRules, CombatRules, ShipRules for BR1 and BR2 from buckroger-re / buckrogers2-re. BR1/BR2 differences are data or options.
- [x] X1 map (below)
- [x] X2a module, import guard, dice, SkillValue/SkillCheck (charhlp.pas SelectSkillCharacter; rulesxxvc 8f1c8da)
- [x] X2b skill prerequisites (charhlp.pas MeetsSkillPrerequisite, c_skill.pas; SkillNo, SkillLevels; rulesxxvc 32eb5a4)
- [x] X2c attack arithmetic (AttackDie, Thac0Modifier by Difficulty, AttackHits, ProjectileRangePenalty, RangePenalty, ArmorValue; rulesxxvc 849f9c5); SaveDie, SaveMade (fxengine.pas RollSavingThrow; d1acc6f)
- [ ] X2d next: skill groups and career skills (gdconsts.pas skillDefinitions, careerSkills), damage (CalcWeaponDamage), initiative, ship-fight arithmetic (shipcbt.pas); then a Rules type over typed characters once gaps 1 and 3 are answered

Handover: every rulesxxvc function names its .pas; check BR2's copy before porting (so far identical); tests are hand-derived from the .pas, TP Random values computed outside Go.
- [ ] X3 BR games keep their own code until wind-down (untouched so far)

**Where BR's rules are today.** buckrogers-core (module `ecl-br`, 1.9k lines) is NOT rules: the byte Record (record.go, raw offsets), VM glue, views, combat map; it imports gobox/game, game/play, ecl. The rules live in `buckrogers-go/rules` and `buckrogers2-go/rules` (15k lines each, near copies: ai, char, rest, ship, spacefight, monster identical; create 151, combat 229, combatturn 330, items 89, skills 89 diff lines). Nothing there is pure: every rule is an `*Engine` method over `*eclbr.Record` bytes and reads tables from the START.EXE data segment (dsImg, `tab`). Most rule-like, little flow: char.go (record, train, hpRoll), skills.go (groups, unlocks: Engine only for asking), combat.go skillCheck/skillVal (:213), ai.go, monster.go, effects.go, ship.go. Rules entangled with flow/drawing (import game/play): create.go, combat.go, combatturn.go, throw.go, spacefight.go (1573 lines: arithmetic sRoll, sToHit, sDamage, hitShip, ram, dodge mixed with screens and asks), rest.go, loot.go, shop.go, items.go. Port path: re-derive each rule from the Pascal (charhlp, combat, combatai, shipcbt, skills, chargen .pas), check it against the Go, never copy the byte-record code.
BR1 vs BR2 reference: SelectSkillCharacter is the same rule (BR2 uses shr and signed div, equal for non-negative values); the BR1 Go keeps the original's quirk that the bands use the last candidate's value (kept in rulesxxvc).

**core/xxvc today:** Tech, Terran, five careers, BattleTactics, Credits, Fuel. Missing: the other races, 83 skills, the eight saves, ship systems beyond fuel ("to confirm").

**Gaps in core for XXVc (for Felix; not changed):**
1. Tables: BR1 and BR2 have different START.EXE data (items, careers, monsters). A ruleset receives nothing but Campaign{MaxLevel, Spells} and Options, so rulesxxvc cannot tell which game's tables to use. Needs a table/edition selector in Campaign (or the tables as data through Begin).
2. `SkillRoll.Difficulty int32` and `Answer.Value` for its grade are untyped numbers; rulesxxvc has `Hardness` (VeryEasy..VeryHard) and `Grade` (FailedBadly..MadeWell). A typed core difficulty/grade, or a documented mapping, is needed.
3. `CreationRules` (core 527b2ee) has only option fields; BR's skills step is a point allocation over 84 skills with unlocks and refunds (skills.pas, pO0DD_09D8). Fields() cannot express it: needs a points step (skills, budget, Spend/Refund legality).
4. `core/imports_test.go` leaves lacks `../../rulesxxvc` (one line in core's test).
5. The BR Go draws from PCG; rulesxxvc draws from core.RNG (Turbo Pascal's Random), matching the reconstructions' order. BR saves carry no RNG to keep.
No BR pure-Pascal vector harness exists (rules1e's TestAgainstPascal has por-re's); rulesxxvc tests use hand-derived vectors from the .pas with TP Random computed outside Go. A BR vectors harness would need buckroger-re's toolchain.
