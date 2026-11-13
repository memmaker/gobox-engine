package importsotsb

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"rules1e"

	"gobox/char"
	"gobox/dax"
	"gobox/exe"
	"gobox/exe/tablespec"
	"gobox/importer/pictures"
	"gobox/pool"
	"sotsb"
)

// ImportDataTo writes the game's maps and records from the DOS originals in
// dos into dst (gobox/dax ImportData over this game's record types) and
// writes the base item table ITEMS as dax.ItemTableFile (as POR's import).
func ImportDataTo(dos, dst string) error {
	l, err := dax.NewLoader(dos)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	if err := dax.ImportData[charRecord, affectRecord, itemRecord](l, dst, char.SheetOfRecord[charRecord], char.ItemOfRecord[itemRecord]); err != nil {
		return err
	}
	return dax.ImportItemTable[rules1e.ItemBase](l, dst, itemTypes, dax.Counted)
}

// ImportTo writes the adventure folder dst from the DOS originals in dos
// (gobox import sotsb, importsotsb): the decompiled scripts with the game's
// fixes, the maps and records, the picture table, the art sheets (the
// Amiga's from a folder amiga beside dos, when there is one), the shipped
// characters and saves, the AdLib music and the Amiga's songs and effects.
func ImportTo(dos, dst string) error {
	if err := ImportScriptsTo(dos, filepath.Join(dst, scriptsDir)); err != nil {
		return err
	}
	if err := ImportDataTo(dos, filepath.Join(dst, dataDir)); err != nil {
		return err
	}
	if err := pictures.WritePictures("sotsb", dst); err != nil {
		return err
	}
	if err := ImportRulesTo(dos, filepath.Join(dst, dataDir)); err != nil {
		return err
	}
	if err := ImportGameTablesTo(filepath.Join(dst, dataDir)); err != nil {
		return err
	}
	if err := ImportJournalTo(filepath.Join(dst, journalDir)); err != nil {
		return err
	}
	amiga := filepath.Join(filepath.Dir(dos), "amiga")
	if err := dax.ImportArt(dos, amiga, filepath.Join(dst, artDir)); err != nil {
		return err
	}
	if err := copyFiles(dos, filepath.Join(dst, savesDir), savePatterns...); err != nil {
		return err
	}
	if err := copyFiles(filepath.Join(dos, "SAVE"), filepath.Join(dst, savesDir), savePatterns...); err != nil {
		return err
	}
	if err := copyFiles(dos, filepath.Join(dst, soundDir), "SOUND.DAX"); err != nil {
		return err
	}
	return copyFiles(amiga, filepath.Join(dst, soundDir), AmigaMusic.Songs, AmigaMusic.Effects)
}

// savePatterns are the shipped characters and saved games (rules/charfiles.go:
// NAME.WHO, .STF, .SFX, .IDS; CHRDATAn.SAV; SAVGAMx.DAT).
var savePatterns = []string{"*.WHO", "*.STF", "*.SFX", "*.IDS", "*.SAV", "SAVGAM*.DAT"}

// copyFiles copies the files of src matching patterns into dst; a src
// without them copies nothing.
func copyFiles(src, dst string, patterns ...string) error {
	for _, pat := range patterns {
		m, err := filepath.Glob(filepath.Join(src, pat))
		if err != nil {
			return err
		}
		for _, f := range m {
			b, err := os.ReadFile(f)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(dst, 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(dst, filepath.Base(f)), b, 0o644); err != nil {
				return err
			}
		}
	}
	return nil
}

// ImportRulesTo writes START.EXE's tables (tables.json, data segment paragraph
// 3554) into dst/rules.json (tablespec) with the importer's own entries:
// the affect ids and the creation start (rules1e.Tables.WithRules, laid out
// as Rules1eLayout).
func ImportRulesTo(dos, dst string) error {
	l, err := dax.NewLoader(dos)
	if err != nil {
		return err
	}
	im, err := exe.Load(l.Path("START.EXE"), 3554)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	return tablespec.WriteRulesWith(sotsb.TablesJSON, dst, []tablespec.Extra{{Name: "affectIds", Value: rules1e.SotsbAffectIDs}, {Name: "spellIds", Value: rules1e.SotsbSpellIDs}, {Name: "itemSpellCode", Value: rules1e.SevenBitItemSpell}, {Name: "spellClasses", Value: rules1e.SotsbSpellClasses}, {Name: "affectVariants", Value: rules1e.SotsbAffectVariants}, {Name: "affectGroups", Value: rules1e.SotsbAffectGroups}, {Name: "giantRule", Value: rules1e.LaterGiants}, {Name: "magicResistance", Value: rules1e.SotsbMagicResist}, {Name: "oldAge", Value: rules1e.SotsbOldAge}, {Name: "casterLevel", Value: rules1e.SotsbCasterLevel}, {Name: "itemPowers", Value: rules1e.SotsbItemPowers}, {Name: "creationStart", Value: CreationStart()}, {Name: "town", Value: Town()}, {Name: "combatGround", Value: CombatGround()}, {Name: "groundIds", Value: rules1e.SotsbGroundIDs}, {Name: "itemKinds", Value: rules1e.SotsbItemKinds}, {Name: "spiritualHammer", Value: rules1e.SotsbHammer}, {Name: "randomTreasure", Value: rules1e.SotsbTreasure()}, {Name: "appraisal", Value: rules1e.SotsbAppraisal}}, im)
}

// ImportGameTablesTo writes the game's data (pool.GameDataFile), the Engine's game tables (pool.GameTablesFile)
// and the tables' layout (rules1e.LayoutFile) into dst. SotSB has no
// portraits of its own (portraits are pool picks, sotsbpics.go); an NPC
// loads its combat icon; its scripts have no Sound command, the duel sounds
// are CALL 45571's (callSound: soundDuelA a step, soundDuelB the fireball,
// any other a step).
func ImportGameTablesTo(dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	if err := pool.WriteGameData(dst, StaticData()); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dst, pool.SoundFile), soundData, 0o644); err != nil { // START.EXE's tables, as pool.SoundData reads them
		return err
	}
	for file, v := range map[string]any{pool.GameTablesFile: GameTables, rules1e.LayoutFile: Layout1e} {
		b, err := json.MarshalIndent(v, "", " ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dst, file), b, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// ImportJournalTo writes the Adventurer's Journal's text and pictures (the
// built-in journal/, gobox's journal format) into dst; the entries that
// name them are the scripts' names/entries.json.
func ImportJournalTo(dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	return fs.WalkDir(journalFS, "journal", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := journalFS.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dst, filepath.Base(p)), b, 0o644)
	})
}
