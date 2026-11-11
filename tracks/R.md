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

### R handover (agent B5, 2026-11-11)
- [x] R2a per-file moves (por-go 5a53a10, scratchpad/mover: free decls of the RuleState files used only by the Engine go to `<file>_engine.go`, fixpoint)
- [x] R2b script VM, presenter and art types out of the RuleState files (por-go 899599b). Bug found: two `init()`s filled `eclOps`, order-dependent on file names; now one literal. Bug-class check: `grep -l '^func init' */rules/*.go`, then any package table written by more than one init.
- [x] R3a `core.MapPos`, `core.BoardPos` (gobox 983f6e2, por-go a68d611); every core action/effect square is BoardPos; por's presenter keeps image.Point behind `point`/`points` (present_engine.go). MapPos has no users yet.
- [x] R3b `gobox/core/goldbox` (money, RNG, dice, Roll, ability tables, Abs/Sign/Cut/Upper/RemoveSubstr) and `gobox/core/grid`; gobox/rules keeps the char.Stats rolls, Flow, MakeFileName (engine); users in por, cotab, pod, sotsb (gobox a5ab1d8; cotab 4c4afc7, pod f3d76b2, sotsb 63bc3f2, por-go 8f402d1)
At this point the RuleState files import only core, core/dnd, rules1e, stdlib, plus: `gobox/geo` (geo.go offMapSolid: geo.OnMap, 16x16; candidate `MapPos` method in core), `gobox/game` (create.go, C's), and `encoding/binary` (files.go, money.go, poolid.go: forbidden in leaves by core's TestLeavesNoRawBytes; the record codecs belong to the save/import side, not rules1e).
Next (R4): move the RuleState files into rules1e. The Engine embeds RuleState (~300 promoted names): rename per identifier with gopls (go.work across por-go and rules1e), exported where the Engine calls them; then rules1e's Rules implements core.Rules/CharacterRules/CombatRules and the Engine calls through them. R5, R6 not started (R6: adopt importer.Manifest and A8's importer.Sequence for Start/End; no second shape).

