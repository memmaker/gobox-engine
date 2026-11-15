package importer

import (
	"errors"
	"gobox/adventure"
	"gobox/rules1e"
	"io/fs"
	"os"
	"path/filepath"

	"gobox/asset"
	"gobox/asset/orig"
)

// ImportTo writes the adventure folder dst from the DOS originals in dos
// (gobox import cotab, the cotab import): the decompiled scripts with the game's
// fixes, the maps and records, the picture table, the art
// sheets (the Amiga's from a folder amiga beside dos, when there is one),
// the shipped characters and saves (SAVE) and the Amiga's sampled effects.
func CotabImportTo(dos, dst string) error {
	if err := CotabImportScriptsTo(dos, filepath.Join(dst, scriptsDir)); err != nil {
		return err
	}
	if err := CotabImportDataTo(dos, filepath.Join(dst, adventure.DataDir)); err != nil {
		return err
	}
	if err := asset.WritePictures("cotab", dst); err != nil {
		return err
	}
	amiga := filepath.Join(filepath.Dir(dos), "amiga")
	if err := orig.ImportArt(dos, amiga, filepath.Join(dst, adventure.ArtDir)); err != nil {
		return err
	}
	if err := copyIfThere(filepath.Join(dst, saveDir), os.DirFS(filepath.Join(dos, "SAVE"))); err != nil {
		return err
	}
	glb, err := os.ReadFile(filepath.Join(amiga, CotabAmigaMusic.Effects))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dst, adventure.SoundDir), 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dst, adventure.SoundDir, CotabAmigaMusic.Effects), glb, 0o644)
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
func CotabImportDataTo(dos, dst string) error {
	l, err := orig.NewLoader(dos)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	if err := orig.ImportData[cotabCharRecord, affectRecord, CotabItem](l, dst, rules1e.SheetOfRecord[cotabCharRecord], rules1e.ItemOfRecord[CotabItem]); err != nil {
		return err
	}
	return orig.ImportItemTable[rules1e.ItemBase](l, dst, rules1e.ItemTypes, orig.Counted)
}

// SoundFile is the sound tables in the data dir (gobox audio's JSON,
// the cotab import writes it: the decoder is the import's, not the engine's).
const CotabSoundFile = "sound.json"
