package importer

import (
	_ "embed"
	"encoding/json"
	"errors"
	"gobox/adventure"
	"gobox/asset"
	"gobox/asset/orig"
	"gobox/audio"
	"gobox/rules"
	"gobox/rules1e"
	"io/fs"
	"os"
	"path/filepath"
)

// portraits is POR's portrait tables (which HEADn/BODYn.DAX block shows each
// portrait number: 1..14 heads and 1..12 bodies are the original's, set 3;
// the rest are every other one the files hold), game data checked in here
// and written to data/assets/data by the import; pool.ReadGameTables reads it.
//
//go:embed por/portraits.json
var portraits []byte

// gameTables is POR's pool.GameTables other than its portraits (data/gametables.json):
// the duel sounds of CALL 0xBA03 (8 a step, 10 the fireball, DS:2612's numbers); any other plays nothing.
// The import's script fixes; the native scripts of the tavern game and the cabinets'
// CAST KNOCK; the inn's and the tavern's Who commands spend the one purse (ECL3 A1B4, ahead of A501).
var porGameTables = adventure.GameTables{ScriptSounds: map[byte]adventure.ScriptSound{
	8:  {Cue: rules.CueStep},
	10: {Cue: rules.CueFireball},
},
	ScriptFixes:  porScriptFixes,
	OriginalSave: porSave,
	Quests:       porQuests,
	Art:          &porArt,
	MapLabels:    porMapLabels,
	Music:        &porMusic,
	Natives: []adventure.NativeAt{
		{ScriptAt: adventure.ScriptAt{Area: 0, Label: "TavernGamble"}, Native: adventure.NativeTavernGamble},
		{ScriptAt: adventure.ScriptAt{Area: 14, Label: "CastKnock"}, Native: adventure.NativeCastKnock},
	},
	PurseWhos: []adventure.PurseWho{
		{ScriptAt: adventure.ScriptAt{Area: 0, Label: "PayPlatinum"}, By: 0},
		{ScriptAt: adventure.ScriptAt{Area: 0, Label: "TavernGamble"}, By: -1},
	},
}

// dataFormat changes when the files ImportData writes change.
const dataFormat = "20" // 20: GameTables.Music.Files, AmigaSongs (P1); 19: Sheet.Race is the typed race (a name), MonsterRace the record byte of a monster; 12: MON<n>.json records as the rules' Sheet (char.SheetOfRecord); 13: GameTables.Art (Pg); 14: no Restoration tip (Rz); 15: GameTables.MapLabels, Music (Ph); 16: script fixes gold only (Gc); 17: area 0 wager text in gold (Df)

// ImportData is the data import: data's GEO, MON and ITEM archives, the
// ITEMS table and START.EXE's sound tables (soundFile) as data/assets/data (DataDir).
// It clears the old data first and writes the picture table as the por import does.
func ImportData(data string) error {
	adv := filepath.Join(data, adventure.AssetsDir)
	if err := os.RemoveAll(filepath.Join(adv, adventure.DataDir)); err != nil {
		return err
	}
	if err := PorImportDataTo(filepath.Join(data, adventure.DOSDir), filepath.Join(adv, adventure.DataDir)); err != nil {
		return err
	}
	return asset.WritePictures("por", adv)
}

// ImportDataTo is ImportData from the DOS files in dos into dst; it clears
// nothing (the adventure import writes beside it, pictures.PicturesFile).
func PorImportDataTo(dos, dst string) error {
	l, err := orig.NewLoader(dos)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	if err := orig.ImportData[adventure.CharRecord, rules1e.AffectFile, rules1e.ItemFile](l, dst, rules1e.SheetOfRecord[adventure.CharRecord], rules1e.ItemOfRecord[rules1e.ItemFile]); err != nil {
		return err
	}
	if err := orig.ImportItemTable[rules1e.ItemBase](l, dst, adventure.ItemTypes, orig.Counted); err != nil {
		return err
	}
	exe, err := os.ReadFile(l.Path("START.EXE"))
	if err != nil {
		return err
	}
	snd, err := audio.DecodeSoundData(exe, adventure.SoundLayout)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dst, adventure.SoundFile), snd, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dst, adventure.PortraitsFile), portraits, 0o644); err != nil {
		return err
	}
	b, err := json.MarshalIndent(porGameTables, "", " ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dst, adventure.GameTablesFile), b, 0o644); err != nil {
		return err
	}
	if err := adventure.WriteGameData(dst, porGameData()); err != nil {
		return err
	}
	if err := ImportJournal("por", filepath.Join(filepath.Dir(dst), adventure.JournalDir)); err != nil {
		return err
	}
	key, err := orig.SourceKey(dataFormat, dos)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dst, "key.txt"), []byte(key), 0o644)
}

// ImportDataIfStale runs ImportData when the data is missing or was
// imported from other archives or by another format.
func ImportDataIfStale(data string) error {
	if _, err := os.Stat(filepath.Join(data, adventure.DOSDir)); errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	key, err := orig.SourceKey(dataFormat, filepath.Join(data, adventure.DOSDir))
	if err != nil {
		return err
	}
	have, err := os.ReadFile(filepath.Join(data, adventure.AssetsDir, adventure.DataDir, "key.txt"))
	if err == nil && string(have) == key {
		return nil
	}
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return ImportData(data)
}

// porAreaKind is the area record's AreaKind word (rules1e.AreaVars +0x1E4) in script addresses.
var porAreaKind = 0x49F2

// porSave is POR's SAVGAMx.DAT (ovr18 LoadGame/SaveGame, char.OriginalSave): the set, the area,
// game and aux records (rules1e's) at script addresses 0x4900, 0x6B00, 0x9700, the code image.
var porSave = rules1e.SaveLayout{
	Lead: rules1e.LeadSet,
	Records: []rules1e.SaveRecord{
		{Base: 0x4900, Cells: rules1e.AreaWords, CellBytes: 2},
		{Base: 0x6B00, Cells: rules1e.GameWords, CellBytes: 2},
		{Base: 0x9700, Cells: len(rules1e.AuxVars{}), CellBytes: 2},
	},
	CodeBase:     0x9900, // DS:49DE, ECL code at pc-0x9900
	CodeSize:     len(rules1e.OriginalSave{}.ECL),
	CodeSaved:    true,
	AreaKindAt:   &porAreaKind,
	RulesRecords: true,
}
