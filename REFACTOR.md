# Engine refactor (drafting, 2026-11-13)

**Goal.** `engine` holds only the Gold Box game loop: rules flow, scripts, combat, camp, town, creation, quests, the canonical save. Everything else moves to where it belongs or is cut: UI to `ui`, one-time conversion to the importer, the adventure format to its own package, dev tools to `cmd/fixture` or tests. **How.** One item per commit, each builds (`go build ./... && go vet ./...`, also in the BR repos); targeted tests only; full suite once at the end. Tick items here as they land; this file is the handover.

## Decisions and answers

**engine vs game: keep both.** `game` is the contract (`game.Game`: State, Allowed, Do; plus the state, action and data types). It has three implementations: `engine` (the four pool games), `buckrogers-go/rules.Engine` and `buckrogers2-go/rules.Engine`. `ui` reads only `game`, so it never depends on one implementation. Dropping `game` would make `ui` and both BR engines import the pool engine. What does go: `game`'s non-contract parts (journal layout `WrapText`/`JournalBook.Layout`/`JournalPictureSize` → `ui`; `PointBuyCost`/`PointBuySpent` → `rules`; check `held.go`, `memberpage.go`).

**facts: cut it.** `known()` (playfacts.go) fills a 36-field `facts` struct (party, purse, shop, temple, combat, vault, picker, page, ...) on every rebuild; `build()` (play.go) then uses one or two of those fields to pick and return one `game.State`. It is an intermediate copy of the engine's state in `game` types, left over from when facts were a separate API for an older presenter. Nothing else reads it. Replacement: `build()` calls the per-mode builders (`placeState`, `itemsState`, `offers`, ...) for the state it returns. Next step after that, per "one authoritative state": the engine writes its `game.State` where it changes instead of rebuilding it.

**Old UI code not yet superseded** (the cut needs these first):
- `game.Page` (`Pane`, `Prompt`, `View` text runs and `Layers` of pictures): the engine's drawn screen is still in the contract, and `ui` draws it (combat reports `f.Page`, `ui/frame.go` timed pages). Fix: replace each use with typed facts (messages, the report's lines, the picture by ID) and let `ui` lay them out; then `page.go` and `game.Page` go.
- Picker: `ui/picker.go`, `poolpick.go`, `parts.go`, `act.go`, `game.go`, `rules.go` call the engine's `OpenPicker`/`PickerView`, which lay out columns, rows and pages. Fix: `game.Picker` lists the candidates (ID, name, picture); `ui` lays out the grid and pager; `pickgrid.go` goes.
- `engine.MapCell` reaches `ui` as `game.Data.MapCell` (pixels per map square, set by `importer/games/por_gamedata.go`). Fix: a `ui` constant; drop the field.
- Not yet checked: `snapshot.go`'s other 100 uses, `ccursor`/`cview` (combat cursor/view corner: engine-only, 19 uses, likely safe to cut once `ui` keeps its own cursor).

## Checklist

- [ ] F1 Adventure format out of `engine` into its own package (`GameTables`, `ArtNames`, file and dir names, `WriteGameData`/`ReadGameData`/`ReadGameTables`, the data types `importer/games` uses); corerules keeps only `Ruleset`, renamed (rulesetseam.go)
- [ ] F2 `importer/games` merges into `importer` (no `engine` import left; its 13 tests that boot `engine.NewPool` move to `engine` or `cmd/gobox`)
- [ ] S1 Original save import (`saveimport_engine.go`, `importSave`, `importSlots`, `origSaveFile`) → importer, at import time; `legacySave` migration cut
- [ ] S2 `SaveEditor` (savefile.go) → `cmd/fixture`
- [ ] A1 Run-time art conversion (`tilePix`, `pngImage`/iconpng.go, the `tiles`/`grounds`/`pngs`/`iconPoses` caches, pic.go symbol sets and colour fixes) → importer and `asset`: the folder holds finished images
- [ ] A2 Sound: keep "which sound/song plays"; check `sound.go`'s driver calls against `audio` and move what is `audio`'s
- [ ] P1 POR-named files (`porpics.go`, `porcreation.go`, `soundtrack.go`, `PorArt`, `PorMusic`, `PorMapLabels`): check each one's users; rename and unify what every pool game uses, make game-only parts adventure data
- [ ] D1 `devviews_engine.go` → build tag or test helper
- [ ] U1 `game.Page` replaced by typed facts in `ui`; `page.go` cut
- [ ] U2 Picker layout to `ui`; `pickgrid.go` cut
- [ ] U3 `snapshot.go`, `TextPaneRows`, `MapCell` cut
- [ ] U4 `ccursor`, `cview` cut
- [ ] G1 `facts`/`known()` cut; `build()` calls the per-mode builders
- [ ] G2 `game`'s non-contract parts out (journal layout → `ui`, point-buy → `rules`)
- [ ] Final: full build/vet/test, all repos
