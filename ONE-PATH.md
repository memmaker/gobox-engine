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
- Auto-walk uses POR's rules for all games: stops on an unvisited square with a
  script, on any key or click, and locked doors count as walls.
- Paladin powers are offered wherever spell casting is, as POR does.

## A: game side
- [ ] por-go: exit hints; listAnswered has no string comparison
- [ ] pod-go: exit hints; paladin powers, Cure Disease reachable
- [ ] cotab-go: exit hints; paladin powers audited
- [ ] sotsb-go: exit hints; paladin powers audited
- [ ] buckrogers-go: exit hints
- [ ] buckrogers2-go: exit hints

## B: gobox
- [ ] Walk: POR's stop rules and locked doors, with tests
- [ ] classic front reads play.State, sends play.Action
- [ ] playrules, game.Rules, game.State, game.Decision deleted
- [ ] Game cmd wiring and tests moved; parity guard against the old path
- [ ] Final: full suites and scripted windowed runs of all six games
