# Unseen questions: checklist

## Goal
A question the game asks (`game.Common.Ask`, a draft's, a Refusal's) must never be
answered without being on screen. The Red Sonja bug: the hall's "Pick New Class:" was
not drawn on the party screen, and the next key answered it with reply 0.

## Cause
Two places decided independently: `askOf` (ui/shown.go) drew a question only on a
whitelist of modes; key handling turned Enter / any key into `game.Reply` regardless.

## How
One predicate says who shows the pending question (`frontState.claimed`: a screen of
its own; else the generic question window). Drawing and answering both go by it.

- [x] Fix the instance: party screen draws a non-menu question (92ab2408)
- [x] B: an answer (ActOK, ActChoose, ActYes, ActNo, ActText) is dropped while the
      game's question is unseen (`Presenter.unseen`), in `Presenter.send`
- [x] D: invariant test over every dev scene: no state with an unseen question
- [x] C: `askOf` draws every question no screen claims (no mode whitelist)
- [x] Full suites: ui, cmd/gobox pass after C; engine last run before B (B, D, C touch ui only). gate.sh not usable: "games[@]: unbound variable" (finds no game repo)

## Round 2: close the gaps
- [x] 1: Enter / "any key" (ActOK) never chooses: Game.Answer drops it on a yes/no or
      a choice of several (TestOKNeverChooses). No exemption for "press a key" needed.
- [x] 2: "seen" is a fact, not a promise: drawing a menu records the question it
      showed this frame; `unseen` compares the game's question with the one recorded
- [x] 3: every headless frame is checked (Front.Draw: Presenter.stuck -> ui.OnStuck; the
      TestMain of ui and cmd/gobox fail the run on one; GOBOX_TRACE logs it in play);
      TestNoUnseenQuestion walks por, cotab, sotsb, pod (ui, cmd/gobox suites pass)

## Open
- pod dev scenes that panic before a frame, not walked (unseen_test.go sceneBroken):
  wildcombat ("rules keep no world word 58", Engine.tileBlock), journal (POR's tale).
- "Shown" is "this frame drew choices", not "drew this very question".
- buckrogers (ShipFight) and other repos' games run no frame of these checks here.

## Detect the class elsewhere
`go test ./ui -run TestNoUnseenQuestion`; any test that draws frames fails on a
stuck one. In play: `GOBOX_TRACE=1` prints "stuck: <state> asks <question> on no screen".
