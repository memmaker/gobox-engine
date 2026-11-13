package importer

import (
	_ "embed"
	"encoding/json"
	"gobox/adventure"
	"gobox/asset/orig"
	"gobox/audio"
	"gobox/rules"
	"gobox/rules1e"
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
var porGameTables = adventure.GameTables{GameSet: 3, // DS:5376 = 3
	ScriptSounds: map[byte]adventure.ScriptSound{
		8:  {Cue: rules.CueStep},
		10: {Cue: rules.CueFireball},
	},
	ScriptFixes:  porScriptFixes,
	OriginalSave: porSave,
	Quests:       porQuests,
	Art:          &porArt,
	MapLabels:    porMapLabels,
	Music:        &porMusic,
	Walls:        wallTextures("por", "", ""), // the floor and sky are the areas' own (walls/por_areas.txt)
	Natives: []adventure.NativeAt{
		{ScriptAt: adventure.ScriptAt{Area: 0, Label: "TavernGamble"}, Native: adventure.NativeTavernGamble},
		{ScriptAt: adventure.ScriptAt{Area: 14, Label: "CastKnock"}, Native: adventure.NativeCastKnock},
	},
	PurseWhos: []adventure.PurseWho{
		{ScriptAt: adventure.ScriptAt{Area: 0, Label: "PayPlatinum"}, By: 0},
		{ScriptAt: adventure.ScriptAt{Area: 0, Label: "TavernGamble"}, By: -1},
	},
}

// dataFormat changes when the files PorImportDataTo writes change.
const dataFormat = "24" // 24: GameTables.Walls (walls/por.txt, por_areas.txt); 23: GameTables.GameSet; 22: no View or Aim on the combat bar (Felix 2026-11-13); 21: data/look.json (writePorLook, as PorImport); 20: GameTables.Music.Files, AmigaSongs (P1); 19: Sheet.Race is the typed race (a name), MonsterRace the record byte of a monster; 12: MON<n>.json records as the rules' Sheet (char.SheetOfRecord); 13: GameTables.Art (Pg); 14: no Restoration tip (Rz); 15: GameTables.MapLabels, Music (Ph); 16: script fixes gold only (Gc); 17: area 0 wager text in gold (Df)

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
	snd, err := audio.DecodeSoundData(exe, porSoundLayout)
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

// porSoundLayout is where START.EXE's driver (segment 08AE) keeps its tables (gobox audio API.md).
var porSoundLayout = audio.SoundDataLayout{Segment: 0x8AE, JumpTable: 545, SpeakerSet: 557, TandySet: 967,
	SpeakerSongs: 1375, TandySongs: 1543, Songs: 21, Enabled: 525,
	Commands: map[int]string{5710: "reset", 5804: "return", 5811: "call", 5701: "switch", 5678: "loop", 5759: "set"}}
