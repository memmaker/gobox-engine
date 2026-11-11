# One binary: every adventure as imported data (checklist and handover)

**Goal (Felix 2026-11-11).** One `gobox` binary plays all six games (POR, COTAB, SotSB, POD, BR1, BR2) from imported data folders `adventures/<id>/` only; no per-game executable, repo or run-time read of originals. The engine has every feature and flow por-go had at the baseline **c70cbee (2026-10-31)**, the last commit before the engine redesign, plus the decisions since (ARCHITECTURE.md). The title sequence changes already done per game stay. Character creation, portrait and combat-sprite selection are one flow for all games: POR's baseline flow, over gobox's one shared, deduplicated portrait and icon collection.

**How.** ARCHITECTURE.md "One binary, adventures are data" is the design. Four agents at most, each on one track below, each commit gated (`gate.sh` under `/private/tmp/por-gate.lock`), staged with `git add -p` only for its own hunks, `git diff --cached --stat` before commit. Targeted tests only during work; full suites and a scripted windowed run per game once at the end. Progress lines in the scratchpad's agentR.log / agentA.log / agentC.log / agentP.log. Every agent ticks its lines here in its own commit.

## Track R: rules and engine loop (agent B3, then successors)
- [x] `Purse(CharID) Copper` in core (gobox d995a8a)
- [ ] actor at the core seam (Felix: option b)
- [ ] POR's rules into rules1e behind core.Rules / CharacterRules / CombatRules; TestAgainstPascal equal
- [ ] gobox's one adventure-folder reader implements core.Adventure; the manifest (id, title, ruleset, start/end sequence, custom fields)
- [ ] engine loop in gobox; `cmd/gobox` lists folders, starts one with its ruleset; `cmd/por`, `por-go/rules` gone
- [ ] rulesxxvc for BR1/BR2 (from buckrogers-core and the BR rules), behind the same interfaces
- [ ] the per-game flow differences (vault, ending, wilderness, doors, stack split/repair, level titles, import offers, difficulty, camp options) as shared features or manifest data

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

## Track P: parity with the baseline and the title sequences
- [x] feature and flow list of por-go c70cbee (PORT_CHECKLIST "Extras", MODERNIZED.md, git log) against today's engine: what is lost
- [x] the title-sequence changes per game: list the commits, keep them in the declarative sequences
- [ ] restore every lost feature in the shared engine
- [ ] per game: plays from its folder only; scripted windowed run to the first fight and a save/load

## Open, unverified
- B's one uncaptured test panic (not reproduced, count=10 and -race)
- in-fight sounds by cue name may be silent; fall-sound fallback; toasts vs pane lines; rest-time line refresh (F list)
