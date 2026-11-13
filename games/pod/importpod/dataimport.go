package importpod

import (
	"gobox/asset"
	"gobox/asset/orig"
	"gobox/rules1e"
	"os"
	"path/filepath"
)

// ImportDataTo writes the game's maps and records from the DOS originals in
// dos into dst (gobox/dax ImportData over this game's record types) and
// writes the base item table ITEMS as dax.ItemTableFile (as POR's import).
func ImportDataTo(dos, dst string) error {
	l, err := orig.NewLoader(dos)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	if err := orig.ImportData[charRecord, affectRecord, libraryRecord](l, dst, rules1e.SheetOfRecord[charRecord], rules1e.ItemOfRecord[libraryRecord]); err != nil {
		return err
	}
	return orig.ImportItemTable[rules1e.ItemBase](l, dst, itemTypes, orig.Bare)
}

// The originals' sound and animation files the folder carries as they are
// (gobox reads them from sound/ and anim/: game.Data.SoundDir, AnimDir):
// the AdLib songs and effects, the animations' songs, the VACE animations.
var (
	soundFiles = []string{adLibMusic, adLibSounds, "P1.SMS", "P3.SMS"}
	animFiles  = []string{"INTRO.ACF", "INTRO.BMA", "FINAL.ACF", "FINAL.BMA"}
)

// The AdLib driver's archives (game.AdLib).
const (
	adLibMusic  = "MUSIC.DAX"
	adLibSounds = "ZOUNDS.DAX"
)

// ImportTo writes the adventure folder dst from the DOS originals in dos
// (gobox import pod, importpod): the decompiled scripts with the game's
// fixes, the maps and records, the picture table, the art sheets (the
// Amiga's from a folder amiga beside dos, when there is one), the shipped
// characters (SAVE), the sound and animation files.
func ImportTo(dos, dst string) error {
	if err := ImportScriptsTo(dos, filepath.Join(dst, scriptsDir)); err != nil {
		return err
	}
	if err := ImportDataTo(dos, filepath.Join(dst, dataDir)); err != nil {
		return err
	}
	if err := asset.WritePictures("pod", dst); err != nil {
		return err
	}
	if err := orig.ImportArt(dos, filepath.Join(filepath.Dir(dos), "amiga"), filepath.Join(dst, artDir)); err != nil {
		return err
	}
	if save := filepath.Join(dos, "SAVE"); exists(save) {
		if err := os.CopyFS(filepath.Join(dst, savesDir), os.DirFS(save)); err != nil {
			return err
		}
	}
	if err := copyFiles(dos, filepath.Join(dst, soundDir), soundFiles); err != nil {
		return err
	}
	if amiga := filepath.Join(filepath.Dir(dos), "amiga"); exists(filepath.Join(amiga, AmigaMusic.Songs)) {
		if err := copyFiles(amiga, filepath.Join(dst, soundDir), []string{AmigaMusic.Songs, AmigaMusic.Effects}); err != nil {
			return err
		}
	}
	return copyFiles(dos, filepath.Join(dst, animDir), animFiles)
}

// exists reports whether path is there.
func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// copyFiles copies names from src into dir; each must be there.
func copyFiles(src, dir string, names []string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for _, n := range names {
		b, err := os.ReadFile(filepath.Join(src, n))
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, n), b, 0o644); err != nil {
			return err
		}
	}
	return nil
}
