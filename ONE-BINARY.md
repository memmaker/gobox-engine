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
- [ ] inventory of chargen in POR baseline c70cbee vs each game's current one (steps, picks, portrait and combat-sprite selection), and gobox's picker
- [ ] one shared flow in gobox on POR's baseline: rules from CharacterRules, presentation in gobox only
- [ ] portraits and combat icons from the one shared, deduplicated collection every game contributes to
- [ ] every game's chargen runs through it (keyboard, mouse, pad; latch)

## Track P: parity with the baseline and the title sequences
- [ ] feature and flow list of por-go c70cbee (PORT_CHECKLIST "Extras", MODERNIZED.md, git log) against today's engine: what is lost
- [ ] the title-sequence changes per game: list the commits, keep them in the declarative sequences
- [ ] restore every lost feature in the shared engine
- [ ] per game: plays from its folder only; scripted windowed run to the first fight and a save/load

## Open, unverified
- B's one uncaptured test panic (not reproduced, count=10 and -race)
- in-fight sounds by cue name may be silent; fall-sound fallback; toasts vs pane lines; rest-time line refresh (F list)
