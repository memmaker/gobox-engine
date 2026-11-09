# One path: the front on play.Game, game.Rules and the adapter gone

Goal (Felix): one code path and a clear, narrow interface. Games are passive
play.Game; gobox's front reads play.State and sends play.Action directly.
How: game-side rules first (agent A), then gobox's walk and front (agent B).

Agreed rules:
- Esc on an optional menu always closes it or backs out to the upper UI level.
  Engine-built menus back out. Script menus: a typed hint names the exit row
  (play.Question.Esc); no hint: Esc selects nothing, the choice is forced.
  The script decompiler writes the hint into the scripts by a heuristic; the
  engine reads only the typed hint, never text.
- gobox has one script decompiler, run at import time for every game; game
  differences are explicit decompilation/data versions.
  Every game interprets the decompiled script directly: no bytecode at runtime.
- Vendor menus (shops, temples, training halls, taverns...): the decompiler folds
  the Yes/No gate and the real menu into one. A folded menu's exit always runs the
  cut No branch (such as MustLeave).
- Auto-walk uses POR's rules for all games: stops on an unvisited square with a
  script, on any key or click, and locked doors count as walls.
- Paladin powers are offered wherever spell casting is, as POR does.

## A: game side
- [x] por-go: exit hints; listAnswered has no string comparison
- [ ] pod-go: exit hints; paladin powers, Cure Disease reachable
- [ ] cotab-go: exit hints; paladin powers audited
- [ ] sotsb-go: exit hints; paladin powers audited
- [ ] buckrogers-go: exit hints
- [ ] buckrogers2-go: exit hints
- [ ] Decompiler folds vendor gates, per game, with comparison tests

## B: gobox
- [x] Walk: POR's stop rules and locked doors, with tests
- [x] classic front reads play.State, sends play.Action
- [x] playrules, game.Rules, game.State, game.Decision deleted
- [x] Game cmd wiring and tests moved; parity guard against the old path
- [x] Final: full suites and scripted windowed runs of all six games

## Progress (B)
- Walk: POR's rules in gobox (walkStep): text stops on a square not in Automap.Walked; a visited square's text does not; blank text never; any other fact (picture, damage, gain) does; questions, doors, off-course steps do. Locked doors are walls. A square per 150 ms. Any key or click stops it (Front.stopWalk, TestAnyInputStopsWalk). Not carried: POR stopped on ECL ops that show nothing (the op table); a silent script keeps walking here.
- Step 1 (0faee48): playrules folded into classic/rules.go; NewGame takes the play.Game. NOT the goal yet: rules.State() still builds a game.State for the builders, and rules.Apply takes game.Decision. That adapter is what step 2 removes.
- Step 2 started: leaf builders take narrow args (98d61ba); walk no-text stops tested (678007d). Remaining: Party, Purse, Loot, Page, Flags, Camp, Sheet, Ask, Battle, Creation, Inventory, Picker, Map; rules.State(), rules.Apply; then game.State, game.Decision and the games.

- One Member: play.Member and the per-game and front member() copies deleted; game.Member is the shared type (member() had dropped Powers for every game, inventory showed none). POR title regression fixed (bb7ab15). Final: all six fights seen in a window; full pass ok.
- Step 3: View reads play.State (view_state.go); game.State, game.Decision and all verbs deleted; Option, Ask, Refusal, Turn, Combat, Result, Report in play (play/shared.go); front-only actions play.Cancel, Aim, OpenItems, SpeedMenu; apply + private rules methods replace Apply and action(); act replaces decision. Guard extended; planted once.

## Progress (decompiler)
- POR fold pass (por-go 8035c30): rules/scriptfold.go, in cmd/eclscript. 4 shops, temple, training hall folded; a vendor's exit runs the cut No. Inn (pays before camp) and tavern gamble (no vendor menu) stay gates. Checked structurally only (TestVendorGatesFolded, TestScriptRoundTrip), not by running gate-No against back-out.
- Open: the decompiler is still in por-go/rules, POR's hand fixes are not patches yet, the other five games still read ECL*.DAX.
- a1a93d6: all classic builders take narrow args. 061cb98: play.PowerKind. Next: ViewOf reads play.State directly, rules.State()/Apply gone, Option/Ask/Refusal into play, game.State/Decision deleted (one commit stack). Another agent owns ecl/ and script loading.
- ecl/script (agent): the decompiler lives in gobox/ecl/script with typed versions V1 (POR), V2 (COTAB, SotSB), V3 (POD), V4 (BR1, BR2); POR hand fixes are import patches, byte-equal by TestImportEqualsCheckedIn; all six games decode and parse. Hints: POR 117, COTAB 65, SotSB 72, POD 13, BR1 45, BR2 49. Folds: POR 7, COTAB 2, SotSB 2, POD 0, BR 0.

Progress (round trip): full-operand Write/Parse survey of all six games: POR, COTAB, SotSB exact; POD 17, BR1 15, BR2 33 areas differ only by appended event stubs (Gotos landing on the original event commands). Folded gates keep their section's Text("") clear. Plain-test fallback: POD 49; BR1 65; BR2 36, 96, 97, 98, 114.
- Step 2/3 done (agent): all six games run decompiled text through script.Library (imported on first use; POR's into its scripts dir with names and patches, the others into the user cache); no game reads ECL byte code and gobox/ecl and brcore have no byte path. TestScriptEqualsByteCode holds 121,925 commands to the byte code (unfolded), cotab's TestGateComparedWithFold the folds. Gates fold when the No ends the vendor's tail or the tail jumps to the No (COTAB area 1, SotSB area 16). POR's area texts are no longer checked in.

## Next: three components (Felix)
gobox (engine) owns the main loop, presentation and UI. Rules hold a ruleset
(1st edition D&D, 2nd edition, ...). Adventure holds a module's real data
(scripts, maps, NPCs, assets). Rules and adventure know nothing of the other
components and are leaves of the dependency tree. The engine ties them together
through two literal Go interfaces, Rules and Adventure, and knows only those, so
any adventure runs on any conforming ruleset.

Settled:
- A `core` module defines the types the other modules need, including the neutral
  ids. It imports nothing of ours. Rules, adventure and engine import only core.
- The script interpreter is the engine's. An adventure supplies scripts on demand.
- The engine runs combat: loop, board, input, presentation. Rules answer who acts
  when, what actions are allowed and how they resolve. Rules stay passive.
- An adventure references neutral typed concepts (ids in core). A ruleset handles
  them as it wants; a mismatch fails clearly at load.

## Queued (Felix accepted the recommendations)
- [ ] Chosen and Card move to a front-only view type
- [ ] COTAB's private steps become direct calls
- [ ] COTAB Keep through the front works
- [ ] SotSB SaveCharacter returns its error, SaveNamed fails on it
- [ ] Heuristic knows NON, if it behaves as NO in those menus
- [ ] POR NO hints read and reported: 0:605 2:566 18:453 14:441 29:602
- [ ] Cleanup: BR1 stray saves, POR TitlePicture, check BR1 flights' StepMs

## POR on gobox/ecl (done)

POR runs on ecl.VM; its duplicate interpreter is gone. Kept as overrides: Compare (day sky), GetTable (effect table). RETURN without GOSUB goes on (eclHost.Exit is empty), DIVIDE keeps no remainder (ecl.NoRemainder). Code bytes go through vm.CodeByte and SetCodeByte. brcore is now module ecl-br, package eclbr.
