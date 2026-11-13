# Engine refactor (drafting, 2026-11-13)

**Goal.** `engine` holds only the Gold Box game loop: rules flow, scripts, combat, camp, town, creation, quests, the canonical save. Everything else moves to where it belongs or is cut: UI to `ui`, one-time conversion to the importer, the adventure format to its own package, dev tools to `cmd/fixture` or tests. **How.** One item per commit, each builds (`go build ./... && go vet ./...`, also in the BR repos); targeted tests only; full suite once at the end. Tick items here as they land; this file is the handover.

## Decisions and answers

**One state, written in `game` types.** `game` is the contract, so `engine` works on it directly: the engine's state that the player sees lives in `game` types and the engine writes it where it changes. No second copy, no `facts`, no `rebuild()`/`build()` translation (play.go, playfacts.go), no per-mode copy functions. `ui` reads that state.

**`ui` never reads old UI code.** Old UI code is anything that says how something is drawn and comes from a game: drawn pages (`game.Page`, `TextRun`, `Layer`), pixel and cell sizes (`MapCell`, `Picker.Cell`, `PaneRows`), screen positions. `game` carries domain facts only, as its own doc says; `ui` decides every layout. Each such field leaves `game`, and `ui` computes it.

**`ui` never imports `engine`.** Today it does not (`go list -deps ./ui` has no `gobox/engine`); a test forbids it from now on.

**engine vs game: keep both.** `game` is the contract (`game.Game`: State, Allowed, Do; plus the state, action and data types). It has three implementations: `engine` (the four pool games), `buckrogers-go/rules.Engine` and `buckrogers2-go/rules.Engine`. `ui` reads only `game`, so it never depends on one implementation. Dropping `game` would make `ui` and both BR engines import the pool engine. What does go: `game`'s non-contract parts: journal layout (`WrapText`, `JournalBook.Layout`, `JournalPictureSize`) → `ui`; point-buy maths (`PointBuyCost`, `PointBuySpent`) → `rules`; check `held.go`, `memberpage.go` the same way.

**facts: removed.** `known()` (playfacts.go) fills a 36-field `facts` struct (party, purse, shop, temple, combat, vault, picker, page, ...) on every rebuild; `build()` (play.go) then uses one or two of those fields to pick and return one `game.State`. It is an intermediate copy of the engine's state in `game` types, left over from when facts were a separate API for an older presenter. Nothing else reads it. It goes with the one-state change above.

**All old UI code goes** (screen drawing, snapshot, picker layout, combat cursor and view, pixel and row sizes), replaced in `ui` where the player still needs it. Where `ui` still reads it today (the fix lands with the cut):
- `game.Page` (`Pane`, `Prompt`, `View` text runs and `Layers` of pictures): the engine's drawn screen is still in the contract, and `ui` draws it (combat reports `f.Page`, `ui/frame.go` timed pages). Fix: replace each use with typed facts (messages, the report's lines, the picture by ID) and let `ui` lay them out; then `page.go` and `game.Page` go.
- Picker: not a `ui` dependency. `ui` has its own `OpenPicker`/`PickerView` (ui/game.go): it sends the `game.OpenPicker` action and lays the grid out itself from `game.Picker`. The engine's `pickgrid.go` still keeps a second layout (`layout`, `perPage`, `pages`: columns, rows, pages; `iconRect` pixel rects). Fix: the engine fills `game.Picker` with the candidates only; its layout code goes.
- `engine.MapCell` reaches `ui` as `game.Data.MapCell` (pixels per map square, set by `importer/games/por_gamedata.go`). Fix: a `ui` constant; drop the field.
- Not yet checked: `snapshot.go`'s other 100 uses, `ccursor`/`cview` (combat cursor/view corner: engine-only, 19 uses, likely safe to cut once `ui` keeps its own cursor).

## Checklist

- [ ] T2 Audit `game` for every presentation field (sizes, cells, positions, colours, drawn text runs); list them here, each with its `ui` replacement
- [x] T1 Test: `ui` does not depend on `gobox/engine` (`go list -deps`), beside TestGameImports
- [ ] F1 Adventure format out of `engine` into its own package (`GameTables`, `ArtNames`, file and dir names, `WriteGameData`/`ReadGameData`/`ReadGameTables`, the data types `importer/games` uses); corerules keeps only `Ruleset`, renamed (rulesetseam.go)
- [ ] F2 `importer/games` merges into `importer` (no `engine` import left; its 13 tests that boot `engine.NewPool` move to `engine` or `cmd/gobox`)
- [ ] S1 Original save import (`saveimport_engine.go`, `importSave`, `importSlots`, `origSaveFile`) → importer, at import time; `legacySave` migration cut
- [ ] S2 `SaveEditor` (savefile.go) stays in `engine` behind `//go:build fixture` (it is the engine's save and script-field logic: moving it would export ~10 internals); `cmd/fixture` builds with `-tags fixture`, and `go vet -tags fixture ./...` joins the checks so the tagged file cannot rot. After S1, which changes `readSlotSave`.
- [ ] A1 Run-time art conversion (`tilePix`, `pngImage`/iconpng.go, the `tiles`/`grounds`/`pngs`/`iconPoses` caches, pic.go symbol sets and colour fixes) → importer and `asset`: the folder holds finished images
- [ ] A2 Sound: keep "which sound/song plays"; check `sound.go`'s driver calls against `audio` and move what is `audio`'s
- [ ] P1 POR-named files (`porpics.go`, `porcreation.go`, `soundtrack.go`, `PorArt`, `PorMusic`, `PorMapLabels`): check each one's users; rename and unify what every pool game uses, make game-only parts adventure data
- [ ] D1 `devviews_engine.go` → build tag or test helper
- [ ] U1 `game.Page` replaced by typed facts in `ui`; `page.go` cut
- [ ] U2 The engine's duplicate picker layout cut (`pickgrid.go` keeps only filling `game.Picker` and taking the pick)
- [ ] U3 `snapshot.go`, `TextPaneRows`, `MapCell` cut
- [ ] U4 `ccursor`, `cview` cut
- [ ] G1 `facts`/`known()` removed
- [ ] G2 One state: the engine keeps the player-visible state in `game` types and writes it where it changes; `rebuild()`/`build()` and the per-mode copies go
- [ ] G3 `game`'s non-contract parts out: journal layout → `ui`, point-buy maths → `rules`, `held.go`/`memberpage.go` checked
- [ ] Final: full build/vet/test, all repos
