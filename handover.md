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
| SOTSB | `e60abcf` (dead key layer gone); `0d17fae` (dead create/modify/slot/walk/title/sheet code gone, −965 lines; creation lives on via Draft); wait-based WIP parked on `refs/wip/play-port` (reference only) | `8e41bd3` VM; `e8ec163` Heal; `f9ca03d` creation as data; `9166bf8` flows on a screen stack; `4ed42ec` Ready/Unready + MusicChange. `db5b85c` Engine is play.Game, cmd/blades on it; `064ccbb` PickMember/Number/EndIcon fixed. `f9dead9` Event/EventKind/Result gone; `ec5f81b` explore+camp verbs are play actions. `0c95e37` combat/shop/loot verbs are play actions; `7433df7` party menu + creation sheet are play actions (Call/partyCall/keepDraft and the party decision types gone; RemoveCharacter per member as POD). `e709305` pickers/icon editor/training/Quit/LootShare/TryDoor are play actions (frontAct), decision() gone. `baeb00e` Do answers Continue/Reply/Leave on the engine, decide() gone. Tests ported to e.State()/e.Do (`8340e80`..`9687439`; two bugs fixed on the way: `6119904` reading the play state changed the selected member, `4c2886a` gobox's pool pick on a member never reached SOTSB through play). `9cf8203` 5c: State built from the engine (e.facts(), gobox data types as values), the game.Rules adapter (Game, call, Apply, gameEvents) and both parity tests gone; SaveSlot/SlotSaver gone, Save/Load round-trip tested; funcFields cap 2. Next: fold facts into the play builders field by field if wanted; knownCopies Decision/Mode still SOTSB's | `sotsb-go/.claude-progress-play.txt` |
| POD | `cc86f49` (yes/no wrappers merged); `b68f757` (ECL VM passive: frames stack + pending kinds, the reference shape for SOTSB); wait-based patch in the session scratchpad (`waitbased-port.patch`, reference only) | `b047c8b` flows take no func; `0386de8` non-combat asks as frames; save-wrap bug fixed via grules.SavingThrow. `3de1e99` combat asks as frames; `b905fe1` Engine is play.Game. Steps 1-3 collapsed as SOTSB: `fc28a16` Event/EventKind/Result gone (play events native); `31e81a6` explore+camp verbs are play actions; `d9b129e` loot money, `ef9d131` shop/temple, `afae1b6` items, `34e2a78` combat verbs are play actions (engine verbs gone, CombatStep stays). Step 4 as SOTSB: `41ebb19` party menu decisions are play's (DropMember one for menu and Alter); `2ed34c1` creation sheet + Modify are play actions (Contract.sheet, engine sheetVerb/sheetDecision); `d864f01` pickers/icon editor/training/Quit/LootShare/labels are play actions (Contract.front), decision() gone; `7b1c455` Do answers Continue/Reply/Leave directly (k.answer, k.back); `27d827c` pool art pick reaches the selected member on any screen + TestStateKeepsSelection (State already read-only); `d167711` old front decisions converted to play before call (fixed cmd/pod TestCreationFront/TestKeyFlow); `42ae5b3` SlotSaver/SaveSlot gone. Step 5 as SOTSB: tests ported to e.State()/e.Do (`90fbf9d`..`19e906b`; on the way the draft's Allowed gained SetStat/Adjust/PointBuy and DiscardCharacter drops Modify's copy); `8bcfbbc` State built from the engine (e.facts(), gobox data types as values), the Contract type, its game.Decision mapping (Apply, run, decision, frontOf, Step, gameEvents, draftVerb) and call()'s game cases gone, contract*.go renamed facts/calls/draft.go, cmd/pod tests on playrules.New only. funcFields stays 2 (as SOTSB); knownCopies Decision/Mode still. Next: fold facts into the play builders field by field if wanted | `pod-go/.claude-progress-play.txt` |
| POR | `e8cf16f` (Bar/Choice/Entry/Pick/EventKind gone); `8cc99ee` (cont stack: GameLoop/RunECL/AreaChange/Encamp/Move); `10203bd` (settle/wait callbacks gone); `0160cc3` (ask closures gone, encounter/tavern as state) | 3b done (`404316e`, `35a9f3e`, `fca984f`, `b6c46d4`, `78354df`, `82bce5a`): no stored continuations left. Step 4 as SOTSB: `eabe6c2` Engine is play.Game (rules/play.go over GameRules; Save/Load(name) now SaveNamed/LoadNamed), `d199c5d` PromptRow24 keyWait outside a fight (class 11), `cdc186d` pool pick reaches the selected member (class 2), `fdf9469` cmd/por on playrules.New (cmd/por/rules.go opens the engine; TestPlayFront), `a3ddd29` party menu + creation sheet native in Do (partyDo; Drop without the adapter's askYes), `1e4ee23` pickers/icon editor/training/TryDoor native (frontDo), `6d8d16d` Do answers Continue/Reply/Leave on the engine, `6e53dab` fight/item/power verbs native, their questions pending data (ext.pq, playask.go; sure/yes/pick/askAgain/askYes gone), `d46549e` camp verbs native, `d1be9e7` item/shop/loot verbs native (shop Leave is play.Leave; treasure sure? pending). `d1abdea` explore verbs native, decision() gone (Do no longer goes through Apply); tests on e.State()/e.Do: `ad303f3` `5740cb4` `8576876` `15808fa` `5ce6356` (Buy/Sell close the Buy list they open: a buy stays Trading) `577b30c` (Struck/Fell named unit 0: EvDamage/EvDown carry it in Unit; a hit's damage now only Attacked's) `b812332`. `5481bab` (fightevents: harm vs HP drop over the whole outcome) `4a7bd6d` (devscene/aimoffer; Walked to ev.To, fight yes/no swapped, a timed spell's aim offered) `f3d9b5f` (combatbar/viewsheet; every member's Sheet, Drop in a fight) `b40ab0e` (nosilentdo/addchar; State's fixed order, Keep's name question, Flee refused) `a0d14a4` facts recorded at the source (CastSpell at the target square, sounds kept) `68573b2` GameRules deleted (Rules()/Apply/decide gone; bridge reads e.gameState()). funcFields 10→8. gobox `parity` no longer builds: ~20 of its test files drive POR through `e.Rules()` (adapter_test, ask_test, calls_test, combat_test, ...): delete or port them (not done: outside POR's files). por `d9bc3f9` game.State gone (private facts, playfacts.go/e.known(); the hall's training box was never Trading: stateMode had no ModeTraining), `a7043ce` gameAsk/reply gone (questions answered with play replies; a yes/no Scene listed Yes, No while Reply 0 answered No), `c3dac0c` PickMember/PickSpell/PickClass/Confirm gone (askSure/askSpell/askTarget/askClass pending data; ChangeClass with no class asks instead of erroring). funcFields stays 8 (none of the 8 is ask plumbing: spellFx, DevAction, EncodePNG, iconDrop, quest given, ...). Left: the engine's internal key layer (Action/Apply/command; WalkTo's auto-walk is reachable only by keys, walk_test), cmd/por/playfront_test on playrules (goes with step 5) | `por-go/.claude-progress-play.txt` |
| COTAB | measured: 246 asking functions (screens 95, combat 67, plumbing 55, ECL 29) | Steps 1-5 done: `6c0d369` Engine is play.Game; `0e5225f` the core takes play actions (decision() gone); tests on State/Do `75a5101`..`07da47f` (contract_test.go gone; cmd tests drive classic through playrules); `acc6fc6` Apply, playOf, gameEvents gone (core events are play's, combat verbs a named turnVerb); `e5d2569` game.State gone, State/allowed read a private facts struct. Step 6 begun: `f189541` frames (frames.go: frame, drive), the party menu a frame; `fab628e` DisplayInput a frame; `966dc9f`+`2c37bc2` the frame stack (runFrames, entryFrames: a frame's question is a checkpoint, so a converted stack replays one decision), Run's loop is mainLoop; `92fe44e` the adventure loop a frame (entryAdventure gone); `3f39142` play.Engaged at combatSetup. Next: ECL (runEvent/eventLoop and the asking commands as frames), then combat, then camp; replay (rules/core.go ask/unwind/deepCopy) goes when no ask is left outside runFrames, then the nocopies entry (unchanged so far: its types are the replay core's) | `cotab-go/.claude-progress-play.txt` |
| BR1/BR2 | shared package `brcore` (../buckrogers-core, game-specific so not gobox): `950fc55` Record/accessors/Fixed/SkillOK (originals' O0F6:2C47 table + each game's culled list)/SkillBase/LoadRecord; `4ae8f7e` ECL address map/Operand/Region/Width/GetN/PutN; `607aa60` `1d2fd85` LoadTables(ds, Layout, culled): creation/level tables, names, race careers/mods, career minimums read from the data segment at each game's Layout (BR1 no longer hand-copies them; this fixed a BR1 port bug: female career 5's combat icon was 0, the original reads 25 past its a0C41 row). Passive: `44bd9bb`/`7976d08` a fight's follow-up is a typed fightEnd; `abae1b9`/`aacd6e6` after-combat (20D8) a stage machine on postT (named stages, waits, typed postAfter); `475f38e`/`373b7c3` shop questions a typed shopWait; `bc4e5ea`/`7b9f87a` ECL waits a typed eclWait (vm.cont gone). ECL `then` a typed frame (3209/3117's return places; BR1 `5e45152`, BR2 `528161b`, BR2's test starts are startPlay); view question a typed viewWait (`773eb5e`/`0ba86e7`); every menu a typed menuQ with a return stack of menuBack (rest, change, order, purge, train, save, rocket, overland; BR2 load-from/quit too; `d3681fc`/`eb479de`), BR2's picker/training/Drop no longer matched by prompt text (`eb479de` `078458c`); the space fight a staged round (spaceStage on spaceT, typed questions, view's caller a typed viewBack; pq and every closure gone; `cb1a6dd`/`e9a50f3`); play.Engaged at fight start (`cdd427b`/`5f16908`). funcFields 1/2 = only the test trace hooks (gobox `b295968` `1291bab` `bfc7f3f`). Game commits: BR1 `03eeaf8` `9c30357` `5e3eba7` `548b7b0` `7035f70` `44bd9bb`; BR2 `a07ba61` `b392e51` `873cbf2` `034bfdb` `bf71416` `7976d08` | Native play.Game in POD's step order (delete play.go/Apply and the game.Rules path; then knownCopies), then the Engine move into brcore (~3400 identical lines are *Engine methods; rest.go/space fight differ only in strings and a few keys: data, not flags). BR1 skill/group names still hand-copied (check against dsImg). Unverified: the space fight and menus were checked by the rules tests only (no scripted playthrough) | `buckrogers-go/.claude-progress-play.txt` |

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
- **Mode never reached:** a mode the State switch handles that the
  game's mode function never returns hides that screen from gobox (POR's
  training box was Assembling). Detect: for each game.Mode case in a game's
  State(), grep its mode function (stateMode) for a `return` of it.
- **Reply order vs listed order:** a yes/no whose Replies are listed in one
  order while Reply N answers in another (POR's Scene: "Yes","No", Reply 0 =
  No). Detect: for every yes/no Ask (Scene, Common, Draft), Do the Reply at
  the index of "Yes" and assert the yes branch ran (POR TestSceneYesNoOrder).
- **Sub-screen left open by Do:** a Do that opens a list or picker to reach
  its engine call (a shop's Buy, a picker's Add) must close it again, or the
  state after it is the wrong kind (Trading becomes a Scene). Detect: for each
  verb, `Do` it and assert the state's Go type is the one it started in when
  the verb doesn't change screens (POR `TestNoSilentDo`, tablelen).
- **Play events naming ev.Unit:** Struck/Fell/Attacked must name the unit the
  engine damaged, not 0 or the actor. Detect: over a whole fight outcome, sum
  each unit's harm from its Struck/Attacked and require that unit's HP to
  drop by it; a Fell unit must be dead or gone (POR fightevents_test).
- **Pending question missing from the play state:** a question the core waits on must show in State() (an Ask, Draft.Ask or Continue in Allowed) and its answer must be allowed, or gobox shows the screen under it and the player can't answer (COTAB: Keep's name after the sheet closed). Detect: at every stop where the old Pending is non-nil, require the play state to carry an Ask/Draft.Ask or Continue, and Do(Reply) of a valid answer to succeed (cotab TestBugClass59).
- **Logic keyed on prompt text:** a question found by comparing `Ask.Prompt` with a string (BR2 picker, training, Drop by ovr(...)) breaks on any reword or translation. Detect: `grep -n "Prompt ==\|Prompt !=" */rules/*.go`; key on the typed wait that asked it (BR `mq`, `view.wait`, `post.drop`).
- **State lost crossing playrules:** a game.State field with no play carrier (Creation.Stats/Derived: BR1/BR2 TECH row; Ask option keys; a hand-kept step list beside Facts' own mode) drops silently on the shipped path while old-path tests pass. Detect: every game's front tests run on `classic.NewGame(playrules.New(e, d), d)` (as main.go) and assert the sheet's StatRows and Derived; `grep -n "play.Question{" */rules/*.go` each copies `Keys`.
- **Hand-copied data-segment tables:** a table typed into the port can stop at the array's declared bounds where the original reads past them (BR1 a0C41[gender, career], rows of 5, read at career 5). Detect: for every literal table whose comment names a DS address (`// d[0-9A-F]{4}`, `a[0-9A-F]{4}[`), compare it with the data segment at that address over the indices the code uses, or read it from the data segment as brcore.LoadTables does.
- **Mode set only on the old entry point:** state a game's old Apply/decide entry set (COTAB `c.gobox = true`) must also be set by Do, or the play path runs the old front's flow (the core asking its own aim menu). Detect: grep the old entry (Apply, call, result) for assignments to engine fields and check Do does the same; drive one aimed Cast through Do on a fresh engine and require no rules-side square/aim question (cotab TestGameFight).
- **Replay clone shares state through an interface:** a clone that copies pointers but returns interface values as they are shares whatever an interface holds with the checkpoint (COTAB deepCopy: a frame on the stack, so every replay stepped the live frame). Detect: for each interface-typed field reachable from the cloned state, mutate through the clone and assert the original is unchanged, or list the interface types deepCopy shares and justify each (cotab core.go frameType).
- **Event replay posed on the after-step state:** one step runs many turns, so its events must pose on the state before the step (board and party), not on the state it ended in (a fight lost: party empty, no figures; foes killed: never shown). Detect: any front code that draws a step's events from `R.State()` taken after the step; every game that starts a fight inside a step emits `play.Engaged{At: fighting}` (grep `Engaged` in each game's combat start).
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
- #6 done (26df0d8, user: named confirmed slots; saving is the engine's): gobox keeps every game's slots through Save()/Load(), always asks a name and confirms an overwrite; play.SlotSaver, SaveNoConfirm, SaveUnnamed gone. play.Slots stays as a read-only import of the originals' own save files.
- #12 kept: classic.GameID only picks the game's own section of the one shared origins.json (the shared portrait collection); gobox never branches on it.
- Port bugs fixed: POR HitMessage skips the dead alone as ovr08:20CC (por ead2bae); BR1/BR2 list race above gender as their sheets (a2cfea5, 0b41cfe).
pod-go rules tests compile and pass again (fc28a16).
