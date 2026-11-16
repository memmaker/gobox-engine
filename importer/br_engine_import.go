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

// brSeries is presentdata/data.go's Series (buckrogers-go), the same in buckrogers2-go rules/data.go:14.
const brSeries = "buck-rogers"

// brStart is what one game's start-up files hold: gametables.json's values, each from its evidence. Both games' sources are
// buckroger-re and buckrogers2-re (the Pascal), buckrogers-go and buckrogers2-go (the old engines), buckrogers-core (shared).
type brStart struct {
	// FirstBlock: eclvm.pas START_ECL_SCRIPT = 16 ("ECL1.dax's first record", both games); rules/adv.go begin() (m.Block = 16).
	FirstBlock int
	// GameSet: the disk a new game starts on (rules/adv.go begin(): e.disk = 1, both games).
	GameSet rules.Set
	// FirstBigPicture: Picture numbers from here show BIGPIC (eclvm.pas FIRST_BIG_PICTURE = 112; rules/eclops.go picture).
	FirstBigPicture int
	// CodeBase and CodeSize: the loaded block's script addresses (buckrogers-core ecl.go AddrCode = 32768, CodeSize = 9217).
	CodeBase, CodeSize int
	// Symbols: Load8x8Set(0, 6) is BORDERS block 0 (BR1 sym8x8.pas BORDER_SLOT = 6), Load8x8Set(202, 4) the data set's 8X8D<set>
	// block 202 (gameinit.pas:243-244 in BR1, :248-249 in BR2).
	Symbols []adventure.SymbolLoad
	// Art: gameinit.pas LoadIconAnim('COMSPR') (BR1 :253, BR2 :253); the title file the game's start sequence shows.
	Art adventure.ArtNames
	// PictureSets: pO124_02D0 reads these PIC blocks from PIC7..9 whatever the disk (rules/eclops.go picOf in BR1, smallPic in BR2).
	PictureSets map[int]rules.Set
	// Walls: the game has a wall texture table (importer/walls/<id>.txt); false: none yet (BR2).
	Walls bool
	// Data: the front's data (gamedata.json) beyond the series and id; nil: not moved over from the old front yet (BR2).
	Data func(id string) game.Data
	// Logbook: the journal's titles (importer/br1/journal, from the printed Log Book); nil: none yet.
	Logbook *adventure.Logbook
}

const (
	brBorderSlot, brGameSlot   = 6, 4
	brBorderBlock, brGameBlock = 0, 202
)

var (
	brSymbols = []adventure.SymbolLoad{
		{Set: brBorderSlot, File: "BORDERS", Block: brBorderBlock},
		{Set: brGameSlot, Block: brGameBlock},
	}
	// brPictureSets is BR1's and BR2's (buckrogers-go eclops.go picOf; buckrogers2-go eclops.go:254-260 smallPic: the same six).
	brPictureSets = map[int]rules.Set{105: 7, 64: 7, 72: 8, 92: 8, 74: 8, 62: 9}
)

// brStarts by game id.
var brStarts = map[string]brStart{
	"br1": {FirstBlock: 16, GameSet: 1, FirstBigPicture: 112, CodeBase: 32768, CodeSize: 9217, Symbols: brSymbols, PictureSets: brPictureSets,
		Art:     adventure.ArtNames{CombatSprites: "COMSPR", Title: "TITLE"}, // br1Start's TITLE
		Logbook: &adventure.Logbook{Box: "Log Book", Title: "Log Entry "}, Data: br1GameData, Walls: true},
	// BR2: buckrogers2-go rules/adv.go:147,161 (block 16), :158 (disk 1), eclops.go:239 (112), buckrogers2-re eclvm.pas:42,48. The
	// Logbook stays unset: buckrogers2-go has no rules/journal to take the titles from (open: BR2's journal).
	"br2": {FirstBlock: 16, GameSet: 1, FirstBigPicture: 112, CodeBase: 32768, CodeSize: 9217, Symbols: brSymbols, PictureSets: brPictureSets,
		Art: adventure.ArtNames{CombatSprites: "COMSPR", Title: "TITLE"}}, // br2_importbr2.go br2Start's TITLE
}

// ImportBREngine is ImportBR plus what the shared engine reads from an adventure folder (engine.New, Init): the game's
// full script layout in place of ImportBR's partial one (the scripts re-imported with it), gametables.json and gamedata.json.
// Tests and `gobox import -shared` call it (Game.ImportShared); the default import stays ImportBR. It replaces ImportBR when the BR engines go (REFACTOR.md B1g).
// A game without a full layout or a start entry in brStarts is an error.
//
// Not written yet: BR1's sound tables (sound.json is the silent one, brSilentSound), portraits.json, the walls' flats and BACK pictures.
func ImportBREngine(id string, e BREXE, dos, dst string) error {
	full, err := fs.ReadFile(embeddedScripts(id), ecl.GlobalLayout)
	if err != nil {
		return fmt.Errorf("%s has no full script layout yet: %w", id, err)
	}
	start, ok := brStarts[id]
	if !ok {
		return fmt.Errorf("%s has no start-up values (brStarts)", id)
	}
	if err := ImportBR(id, e, dos, dst); err != nil {
		return err
	}
	if err := writeRoster(dst); err != nil {
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
	if start.Logbook != nil { // BR2 has no journal yet
		if err := ImportJournal(id, filepath.Join(dst, adventure.JournalDir)); err != nil {
			return err
		}
	}
	data := filepath.Join(dst, adventure.DataDir)
	front := game.Data{Series: brSeries, ID: id}
	if start.Data != nil {
		front = start.Data(id)
	}
	if err := adventure.WriteGameData(data, front); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(data, adventure.SoundFile), []byte(brSilentSound), 0o644); err != nil {
		return err
	}
	tables := adventure.GameTables{
		Dialect:      rules.DialectBR,
		AddNPCIcon:   true, // saveload.pas AddCreatureToParty loads the NPC's CPIC icon (both games)
		FirstBlock:   start.FirstBlock,
		GameSet:      start.GameSet,
		Logbook:      start.Logbook,
		PictureSets:  start.PictureSets,
		BigPictures:  adventure.BigPictures{First: start.FirstBigPicture},
		OriginalSave: rules1e.SaveLayout{CodeBase: start.CodeBase, CodeSize: start.CodeSize}, // eclLayout reads only these two
		Symbols:      start.Symbols,
		Art:          &start.Art,
	}
	if start.Walls {
		// the 3D view's textures: buckrogers-go rules/wallindex.txt as walls/<id>.txt (chunk (value-1)/5+1, wall (value-1)%5+1); the
		// flats and the BACK picture are not published yet (a view's floor and sky: backFlats by BACK block, not by area)
		tables.Walls = wallTextures(id, "", "")
	}
	return WriteJSON(filepath.Join(data, adventure.GameTablesFile), tables)
}
