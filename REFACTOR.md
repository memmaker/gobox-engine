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

**Buck Rogers: data folders only** (Felix, 2026-11-13). BR1 and BR2 stay supported, but like the pool games: `data/br1`, `data/br2` adventure folders made by gobox's importer and played by the one `engine`. No own repos, no own binaries (`playbr1`, `playbr2`), no second and third `game.Game` implementation. Today they are ~40k lines in four repos: buckrogers-go 15k and buckrogers2-go 16k (each its own `rules.Engine`, importer, launcher), buckrogers-core 2.6k (shared BR engine parts: space, ships, combat map), rulesxxvc 6k (the XXVc ruleset, already behind the `Ruleset` seam). Direction: `importbr1`/`importbr2` → `importer`; rulesxxvc → a `rules` ruleset; what only BR has (space fight, ships, ship sheet) → `engine`, switched on by adventure data and the ruleset, never by game name; the two BR engines go. A project of its own, after this refactor (B1 below). The T2 rows for BR types (`Blast`, `Blink`, `SpaceFight`, `ShipSheet`) are decided as part of it.

## Checklist

- [ ] B1 Buck Rogers as data folders (above): plan it as its own checklist once this refactor is done; until then the BR repos keep building against gobox

- [x] T2 Audit `game` for every presentation field (sizes, cells, positions, colours, drawn text runs); list them here, each with its `ui` replacement
- [x] T1 Test: `ui` does not depend on `gobox/engine` (`go list -deps`), beside TestGameImports
- [x] F1 Adventure format out of `engine` into its own package (`GameTables`, `ArtNames`, file and dir names, `WriteGameData`/`ReadGameData`/`ReadGameTables`, the data types `importer/games` uses); corerules keeps only `Ruleset`, renamed (rulesetseam.go)
- [x] F2 `importer/games` merges into `importer` (no `engine` import left; its 13 tests that boot `engine.NewPool` move to `engine` or `cmd/gobox`)
- [x] F3 `engine` no longer imports `importer` (title.go, ending_engine.go read the manifest's `Sequence`, `TitleScene`, `TitleData`): that format moves to `adventure`; a test forbids `engine` → `importer`. pod/sotsb/cotab's own `*DungeonTiles` use `adventure.DungeonTiles`
- [x] S1 Original save import (`saveimport_engine.go`, `importSave`, `importSlots`, `origSaveFile`) → importer, at import time; `legacySave` migration cut
- [x] S2 `SaveEditor` (savefile.go) stays in `engine` behind `//go:build fixture` (it is the engine's save and script-field logic: moving it would export ~10 internals); `cmd/fixture` builds with `-tags fixture`, and `go vet -tags fixture ./...` joins the checks so the tagged file cannot rot. After S1, which changes `readSlotSave`.
- [x] A1 (partly: Amiga own colours into `asset.Block.Image`; PNG decoding and icon cut into `asset`. Remainder goes with U1: combat tiles (`tilePix`/`tiles`/`groundOf`/`grounds`) become picture IDs in `game.Combat` that `ui` loads, then `ArtFormat` bump and re-import; the `pngs`/`iconPoses` caches go once `ui` gets raw PNG bytes; check whether symbol-set masks can be applied at import) Run-time art conversion (`tilePix`, `pngImage`/iconpng.go, the `tiles`/`grounds`/`pngs`/`iconPoses` caches, pic.go symbol sets and colour fixes) → importer and `asset`: the folder holds finished images
- [x] A2 Sound: keep "which sound/song plays"; check `sound.go`'s driver calls against `audio` and move what is `audio`'s
- [x] P1 POR-named files (`porpics.go`, `porcreation.go`, `soundtrack.go`, `PorArt`, `PorMusic`, `PorMapLabels`): check each one's users; rename and unify what every pool game uses, make game-only parts adventure data
- [x] D1 `devviews_engine.go` and the `-dev` wiring are left out of a release build by `-tags nodev` (opt-out, so the tests that use the scenes stay in the default run); `go build -tags nodev ./...` joins the checks
- [x] U1 `game.Page` replaced by typed facts in `ui`; `page.go` cut
- [x] U2 The engine's duplicate picker layout cut (`pickgrid.go` keeps only filling `game.Picker` and taking the pick)
- [x] U3 `snapshot.go`, `TextPaneRows`, `MapCell` cut
- [x] U4 `ccursor`, `cview` cut
- [ ] G1 `facts`/`known()` removed
- [ ] G2 One state: the engine keeps the player-visible state in `game` types and writes it where it changes; `rebuild()`/`build()` and the per-mode copies go
- [x] G3 `game`'s non-contract parts out: journal layout → `ui`, point-buy maths → `rules`, `held.go`/`memberpage.go` checked
- [ ] N1 Shared font pool: the importer adds each game's fonts to one pool; `ui` cycles through all; `Data.Font`/`Fonts`, `TextRun.Upcase` leave `game` (ruling 2)
- [ ] N2 Title, ending and story screens as declarative scripts plus exported picture elements; `ui` composes them (ruling 3); `Layer` leaves `game`
- [x] N3 The importer cuts the menu logo into a finished image; `Data.Logo` goes (ruling 10)
- [x] Y1 Ponytail audit (2026-11-13): delete production functions only tests call (~140, ~1.5–2k lines; list: deadcode without -test, minus `fixture/`), with the tests that only test them; true test helpers move to `_test.go`
- [x] Y2 Delete the 5 functions nothing calls (`ecl.NeedsPlainTests`, `importer.CotabScriptLayout`, `SotsbScriptLayout`, `ImportAssets`, `rules.SymbolBlock`)
- [x] Y3 Code only the dev scenes reach since D1 (`firstOkay`, `okay`, `blockedRet`, `fileBytes`, `heldOf`, `entryNamed`, `invUp`, `fullPicture`, `firstMonster`, `ecl.ArgsTrace`): delete or move into devviews_engine.go
- [x] Y4 (nothing folded: `ui/input` would cycle through `ui/combat`; `idgen`, `winicon` are go:generate tools) Fold the one-file packages `ui/input`, `rules1e/idgen`, `cmd/gobox/winicon` into their parents where no import cycle forces them
- [ ] E1 `engine` stops importing `asset`/`asset/orig`: pictures leave as IDs (after U1, N2, the A1 remainder); colour constants become the ruling-1 colour names; the remaining original-data reads (`orig.NewLoader` in dataimport.go, `orig.Monster` records, `orig.RulesFile`) move to the importer's output. A test forbids it
- [ ] E2 `engine` stops importing `audio`: it publishes which song/cue/voice line plays as typed facts in `game`; `ui` (host) drives `audio` (sound.go's `audio.Driver` use, voice.go's Vorbis decoding). A test forbids it
- [x] R1 `cmd/gobox` stops importing `rules1e` (`rules1e.Settings{PathSearch}`, `rules1e.ReadRoster`): behind `rules`/`engine`
- [x] R3 `rules4e` imports `engine` (so `engine` cannot pick the 4e combat rules itself and `cmd/gobox` hands it in): a ruleset must depend on `rules` only; then `cmd/gobox` passes a `Combat4e` setting
- [ ] R2 `engine` reaches the rules only through `rules` (today ~600 uses of `rules1e` types: `Character`, `Item`, classes, slots, directions). Needs its own plan: a ruleset-neutral character and item model in `rules`
- [ ] D2 `game` and `adventure` both define `Sequence`, `Layer`, `Credit`: one canonical definition (with N2)
- [x] K1 `TestKeyDispatch` fails since the app → ui merge: `ui/shell.go` (`ScriptTimed`) and `ui/shellfront.go` (F12, Shift-F12) name key codes instead of dispatching through the key table
- [x] K2 `SOUND.OFF` is read at start (engine/sound.go) but nothing writes it since `soundToggle` was cut as unreachable: drop the read or wire the toggle
- [ ] C1 Credits: `Credit.X/Y/Colour`, `Data.CreditBoxes`, `CreditGaps` (original two-column credits; ruling needed)
- [ ] U5 Per-game values U1 flattened to one ui constant (cotab ready mark, door colour, Amiga combat ground grey): restore the originals' values
- [x] U6 `ui` does not scale `StepMs`/`Ask.Ms` by the player's speed (ruling 5); `Walls.FloorRGBA/CeilRGBA` still RGBA, not a palette index (ruling 7)
- [x] L1 Load regression (S1 cut `legacySave`): every shipped/dev slot with Area/Game/Purse at the top of GAME.JSON was refused ("unexpected end of JSON input"). Fixed on disk once (12 por dev slots under data/por/sources/save/saves; backups in the session scratchpad). Bug-class check: a test that loads every slot under data/*/sources/save/saves; also `loadFrom` drops gob decode errors of MAPS/VOICED/BESTIARY/... (golden rule). sotsb home `town.sav` holds an original SAVGAMA.DAT slot (no GAME.JSON): convert it with the importer's save conversion
- [ ] Final: full build/vet/test, all repos

## T2 audit: presentation fields in `game`

`game/game.go:4` ("no colours, cells, rects, inks or positions") and the `Data` doc ("no drawing, no positions") are false today. P = layout/ink, leaves `game`; T = presentation timing; D = domain, keeps.

**A. Pages and layers (U1).** `Page` (Pane, Prompt, View, Layers) P → typed facts (`Common.Text`, `Report.Text`, a typed prompt, `Scene.Text`, `Common.Picture`/`Scene.Pictures`); `ui` builds the page (`ui/view_state.go:90` already converts back). `Page.PaneRows` P → drop, `ui` has `maxPaneRows`. `TextRun.X/Fg/Bg/Upcase` P → drop; `ui` maps `Tone` to colour. `TextRun.Blink` T → a domain `Emphasis` flag. `Layer.At` P → `ui` places pictures. `Layer.Frames/FrameMs` T → frames are asset data on `Picture`, `ui` picks speed. `Layer.MinParty` D → onto a typed `Scene` step; `MsgDelayed` T → drop. `Common/Scene/TitleStep.Layers`, `Data.TitleBackdrop` → `[]Picture` (`Scene.Layers` duplicates `Scene.Pictures`).

**B. Cells, pixels, inks.** `Data.MapCell` → `ui` constant (10). `Picker.Cell` → `ui` derives from picture W/H. `SheetPage.PortraitAt` → drop. `Results.EndAt` → drop (no `ui` reader; likely dead). `Scene.Placed`/`Placed.At` → ordered text lines, keep `Tone`. `Credit.X/Y/Colour`, `Data.CreditBoxes/CreditGaps` → drop, `ui` flows credits. `Data.LabelInks`, `DoorInk`, `SelectInk` → drop, `ui` maps `Label.Kind`. `Data.Logo` → importer cuts the logo into a finished image (A1). `Data.Font/Fonts` (`Typeface`), `Data.Border`, `Data.Palette` → asset data shipped by the importer.

**C. Timing, cursor, view.** `Report.Ms/StepMs/Pose`, `Shot/CastSpell.StepMs`, `Fell.Ms` T → drop; `ui` has `flightMs`/`minStepMs`; `Pose` → an acting fact. `Report.Page` → as `Page`; `Report.Screen` D keeps. `Ask.Ms`/`Question.Ms` → typed "no key needed" if a real wait, else drop. `Combat.Cursor/CursorOn/CursorAt/View` → drop with U4, `ui` owns cursor and scroll. `Combat.Floor`/`Battlefield.Floor` colour → drop, keep `FloorTile`. `Blast.At` → target system/side; `Blink` facts keep, timing doc goes. `SpaceFight.Pause/StatusLine`, `ShipSheet`: keep, drop row/col wording from docs. `PaceAnim`, `Note.At` → check; `ui` stamps notes.

**D. Layout functions.** Journal layout → `ui` (G3). `Read.Lines` marker bytes (`\x00` title, `\x01` picture, `\x02` rows) → typed segments (rule 6).

**Domain, keep:** map coordinates and sizes, picture IDs/pixels, names and IDs, `Walls` names and map facts, `Intro.Pictures`, sounds and music, hotkeys and bindings, `Tone`, `Texts`.

**Rulings (Felix, 2026-11-13):**
1. Text colours: text keeps its original colour as an EGA colour name (a typed enum in `game`), not raw ink numbers; `ui` maps names to its palette.
2. Fonts: the importer adds each game's fonts to one shared font pool; `ui` lets the player cycle through all of them. Upper-case glyphs are a property of a font's glyph table; `TextRun.Upcase` goes.
3. Title, ending and story screens: the importer decodes them into declarative scripts (steps referring to picture elements exported separately); `ui` composes the elements from the script, reproducing the original flow and look. `Layer.At` leaves `game`.
4. Text over the 3D view: the engine keeps the original line wrapping (lines in order); `ui` decides placement. `Placed.At` goes.
5. Timed holds: holds from the original stay as typed facts with the original duration (`ui` scales them by the player's speed); gobox-invented holds (`Fell.Ms`/`fallHoldMs`) go. Details below.
6. Blink: kept as a text property (the originals use it); `ui` draws it.
7. A map's own flat floor/ceiling colour stays as a palette index in the map data; the combat fallback colour goes.
8. `Data.ReadyMark` goes; `ui` owns the glyph.
9. `IconEditor.Cells` goes; colours and choices stay.
10. The importer cuts the menu logo into a finished image; `Data.Logo` goes.
Settled from the code: `Results.EndAt` goes (no `ui` reader); `Data.Palette` moves to the art data.

**Decided without Felix (he was away, 2026-11-13; revisit if wrong):**
- C1 credits: the credit colour stays as an `Ink`; the importer writes the credits as an N2 declarative script carrying the column placement; the `Credit`/`CreditBoxes`/`CreditGaps` fields leave `game` with N2.
- U5 per-game look values (cotab ready mark, door colour, Amiga bare-ground grey): shipped as data in each adventure folder that `ui` reads without knowing the game (gobox stays game-agnostic and the originals' look comes back).
- K2 `SOUND.OFF`: the read goes (nothing writes it; the sound on/off is the host setting).

**5. Timed holds, what each is:** `Shot.StepMs`/`Report.StepMs`: the original's missile pace per projectile (10 pointed, 20 tumbling, 50 spinning, 30 a spell; BR O08F:2B26) — needed for accurate playback. `Ask.Ms`: the script pause op (BR op 58: game speed × 100 ms; pool from the script wait) — a script event, needed. `Fell.Ms`: `fallHoldMs = 500`, gobox's own replay hold — not original. BR space fight `Report.Ms`: unchecked. Recommendation: keep the original-sourced ones as typed facts (the original duration, `ui` scales by the player's speed), drop gobox-invented holds (`Fell.Ms`).
