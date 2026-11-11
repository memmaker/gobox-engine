# Track A: import and run-time data (agent A8, 2026-11-11)

Goal: `gobox import <game>` writes `adventures/<id>/` (scripts, art, data, manifest) for all six games; nothing at run time reads originals. Importers depend only on gobox/importer, dax, asset (and the game's own decoders), never gobox/app or front.

## Checklist
- [x] one shared Gold Box maps/records decoder: gobox/dax data.go (ImportData, Blocks, Monsters, ItemBlocks, Records, Table); cotab/sotsb/pod `rules.ImportDataTo` writes GEO<n>.bin/.txt, MON<n>.json, ITEM<n>.json, ITEMS into data/; TestDataEqualsDAX each
- [x] cotab/sotsb sound tables decoded at import (import<game>.importSound, data/sound.json); importcotab/importsotsb tests: equal to rules/sounddata.json
- [x] BR1/BR2 importers: ecl-br/importbr (scripts V4 no fixes, art, GEO/MON0/ITEM0, data/tables.json = xxvc.Tables with ItemTypes from ITEMS and Items from ITEM0 block 1); importbr1, importbr2 register; TestImportEqualsDecode each
- [x] importer.Manifest.Start/End as `*importer.Sequence` (Format 2), filled from Game.Start/End (shape below)
- [ ] xxvc.Tables.ItemWords (START.EXE word table) and Monsters (typed from MON0CHA by eclbr record offsets) in importbr (X3 asked)
- [ ] portrait/icon map (C5 asked): per game, DAX picture numbers (portrait head, body, icon head, icon body) to the pool / buck-rogers collection entry IDs, a typed table in data/ plus a shared reader (sprites/cmd/poolsprites already knows origins.json per game: build it from that)
- [ ] BR zone1 revivesParty ($7EE6, != 0) and skipsCombatItems ($7EE3, == 1) as typed bools (coordinator): these are run-time ECL-written zone bytes (postcbt.pas 120, 224), not static area data. Proposal: BR's script layout (scripts/layout/global.txt: `7EE3 1 byte game SkipsCombatItems`, `7EE6 1 byte game RevivesParty`) written by importbr, so the engine reads them by name and maps them to core.FightTerms.NoItemLoot / RevivesParty. Needs ecl/script to accept a BR layout (BR runs with no names/layout today, so the texts change: update the runtime cache test with it). Not done.
- [ ] sound: pod (no PC-speaker tables: AdLib MUSIC/ZOUNDS), BR1/BR2 (MUSIC.DAX); AdLib to MIDI, Amiga SOUND.BIN to data; then drop embedded sounddata.json (cotab, sotsb) and run-time DAX reads once B5's folder reader reads data/sound.json
- [ ] title/end sequences filled per game into Game.Start/End (from each game's title.go / gamedata.go Intro, TitleSteps, CreditLines, CreditBoxes, TitleBackdrop; PARITY-BASELINE "Title sequences per game")
- [ ] cotab/sotsb/pod import tests still use gobox/app (app.ImportArt) for the reference art: switch to dax.ImportArtIfStale as BR's do

## Manifest.Start/End shape (for B5's folder reader)
`Sequence{Steps []SequenceStep; Credits []Credit; Boxes []Cells; Backdrop *Art; Short []int}`; `SequenceStep{Kind StepKind (StepPicture | StepCredits | StepAnim); Layers []Layer; Page; Anim; Song; HoldMs (0: until key/click); Lightning; Whole}`; `Art{File, Block}`, `Layer{Art; At Cells; Frames; FrameMs}`, `Cells{X1,Y1,X2,Y2}` (40x25 cells), `Credit{X, Y, Colour, Text, Page}`. Maps one to one onto game.Intro, play.TitleStep, Data.CreditLines/CreditBoxes/TitleBackdrop/NoCredits (Short), game.Sequence (StepAnim).

## coordinator:
- B5: por-go rules/dataimport.go should call `dax.ImportData[charRecord, affectRecord, Item]` and `dax.Table[itemBase]` instead of its own walk (same output names).
- BR record types stay raw arrays (eclbr.Record [259]byte, Item [62]byte, effects [9]byte) in MON0.json: BR has no typed record struct; the typed form is xxvc.Tables.Monsters (open item above).
