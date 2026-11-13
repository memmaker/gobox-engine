package importcotab

import (
	"errors"
	"gobox/rules1e"
	"io/fs"
	"os"
	"path/filepath"

	"gobox/asset"
	"gobox/char"
	"gobox/dax"
)

// ImportTo writes the adventure folder dst from the DOS originals in dos
// (gobox import cotab, importcotab): the decompiled scripts with the game's
// fixes, the maps and records, the picture table, the art
// sheets (the Amiga's from a folder amiga beside dos, when there is one),
// the shipped characters and saves (SAVE) and the Amiga's sampled effects.
func ImportTo(dos, dst string) error {
	if err := ImportScriptsTo(dos, filepath.Join(dst, scriptsDir)); err != nil {
		return err
	}
	if err := ImportDataTo(dos, filepath.Join(dst, dataDir)); err != nil {
		return err
	}
	if err := asset.WritePictures("cotab", dst); err != nil {
		return err
	}
	amiga := filepath.Join(filepath.Dir(dos), "amiga")
	if err := dax.ImportArt(dos, amiga, filepath.Join(dst, artDir)); err != nil {
		return err
	}
	if err := copyIfThere(filepath.Join(dst, saveDir), os.DirFS(filepath.Join(dos, "SAVE"))); err != nil {
		return err
	}
	glb, err := os.ReadFile(filepath.Join(amiga, AmigaMusic.Effects))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dst, soundDir), 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dst, soundDir, AmigaMusic.Effects), glb, 0o644)
}

// copyIfThere copies src into dir; a src that is not there copies nothing.
func copyIfThere(dir string, src fs.FS) error {
	if _, err := fs.Stat(src, "."); errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return os.CopyFS(dir, src)
}

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
	if err := dax.ImportData[charRecord, affectRecord, Item](l, dst, char.SheetOfRecord[charRecord], char.ItemOfRecord[Item]); err != nil {
		return err
	}
	return dax.ImportItemTable[rules1e.ItemBase](l, dst, itemTypes, dax.Counted)
}

// SoundFile is the sound tables in the data dir (gobox audio's JSON,
// importcotab writes it: the decoder is the import's, not the engine's).
const SoundFile = "sound.json"
