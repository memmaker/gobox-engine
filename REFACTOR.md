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

### Order of work (Felix, 2026-11-15): R2, then G2 (b, c), then B1, then G3

Handover: each item below is one commit or one track (`foldwt.sh new <track>`, landed with
`foldwt.sh land <track>`, which gates). Sonnet agents take the mechanical items, one track each, on
disjoint files; their work is read and gated before it lands. Guards that keep rules and UI out of
the engine: `engine/importguard_test.go` (no ruleset import; a list of the files still to fix that may
only shrink), `game/noui_test.go` (no drawing package, no screen position), `ui` never imports
`engine` (T1). A new concern in the engine must fail one of them.

- [x] R2 `engine` reaches the rules only through `rules` (2026-11-15): no non-test file of `engine`
  imports a ruleset; `engine/importguard_test.go` forbids it, with no list of exceptions (the guard
  had been dead: it matched `rules1e`, the import is `gobox/rules1e`, and passed on 17 files).
  - [x] R2a rule names (track r2a): the saves, races and alignments come from `Rules.Catalog()`;
    the AD&D names the scripts still read are in `rules/adnd.go` (see R2f)
  - [x] R2b flight looks and the bestiary kind (track r2b): `rules.FlightFact`, `rules.BeastKindFact`
  - [x] R2c character files (track r2c): `rules.FileFormat`, a capability the ruleset has;
    `rules.CharacterFiles`. Unverified: no diff of real saved files before and after, one round-trip test
  - [x] R2d the ruleset's own data (track r2d): `rules.Source` and `rules.Builder`; `rules1e` reads
    and caches its monsters, treasure and item table; the vault hands over opaque records
    (`rules.AddTreasureRecords`); package `pool` is the one place that names the pool games' rulesets
    (`pool.New`, `pool.Settings`, `pool.ReadRoster`)
  - [x] R2e the guard has no list. The engine's tests still reach into the ruleset: 97 of 188 test
    files import `rules1e`, about 760 reads of its state (`e.state()`, `e.member()`); they stay until G2
  - [ ] R2f the engine names no AD&D thing at all: `rules/adnd.go` (Ranger for PartySurprise, Knock
    at a door, the thief skills and the petrify save of the Sel script fields) goes; the ruleset says
    which fact a script field reads, and answers the two Checks by what they ask, not by class or
    spell. With B1d at the latest: XXVc has none of these
- [ ] G2 One state. Inventory (G2a, 2026-11-15, an agent's read of play.go, gamestate*.go,
  gamecombat.go, view_engine.go, present_engine.go, rules/effect.go; unread: the rules' method
  bodies, ui):
  - After each action `rebuild()` builds the whole `game.State`: per member about 30 ruleset reads
    for the sheet page, one `Allowed(UsingItem)` per carried item, on the overland 1584
    `Check(AreaTileIs)`; `StateBuilds` is the version the ui caches by.
  - The build is not a pure read: `battle()` and `Note()` call `present()` (queued effects applied),
    `applyItems` (creationState, offers) and `EffectsList` and `e.as` write `WordWho`, `canCastNow`
    rewrites and restores the spell list and `e.sel`, `userLabels` creates map entries, `selected()`
    resolves `EmptyMember`, `rebuild()` resets `lootSeen`.
  - Disk per build: `mapName()` (names/pictures.txt), `BuildCharList()` twice (stateOf, allowed).
  - The mode is derived, not stored: `stateMode()` reads about 12 inputs, `stateOf` overrides it by
    5 more (ending, intro, sequence, results, handout), `build()` by `e.creation`.
  - The rules announce almost no change: `present.apply` knows Narrated, Turned, Casting, Shot,
    Sounded, Walked, Attacked, Struck, Fell (the fight's board). Hit points, items, money, clock,
    position and 99 `setWorldWord` sites change with no effect; Healed, Hurt, Traded are produced and
    not applied; Gained, Lost, Took, Fled, Left, Placed, CloudOn/Off, TerrainSet have no producer.
  - Two copies copied again: `fightState()` builds `game.Combat`, `stateOf` copies part into
    `game.Fighting` (and `allowed()` reads the first); `members()` builds the engine's `Member`,
    `partyState()` a `game.Member`. `pushCombat` calls `build()` mid-step for `game.Engaged`.
  - Never set by the engine (BR's engines may; decide with B1): `Over`, `ShipFight`,
    `Scene.Speaker/Pictures/Screen`, `Common.NoSave/NoLoad/Mark`, `Where.Day/Month/Year/Music/Detail`,
    `Trading.Name/Pooled/Vault`, `Looting.Status`, `Draft.MinHP/MaxHP/Step/Stats/Derived`,
    `Unit.Dead/Gone/Pose`, `Battlefield.FloorTile/Clouds`, `Item.Cursed/Charges`.
  - [x] G2a the inventory above
  - [~] G2b the build is a pure read (track g2b, landed): queued effects are applied in `rebuild()`
    before the build; the `WordWho` writes of `EffectsList` and `fightState` are gone; `canCast(id)`
    asks the member's spell menu (no `e.sel`, no spell list rewritten); map labels read without
    creating; `lootSeen` reset where a treasure opens. Guard: `TestBuildReadsOnly` (12 scenes; it
    caught only the queued effects, the rest is by reading). Left, both go with G3: `applyItems` in
    `creationState` and `offers` (the rules rewrite the character to get AC and THAC0), `selected()`
    resolving `EmptyMember` for an empty party. Changed: a question during the treasure no longer
    announces the loot again
  - [x] G2c no disk read in the build (track g2c, landed): the map names are read once, the roster
    list when `charlist.txt` is written or the save folder changes (`rosterChanged`). Guard:
    `TestBuildReadsNoDisk` (party menu, hall, shop, loot, fight; not the overland),
    `TestRosterListRefreshed`. Noted, not looked into: Quick Start's per-member save leaves one
    name in charlist.txt
  - (G2e, the mode as a stored value: dropped. `stateMode()` is a pure function of state the
    engine owns already (what it waits on: `pending`, the top continuation, `menu.on`, the
    question, `ext.intro`); a stored mode beside those would be a second source of the same fact.
    Making that wait state itself one value is the engine's continuation machinery, not this item)
  - (G2d, one copy per thing inside the engine: not done on its own. The copy chains
    `members()`/`partyState()`, `sheetPage()`/`gameSheetPage()`, `battle()`/`fightState()` are
    deleted whole by G3; shortening them first is work thrown away)
  - RULING (Felix, 2026-11-15), replaces G2f: one source of truth for every piece of state
    (CLAUDE.md). State is split by owner and never copied across:
    the ruleset owns characters, items, money, spells, effects, time, the fight's units and ground,
    and the ui reads them from a read-only view of the rules when it draws; the engine owns where
    the party is, the screen, the pending question, what was seen, quests and journal marks, in one
    small struct of `game` types it writes where they change; the adventure's data is loaded once;
    the ui owns presentation only. No notifications from the rules are needed.
- [ ] G3 The contract by owner (after B1, when `engine` is the only `game.Game`): `game.State`
  shrinks to the engine-owned part (the party is ids and the selected one); `game.Game` gives the
  ui the rules' read-only view; the ui reads sheets, items, spells, units from it, screen by screen;
  the allowed actions become a question asked when needed; `rebuild()`, `build()`, the copy chains
  and the engine's own `Member`, `SheetPage`, `Battle`, and `game.Combat`, `game.Member`,
  `game.Item`, `game.SheetPage`, `game.Unit` as copies go; a version counter bumped per action
  tells the ui to redraw. Known sizes: `.Party` read in 41 places of ui, 92 test sites build
  `game.Member`. To pin: the engine acts and the ui draws never interleave
- [ ] B1 Buck Rogers as data folders (decision above; survey 2026-11-15). BR1 and BR2 are two engines
  of ~15k lines each (near copies), `buckrogers-core` (2.6k: its own script VM, combat map, ships,
  space, the importer) and `rulesxxvc` (8k, already a `rules` ruleset, but not what runs: the BR
  engines carry their own rule code). They run in the one binary on `ui`, not on `engine`.
  - [x] B1a pinned: `asset/brmeshes_test.go` TestBuckRogersMeshesKept (the 19 models with model,
    Blockbench source and texture, no other; each a tile a combat map can produce, its own picture
    there; props.json names only such tiles). The frame counts per set are copied from BR1's
    rules/cmap.go until B1c
  - [ ] B1b one script machine. Survey 2026-11-15 (an agent's line-by-line read; BR2's Pascal in
    buckrogers2-re was read, BR1's is not on this machine; BR2's scripts are not imported here).
    ARCHITECTURE.md section 2 is partly stale: `gobox/ecl` values are int64 already (the 16-bit cut
    is the engine's store, `cutToWidth`), V4 with FOR/NEXT is a dialect of `gobox/ecl` already, and
    both BR engines already decode, parse and load scripts with it. Only the executor is separate:
    `eclbr.VM` (159 lines) and each BR engine's ecl.go/eclops.go. What differs:
    - reads agree: BR zero-extends bytes and words (verified in BR2's ReadGbvm), compares signed
      32-bit; only long operands (types 4, 5) can be negative, and `ecl.VM.Arg` reads those as 0
    - stores differ: Pool keeps 16 bits in one cell per address whatever the operand's type; BR
      writes as many bytes as the destination operand's type into byte memory (a word at a also
      writes a+1). `ecl.Layout` has one cell per address, no aliasing
    - BR has no layout: BR1's scripts use 304 raw addresses (248 area, 37 party, 12 aux, 7 view),
      121 as words, 4 as longs; 59 fixed indices in the BR1 engine's Go. `ecl.VM` panics on a
      variable with no Ref. BR2: not counted
    - Return with an empty call stack: Pool goes on, BR ends the script
    - V4 has no Skip table (BR's rule differs from `Fixed` for 12, 31, 52, 54, 59, 60 and the
      counted commands) and no Calls map (BR dispatches Call on raw addresses, V2's keys)
    - FOR/NEXT: one `LoopAt` and two area cells; not in `ecl.VM` nor the engine's ops
    - waits: BR keeps one wait and one `Next` frame in the VM (`Wait`, `WaitOp`, `Frame`,
      `scriptDone`'s switch: the block chain, depart/arrive, look, input); gobox keeps a stack of
      continuations beside a passive VM. Neither saves a script's wait
    - random: BR uses PCG, the original and gobox the Turbo Pascal generator; changing it changes
      every seeded BR test
    - of the commands the plan called BR's own, 30, 50, 54, 56, 57, 58, 62, 65 have a Pool
      counterpart in the engine by name and arity (unrun); BR's own are 31 SpaceFight, 34/35 Skill,
      63 Logbook, 64 TakeItem, 66, 67 Sound, 69 WallDistance, 70/71, 72 Stash, 73 Damage6, 74 Duel,
      75 Shop, 76 AreaKind
    Steps:
    - [ ] B1b1 `gobox/ecl` can run V4, the Pool dialects unchanged: long operands read, FOR/NEXT,
      V4's Skip and Calls, Return on an empty stack by dialect; tests on hand-assembled V4
    - [ ] B1b2 BR's memory as a layout: a scan of BR1's (and BR2's, once imported) scripts for
      what the one-cell model cannot hold (a word and its second byte both used, a byte destination
      given more than 255); from it either a generated full layout (unnamed cells as `WordXXXX`,
      the engine's 59 fixed indices named) or a ruling on byte memory (CLAUDE.md: no raw memory)
    - [ ] B1b3 the store by dialect (16-bit cell, or the operand's width) as data of the Version
    - [ ] B1b4 BR's waits and frames as engine continuations; with B1f
    - [ ] B1b5 one random generator (`rules.RNG`); BR's seeded expectations derived again
    - [ ] B1b6 `eclbr.VM` deleted
  - [ ] B1c the importer: `ecl-br/importbr`, `importbr1`, `importbr2` into `importer`; `data/br1`,
    `data/br2`; nothing read from the originals at run time (BR2's START.EXE and GAME.OVR, BR1's
    data segment become import output)
  - [ ] B1d `rulesxxvc` into the gobox module beside `rules1e`; it is what runs. Every difference
    from the BR engines' rule code is listed; ruling per difference (default: open, ask Felix)
  - [ ] B1e what only BR has, into `engine` (space fight, ships, skills, jetpack, throw), switched
    on by adventure data and the ruleset, never by game name; `ui` parts exist (spacefight.go,
    shipsheet.go)
  - [ ] B1f BR1 on `engine`; its tests moved or replaced; then BR2
  - [ ] B1g `playbr1`, `playbr2`, the two BR engines, `buckrogers-core`, the four repos deleted;
    `go.mod` has no replace to a sibling game; gate.sh and foldwt.sh without them

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
- [x] G1 `facts`/`known()` removed
- (G2, open: see the order of work above) One state: the engine keeps the player-visible state in `game` types and writes it where it changes; `rebuild()`/`build()` and the per-mode copies go. Deferred (G1 agent): the visible state lives in rules, script words and battle structs; writing it where it changes needs change events from rules (party, items, purse, clock, position), then automap, then combat, then replacing build/stateOf/allowed. Today rebuild runs once per action, not per frame
- [x] G3 `game`'s non-contract parts out: journal layout → `ui`, point-buy maths → `rules`, `held.go`/`memberpage.go` checked
- [x] N1 Shared font pool: the importer adds each game's fonts to one pool; `ui` cycles through all; `Data.Font`/`Fonts`, `TextRun.Upcase` leave `game` (ruling 2)
- [x] N2 Title, ending and story screens as declarative scripts plus exported picture elements; `ui` composes them (ruling 3); `Layer` leaves `game`
- [x] N3 The importer cuts the menu logo into a finished image; `Data.Logo` goes (ruling 10)
- [x] Y1 Ponytail audit (2026-11-13): delete production functions only tests call (~140, ~1.5–2k lines; list: deadcode without -test, minus `fixture/`), with the tests that only test them; true test helpers move to `_test.go`
- [x] Y2 Delete the 5 functions nothing calls (`ecl.NeedsPlainTests`, `importer.CotabScriptLayout`, `SotsbScriptLayout`, `ImportAssets`, `rules.SymbolBlock`)
- [x] Y3 Code only the dev scenes reach since D1 (`firstOkay`, `okay`, `blockedRet`, `fileBytes`, `heldOf`, `entryNamed`, `invUp`, `fullPicture`, `firstMonster`, `ecl.ArgsTrace`): delete or move into devviews_engine.go
- [x] Y4 (nothing folded: `ui/input` would cycle through `ui/combat`; `idgen`, `winicon` are go:generate tools) Fold the one-file packages `ui/input`, `rules1e/idgen`, `cmd/gobox/winicon` into their parents where no import cycle forces them
- [~] E1 `engine` stops importing `asset`/`asset/orig`: pictures leave as IDs (after U1, N2, the A1 remainder); colour constants become the ruling-1 colour names; the remaining original-data reads (`orig.NewLoader` in dataimport.go, `orig.Monster` records, `orig.RulesFile`) move to the importer's output. A test forbids it. Done: no asset/orig in engine; text colours game.Ink; save slots and data-folder IO in adventure; Amiga player built by audio. Done too: icons/portraits leave as PNG bytes, ui decodes (pixel-identical); Battle.Floor dropped. Remaining (needs pixel checks per game): 1. combat tiles: sqrpaci is relaid (LayTiles per fight, layAreaTiles) and recoloured in place at run time, so the engine must first name which file/block/frame sits in each slot and recolouring must become a typed fact before ui can load tiles by reference; groundOf counts the ground per set at run time; 3. 3D sprites, pic.go layers, recolourFrame (palette animation as a typed fact); 4. creationhost Pick/Set; 5. then TestEngineNotAsset
- [x] E2 `engine` stops importing `audio`: it publishes which song/cue/voice line plays as typed facts in `game`; `ui` (host) drives `audio` (sound.go's `audio.Driver` use, voice.go's Vorbis decoding). A test forbids it
- [x] R1 `cmd/gobox` stops importing `rules1e` (`rules1e.Settings{PathSearch}`, `rules1e.ReadRoster`): behind `rules`/`engine`
- [x] R3 `rules4e` imports `engine` (so `engine` cannot pick the 4e combat rules itself and `cmd/gobox` hands it in): a ruleset must depend on `rules` only; then `cmd/gobox` passes a `Combat4e` setting
- (R2, open: see the order of work above) `engine` reaches the rules only through `rules` (today ~600 uses of `rules1e` types: `Character`, `Item`, classes, slots, directions). Needs its own plan: a ruleset-neutral character and item model in `rules`
- [x] D2 `game` and `adventure` both define `Sequence`, `Layer`, `Credit`: one canonical definition (with N2)
- [ ] P1 The engine places and draws nothing (2026-11-14, Felix). Goal: the game contract cannot express a screen position, and no engine-side package (game, engine, ecl, adventure, rules1e, rules4e) depends on drawing. Enforced by `game/noui_test.go` (TestEngineDrawsNothing, TestEnginePlacesNothing), no allowlist. Rulings: world squares use `rules.MapPos`/`rules.BoardPos` (sizes too), never `image.Point`; sequences and scenes say placement declaratively ("centred", new lines); the one exception is stacking over a background (title text on a picture): normalised 0..1 coordinates; pre-wrapped text (`[][]TextRun`) is allowed, it keeps the original's design; image building leaves the engine (E1's remainder); wherever the presenter can place a thing itself it does, and the engine says only what is shown (a sprite's offset is its art's own data). Steps:
  - [x] P1a `game` fields: world coordinates to MapPos/BoardPos
  - [x] P1b The engine publishes facts, never positions: `Drawn`/`Layers`/`Picture.Off` leave `game`; `Common.Picture` (+ Frames/Delays), `Common.Portrait`, Full; a sprite's offset is its art's own (CellX/CellY); the engine's positioned page model goes (the UI knows when a thing is cancelled and asks the engine). Placement only in title/ending sequences (0..1 coordinates). Rejected: slot/placement tags on draw calls
  - [x] P1e The text page as facts: the pane's lines, the prompt and the view's pre-wrapped lines are published as text; `e.char(x, y, …)`, `Flow` into cell boxes, `blank`, `unwrite` and `Draw3DView` (only setWin + unwrite) go; text leaves on events, not by rectangle
  - [x] P1f The classic 3D renderer (engine/classic3d_test.go, TestClassicViews for phlan-poc/compare/compare.py) moves out of the engine into its own test/debug tool; it stays for comparison sheets only
  - [ ] P1c image building out of the engine (E1 remainder): engine no longer depends on image/draw via `asset`
  - [ ] P1d full suites, screenshots of titles, endings, story scenes per game
- [ ] Q1 Open items carried from ONE-BINARY.md and ARCHITECTURE.md (checked 2026-11-14):
  - [x] native scripts gone: CastKnock and the tavern game are script commands (`Scratch6 = CastKnock()`, `TavernGame()`, gobox ecl V1 67/68) placed by POR's import fixes 30 and 58; the label hijack (`GameTables.Natives`) is cut. Bug class found: a call named like a label parsed as a Gosub to it (TavernGamble() looped); the parser now refuses a name that is both
  - [ ] name the unnamed script cells (`WordXXXX`, 585 distinct in COTAB, SotSB, POD scripts) from their script use
  - [x] `WriteGbvm($2CB, …)` in cotab-re and sotsb-re EclSurprise: a reconstruction bug (the zone-1 offset, not $7ECB; WriteGbvm dropped it). Fixed cotab-re 5c1d47c, sotsb-re 4f469e9; the Go engine was right (`WordSurprise` at 7ECB). Detect: literal Read/WriteGbvm addresses below a game's lowest zone; none left
  - [x] `ui/blobber` `FloorRGBA`/`CeilRGBA`: no change needed; ruling 7 holds, `game` carries `FloorInk`/`CeilInk` (Ink), and `ui` resolves the RGBA for the renderer
  - [ ] exit hints (`Question.Esc`) for POD, COTAB, SotSB, BR script menus (por-go parity dropped, Felix 2026-11-14: mostly there)
  - [ ] unverified, needs a run: in-fight sounds by cue name silent; fall-sound PC/AdLib fallback outside fights; toasts replacing pane lines outside fights; rest-time line refresh only on the 5-step tick; song 6 overland / song 4 entering Phlan; camp fire animation; wiped party back to the party menu
- [x] K1 `TestKeyDispatch` fails since the app → ui merge: `ui/shell.go` (`ScriptTimed`) and `ui/shellfront.go` (F12, Shift-F12) name key codes instead of dispatching through the key table
- [x] K2 `SOUND.OFF` is read at start (engine/sound.go) but nothing writes it since `soundToggle` was cut as unreachable: drop the read or wire the toggle
- [x] C1 Credits: `Credit.X/Y/Colour`, `Data.CreditBoxes`, `CreditGaps` (original two-column credits; ruling needed)
- [x] U5 Per-game values U1 flattened to one ui constant (cotab ready mark, door colour, Amiga combat ground grey): restore the originals' values
- [x] U6 `ui` does not scale `StepMs`/`Ask.Ms` by the player's speed (ruling 5); `Walls.FloorRGBA/CeilRGBA` still RGBA, not a palette index (ruling 7)
- [x] L1 Load regression (S1 cut `legacySave`): every shipped/dev slot with Area/Game/Purse at the top of GAME.JSON was refused ("unexpected end of JSON input"). Fixed on disk once (12 por dev slots under data/por/sources/save/saves; backups in the session scratchpad). Bug-class check: a test that loads every slot under data/*/sources/save/saves; also `loadFrom` drops gob decode errors of MAPS/VOICED/BESTIARY/... (golden rule). sotsb home `town.sav` holds an original SAVGAMA.DAT slot (no GAME.JSON): convert it with the importer's save conversion
- [x] Final: full build/vet/test, all repos (2026-11-13): BR repos all ok; gobox all ok but engine TestPhlanHistory (fails the same at commits long before this refactor with today's data: pre-existing). Found: the tests' por tree (sources/assets, ImportDataIfStale) is a second import path beside PorImport; it missed look.json (fixed by one writePorLook) and games/por/cmd/importstale's relative paths look wrong. Fold the test tree into the one import (follow-up)

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
