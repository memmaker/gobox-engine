# Handover: one canonical game contract, passive games (2026-11-06)

## Goal (user decisions)

0. **HARD RULE: gobox has no per-game settings.** It is oblivious to which game it runs: no flags, modes or constants that games set differently, and no game names. Code is either unified (the originals agree) or it stays in the game. Game data passed in is fine.

1. **One canonical way per concern.** Games keep no private copies of gobox
   plumbing such as Ask, Question, Decision, Event, Mode or Result. The contract is
   `gobox/game/play`: `State()`, `Do(Action) (Outcome, error)`, `Save`, `Load`.
2. **Games are passive.** A game's resting state is always "waiting for player
   input", never "in the middle of something". `Do` changes state and returns.
   **No callbacks, continuations, coroutines or replay engines.** A pending
   decision is plain data in the state, and the player's answer is the next
   action.
3. Front state (typed-so-far, pause timing, Esc handling, layout) belongs to
   gobox. Games never draw UI or handle keys.

## Design every game follows

- **ECL VM:** the VM struct holds the pc, the vars, a frame stack (gosub, plus
  what used to be `then` chains, as frame kinds), and a
  `pending *play.Question` with the var its answer goes to. `run()` executes
  until an asking op or until the stack is empty. A reply stores the answer,
  advances the pc and runs again. All of it can be saved.
- **Flows** (shop, camp, loot, combat, party menu, creation, temple): each is
  an explicit sub-state struct on a small screen-state stack, with adventure at
  the bottom. `then`/`done`/`back` closures become "pop to the parent".
- **Answers:**
  - Esc on a question: `play.Leave{}`.
  - Member pick: `play.SelectMember{Member}`.
  - List row flags: `play.Question.Rows []game.Line`.
  - Temple services, including Cure Disease: `play.Heal{Service, Member}`.
  - A pause: a `Continue` question.
- **Work order inside a game:** convert one flow per commit, keeping the rules
  tests green; the old adapter may read the new state meanwhile. When no
  closure is left, implement `play.Game` natively and delete the game's
  adapter (play.go, gameadapter/gamecalls, contract*, the copy of State)
  together with its private plumbing types.
- **Builders:** give one bounded step at a time (for example "ECL VM only").
  Whole-game passes stalled three times.

## Phases (also in game/play/MIGRATION.md, "Order")

1. **gobox: done.**
   - One `play.Question`.
   - Renames in front: `Menu.Commands`, `Ask.Entries`.
   - Guards in `parity/nocopies_test.go`:
     - `TestNoPrivatePlumbing`: `knownCopies` only shrinks.
     - `TestPassive`: `funcFields` caps each game's stored callbacks and only
       shrinks.
2. **Passive rebuild per game:**
   - SOTSB and POD are near-twins; keep them identical.
   - Then POR.
   - Then COTAB. Its replay engine goes when its last asking routine has an
     explicit step.
   - Then BR1 and BR2: extract their shared code into one package first.
3. **Final step:** when no game uses the old path, delete `game.Rules`, the
   Decision and flow fields in game.go, `front/playrules`, and classic's
   game.* conversion. gobox then reads play directly into front.

## Status

| Game | Done | Next | Progress file |
|---|---|---|---|
| SOTSB | `e60abcf` (dead key layer gone); `0d17fae` (dead create/modify/slot/walk/title/sheet code gone, −965 lines; creation lives on via Draft); wait-based WIP parked on `refs/wip/play-port` (reference only) | `8e41bd3` VM; `e8ec163` Heal; `f9ca03d` creation as data; `9166bf8` flows on a screen stack; `4ed42ec` Ready/Unready + MusicChange. `db5b85c` Engine is play.Game, cmd/blades on it; `064ccbb` PickMember/Number/EndIcon fixed. `f9dead9` Event/EventKind/Result gone; `ec5f81b` explore+camp verbs are play actions. `0c95e37` combat/shop/loot verbs are play actions; `7433df7` party menu + creation sheet are play actions (Call/partyCall/keepDraft and the party decision types gone; RemoveCharacter per member as POD). `e709305` pickers/icon editor/training/Quit/LootShare/TryDoor are play actions (frontAct), decision() gone. `baeb00e` Do answers Continue/Reply/Leave on the engine, decide() gone. Tests ported to e.State()/e.Do (`8340e80`..`9687439`; two bugs fixed on the way: `6119904` reading the play state changed the selected member, `4c2886a` gobox's pool pick on a member never reached SOTSB through play). Left on the adapter: TestPlayParity, TestPlayStateDiff (delete with it), TestAskAllowsOnlyReplies (port with 5c). Next: 5c the game.State builder (gameadapter State, ~240 lines) becomes the play.State builder, then port the ~220 test uses of g.Apply/g.State and delete gameadapter.go/gamecalls.go' Game | `sotsb-go/.claude-progress-play.txt` |
| POD | `cc86f49` (yes/no wrappers merged); `b68f757` (ECL VM passive: frames stack + pending kinds, the reference shape for SOTSB); wait-based patch in the session scratchpad (`waitbased-port.patch`, reference only) | `b047c8b` flows take no func; `0386de8` non-combat asks as frames; save-wrap bug fixed via grules.SavingThrow. `3de1e99` combat asks as frames; `b905fe1` Engine is play.Game. Steps 1-3 collapsed as SOTSB: `fc28a16` Event/EventKind/Result gone (play events native); `31e81a6` explore+camp verbs are play actions; `d9b129e` loot money, `ef9d131` shop/temple, `afae1b6` items, `34e2a78` combat verbs are play actions (engine verbs gone, CombatStep stays). Next: step 4 party/creation (Call/direct), step 5 State from the screen stack and delete the Contract adapter (contract*.go, decision(), the k.Apply tests); wait for SOTSB's 4-5 as the reference shape | `pod-go/.claude-progress-play.txt` |
| POR | `e8cf16f` (Bar/Choice/Entry/Pick/EventKind gone); `8cc99ee` (cont stack: GameLoop/RunECL/AreaChange/Encamp/Move); `10203bd` (settle/wait callbacks gone); `0160cc3` (ask closures gone, encounter/tavern as state) | 3b done (`404316e`, `35a9f3e`, `fca984f`, `b6c46d4`, `78354df`, `82bce5a`): no stored continuations left. Step 4 (native play.Game) is parked until SOTSB's native play.go lands as the reference shape (POR builder found no native example; 36 tests and the classic front ride GameRules); native play.Game; delete GameRules, reply()/gameAsk(), PickMember/PickSpell/PickClass/Confirm | `por-go/.claude-progress-play.txt` |
| COTAB | measured: 246 asking functions (screens 95, combat 67, plumbing 55, ECL 29) | starts after one of the three above lands; replace the replay engine | — |
| BR1/BR2 | — | shared package, then passive | — |

gobox commits:
- `6a2cfee`: passive rule and TestPassive.
- `aa73ed3`: Question.Rows; POR removed from knownCopies.
- Phase 1 commits earlier the same day.

## Checks

- **Guards:**
  `cd parity && GOWORK=off go test -run 'TestNoPrivatePlumbing|TestPassive' .`
- **When a game deletes plumbing:** remove its names from `knownCopies` and
  lower its `funcFields` cap. POD's entry still lists five types; remove them
  only once they are really gone.
- **Rules tests:** run `go test ./...` inside each game and inside each gobox
  module. Every gobox directory is its own module.
- **Failing before this work (don't chase):**
  - 29 of 48 parity tests.
  - sotsb `cmd/blades` `TestFlowOldNew`.
  - `game` `TestGameImports`: cotab→front/classic, por→sprites.

## Rules

- Commit only your own files; other sessions work in the same repos.
- Never commit .DS_Store.
- Run tests when work is complete.
- Scripted runs stay windowed.

## Shared-code extraction (user: "Yes, start them", 2026-11-06)

Plan: pool-shared-analysis.md (committed 6d8058e). Items 1-4 are running in one builder:
- dice/RNG and identical helpers
- combat geometry
- derived stats
- POR on the shared AttackHits/SavingThrow (POD's switch and its unsigned save bug go to the POD builder)

The builder skips files the passive-rebuild builders are editing. Progress: `.claude-progress-extract.txt`. Items 5-10 (affects, ECL VM into gobox/ecl, AI, spells, records, items) wait for the passive flows.

Extraction items 1-4 done (gobox d4da8dc, 092f785, fbcb576, 96f41d5, 61509da; per-game commits in .claude-progress-extract.txt): about 430 lines removed per game. Four divergences (nearSort tie order, directionTo strictness, POD StrBonus guard, COTAB Dice byte cast) are being checked against the originals. SOTSB TestJoinedName flake predates the extraction (7/30 at e8ec163), suspected `selected()` state in the join flow.
Divergences settled against the originals: nearSort tie order was a SOTSB/POD port bug (fixed, gobox c8203d4, lowDirFirst removed); Dice takes bytes in all four originals (gobox rules.Dice casts, TestDiceByteArgs); directionTo strictness is a genuine difference: gobox's grid.DirectionTo has no flag (SOTSB/POD rule) and COTAB keeps its own in its game code (a13b3aa, cotab 106b6c7). POD's StrBonus guard: POD's own. Audit of gobox for other per-game switches: per-game-settings-audit.md (running). Bug class: a per-game switch in shared code must trace to each original's routine, never to how ports happen to differ.

## Must move into gobox (user, 2026-11-06)

ECL VM · combat AI · spell tables and effects, memorizing, scribing · character records · items, purse, treasure split · affect engine keyed by affect ids · combat · derived stats.

Under the hard rule: what the originals do the same way is unified; where an original genuinely differs, the difference must be expressed as game data (tables, ids, records the game loads), never as a flag. If it can't be expressed as data, that piece stays in the game. Check every apparent difference against the -re projects or the decomp first, since many are port bugs.

Order (to avoid colliding with the passive rebuild):
1. Now, untouched by the rebuild: derived stats (rest: saves, thief skills, HP, XP, levels), spell tables, memorizing and scribing data, character records (gobox/char).
2. After SOTSB/POD collapse step 5 and POR step 4: the ECL VM (POD's passive VM as the base), the affect engine, items/purse/treasure (coordinate with the shared-inventory work), combat and combat AI.

## Resumed 2026-11-06 (new session, after the usage limit)

The previous session's builders (audit fixes, passive rebuild) are gone. Their uncommitted work is still in sotsb-go and pod-go (rules/*, see `git status`); check whether it builds before continuing it.
Audit (per-game-settings-audit.md), checked against the originals (notes: this session's originals-progress.txt):
- Done: #2 FreeAim (bd6f47c), #10 PointBuy (bae1a3c), #7 NoHotkeys (8f54636: lists get no invented keys in any game), #3 brVerbs (a7588ba), #11/#14 comments (00a3e95), #8 Unmarked (4350b18: every original marks the selected member), #4 rules.Config (ae4e4b0: the stoned skip and per-member roll stay in sotsb, O07D:210D, O050 @2E43), #5 CreditFrame (260eb02: Data.CreditBoxes, POR passes ovr02:0206's two), #1 CreationFlow (971efe4: no alignment field means one field per row; BR2 drops its dummy Alignment), #9 OneDisk (61a6b5a: no sibling fallback; POR's hits were tests reading with set 1).
- Open: #6 SaveNoConfirm/SaveUnnamed: all six originals save to letters A-J unasked; the user picks original letter slots or named slots with overwrite confirm for all. #12 GameID keys the portrait origins (optional: pass the origins as data).
- Port bugs found, not fixed: POR's trap damage skips stoned/gone members as a "fix" (original ovr08:20CC skips only dead); BR1 and BR2 list Gender before Race, the originals draw race above gender.
pod-go rules tests compile and pass again (fc28a16).
