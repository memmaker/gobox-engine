# Handover: one canonical game contract, passive games (2026-11-06)

## Goal (user decisions)

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
| SOTSB | `e60abcf` (dead key layer gone); `0d17fae` (dead create/modify/slot/walk/title/sheet code gone, −965 lines; creation lives on via Draft); wait-based WIP parked on `refs/wip/play-port` (reference only) | `8e41bd3` ECL VM passive (POD shape); next: align with POD, then flows a–d as POD, flows, native play.Game; fix the adapter bugs: Ready/Unready, temple Heal, PickMember replies, Number ask, EndIcon Keep, EvSong→MusicChange | `sotsb-go/.claude-progress-play.txt` |
| POD | `cc86f49` (yes/no wrappers merged); `b68f757` (ECL VM passive: frames stack + pending kinds, the reference shape for SOTSB); wait-based patch in the session scratchpad (`waitbased-port.patch`, reference only) | flows in progress: temple Heal fix, `rolled`, Party/shop/vault/temple/camp/Combat/AfterCombat without funcs, flow.go asks; then native play.Game | `pod-go/.claude-progress-play.txt` |
| POR | `e8cf16f` (Bar/Choice/Entry/Pick/EventKind gone) | RunECL frame stack (11 sites); settle/closed/opens and `es.hold` → screen stack; delete `wait(...)` and 33 `then` params; native play.Game; delete GameRules, reply()/gameAsk(), PickMember/PickSpell/PickClass/Confirm | `por-go/.claude-progress-play.txt` |
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
