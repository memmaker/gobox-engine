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
- [ ] classic front reads play.State, sends play.Action
- [ ] playrules, game.Rules, game.State, game.Decision deleted (playrules, game.Rules, game.Mapped, classic.Game.R, por wrapper done; game.State, game.Decision open)
- [x] Game cmd wiring and tests moved; parity guard against the old path
- [ ] Final: full suites and scripted windowed runs of all six games

## Progress (B)
- Walk: POR's rules in gobox (walkStep): text stops on a square not in Automap.Walked; a visited square's text does not; blank text never; any other fact (picture, damage, gain) does; questions, doors, off-course steps do. Locked doors are walls. A square per 150 ms. Any key or click stops it (Front.stopWalk, TestAnyInputStopsWalk). Not carried: POR stopped on ECL ops that show nothing (the op table); a silent script keeps walking here.
- Step 1 (0faee48): playrules folded into classic/rules.go; NewGame takes the play.Game. NOT the goal yet: rules.State() still builds a game.State for the builders, and rules.Apply takes game.Decision. That adapter is what step 2 removes.
- Step 2 started: leaf builders take narrow args (98d61ba); walk no-text stops tested (678007d). Remaining: Party, Purse, Loot, Page, Flags, Camp, Sheet, Ask, Battle, Creation, Inventory, Picker, Map; rules.State(), rules.Apply; then game.State, game.Decision and the games.
