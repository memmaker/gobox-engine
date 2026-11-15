package importer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gobox/adventure"
	"gobox/asset/orig"
	"gobox/ecl"
	"gobox/game"
	"gobox/rules"
	"gobox/rules1e"
)

// What BR1's start-up files hold, each from its evidence.
const (
	// brFirstBlock: buckroger-re eclvm.pas START_ECL_SCRIPT = 16 ("ECL1.dax's first record"); buckrogers-go rules/adv.go begin().
	brFirstBlock = 16
	// brGameSet: the disk a new game starts on (buckrogers-go rules/adv.go begin(): e.disk = 1).
	brGameSet rules.Set = 1
	// brFirstBigPicture: Picture numbers from 112 show BIGPIC (eclvm.pas:444; buckrogers-go rules/eclops.go picture).
	brFirstBigPicture = 112
	// brCodeBase and brCodeSize: the loaded block's script addresses (buckrogers-core ecl.go AddrCode = 32768, CodeSize = 9217).
	brCodeBase = 32768
	brCodeSize = 9217
	// brBorderSlot and brGameSlot: Load8x8Set(0, 6) is BORDERS block 0 (sym8x8.pas BORDER_SLOT = 6), Load8x8Set(202, 4) the data
	// set's 8X8D<set> block 202 (gameinit.pas:243-244).
	brBorderSlot, brGameSlot   = 6, 4
	brBorderBlock, brGameBlock = 0, 202
	// brSeries is presentdata/data.go's Series (buckrogers-go).
	brSeries = "buck-rogers"
)

// ImportBREngine is ImportBR plus what the shared engine reads from an adventure folder (engine.New, Init): the game's
// full script layout in place of ImportBR's partial one (the scripts re-imported with it), gametables.json and gamedata.json.
// Only tests call it; it is not in the registry or `gobox import`. It replaces ImportBR when the BR engines go (REFACTOR.md B1g).
// A game without a full layout (BR2) is an error.
//
// Not written yet: sound.json (no layout of BR1's sound tables), portraits.json, the walls (BR has no WALLDEF), BR1's PIC file
// remap per picture number (picOf in buckrogers-go rules/eclops.go: 105, 64 -> PIC7; 72, 92, 74 -> PIC8; 62 -> PIC9).
func ImportBREngine(id string, e BREXE, dos, dst string) error {
	full, err := fs.ReadFile(embeddedScripts(id), ecl.GlobalLayout)
	if err != nil {
		return fmt.Errorf("%s has no full script layout yet: %w", id, err)
	}
	if err := ImportBR(id, e, dos, dst); err != nil {
		return err
	}
	l, err := orig.NewLoader(dos)
	if err != nil {
		return err
	}
	scripts := filepath.Join(dst, "scripts")
	if err := os.Remove(filepath.Join(scripts, ecl.PartialLayout)); err != nil {
		return err
	}
	g := filepath.Join(scripts, ecl.GlobalLayout)
	if err := os.MkdirAll(filepath.Dir(g), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(g, full, 0o644); err != nil {
		return err
	}
	if _, err := ecl.Import(ecl.V4, orig.DAXSource(l), os.DirFS(scripts), scripts, ecl.FoldGates, nil); err != nil {
		return err
	}
	data := filepath.Join(dst, adventure.DataDir)
	if err := adventure.WriteGameData(data, game.Data{Series: brSeries, ID: id}); err != nil {
		return err
	}
	tables := adventure.GameTables{
		Dialect:      rules.DialectBR,
		FirstBlock:   brFirstBlock,
		GameSet:      brGameSet,
		BigPictures:  adventure.BigPictures{First: brFirstBigPicture},
		OriginalSave: rules1e.SaveLayout{CodeBase: brCodeBase, CodeSize: brCodeSize}, // eclLayout reads only these two
		Symbols: []adventure.SymbolLoad{
			{Set: brBorderSlot, File: "BORDERS", Block: brBorderBlock},
			{Set: brGameSlot, Block: brGameBlock},
		},
		Art: &adventure.ArtNames{CombatSprites: "COMSPR", Title: "TITLE"}, // gameinit.pas LoadIconAnim('COMSPR'); br1Start's TITLE
	}
	return WriteJSON(filepath.Join(data, adventure.GameTablesFile), tables)
}
