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


### R handover (agent B6, 2026-11-11)
- [x] R6 (first part): `importer.Read/List` (importer/folder.go: the one folder reader, over importer.Manifest; valid = Format, ID = folder name, ruleset); `cmd/gobox` (own module, gobox/cmd/gobox): `gobox import ...` = importer.Main, else lists adventures/ and plays one (`-adventures`, `-data`, host and debug flags, app.Flags); exe icon (icon.ico, syso, ./winicon moved from por-go); `app.Start` (what every engine gets); `present.HostOptions`/`HostOptionRows` (from cmd/por/host.go); por-go `playpor.Play` (cmd/por's body); `rules.New(adv, data)`: scripts/art/data from the adventure folder, no key.txt check, no POOL.CFG read (defaults E P F = POR's file; the import no longer copies it, dataFormat 4); cmd/por keeps only -setup / -import-disk2 (gameimport); run.sh builds gobox, imports, plays.
- Open in R6: (1) cmd/gobox picks the engine by manifest ID (`players`, por only) until R4/R5; then by Ruleset. (2) Manifest.Start/End is not read: importpor sets none and POR's title is still built in rules/play.go, title.go; convert importer.Sequence to the play types when A8 fills it. (3) host dir `-data` still holds speech/, fonts/, dasp/, soundtrack/, amiga/SOUND.BIN, amiga/TITLE.PNG: the Amiga ones are import output and belong in the folder (A8). (4) data/pictures.json (A9, sprites.ReadPictures) not yet read by playpor/the engine. (5) js build: fails in gobox/gpu (gogpu-wgpu hal/allbackends has no js files); pre-existing: por-go c70cbee's cmd/por fails the same way (scratch export).
- [x] Bug-class sweep (B5's eclOps): scanner (B6 scratchpad initscan: package vars written by more than one init(), through `x := &v` aliases and called package funcs) over por-go, cotab, sotsb, pod, buckrogers-go, buckrogers2-go, buckrogers-core, rulesxxvc, rules1e, gobox: no hits. Cross-package: importer.Register's table is order-free (IDs sorted, duplicates panic).
- Next: R4 after C6 (C6 committed por-go 6afc10e: ask main to sequence).

### Note for the engine loop (coordinator, 2026-11-11)
core.FightTerms fields the engine fills at fight setup: `PartyFacing` (X6: the party's map facing); `Area`, `RevivesParty`, `NoItemLoot` from BR's named script variables areaKind, revivesParty, skipsCombatItems (A10 names them in the BR script layout). POR leaves Area/RevivesParty/NoItemLoot zero; it fills PartyFacing if its rules read it.
- After R4/R5 (A10): once BR runs on cmd/gobox, BR's fight setup fills FightTerms from the named BR script variables RevivesParty (V_7EE6), SkipsCombatItems (V_7EE3), AreaKind (V_4BE6; $A2/$A8 = wilderness), and the party's facing.

### R handover (agent B7, 2026-11-11)
- [x] R6 1d: TestHostOptionsSettingsJSON (front/present/hostsettings_test.go; gobox 2a0f50a)
- [x] R6 1b: rules.New reads data/pictures.json into Engine.pics (por-go 61ecdc9). Bug fixed: ImportDataTo's RemoveAll deleted importpor's pictures.json. ImportDataTo now clears nothing. Bug-class check: an import step that clears a directory that other steps write into (A10: importer.CheckContents / TestImportContents).
- [ ] 1c: sent main the list of files to move (SOUND.BIN, dasp, speech, soundtrack, TITLE.PNG; read sites sound.go:225/229, voice.go:83, soundtrack.go:26, title.go:113). Switch those reads to e.adv once A10's import writes them.
- [ ] 1a: blocked. Nothing fills or plays importer.Sequence. Proposed to main: a gobox converter Sequence -> play.Scene; POR's engine-driven title (TitleFlash/Credits/Done, logo cut) comes out once importpor fills Start. Waiting for main's OK.
- [ ] 2 (Lost features), 3 (R4/R5): not started.
