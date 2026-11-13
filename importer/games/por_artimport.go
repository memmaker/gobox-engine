package games

import (
	"path/filepath"

	"gobox/adventure"
	"gobox/asset/orig"
)

// ImportArt is the art import: data's DOS graphics archives, and the Amiga
// ones standing in for them, as sheets in data/assets/art (dax.ImportArt).
// Nothing at run time reads the archives' art.
func ImportArt(data string) error {
	return orig.ImportArt(filepath.Join(data, adventure.DOSDir), filepath.Join(data, adventure.AmigaDir), filepath.Join(data, adventure.AssetsDir, adventure.ArtDir))
}

// ImportArtIfStale runs ImportArt when the art is missing or was imported
// from other archives or by another format (dax.ImportArtIfStale).
func ImportArtIfStale(data string) error {
	return orig.ImportArtIfStale(filepath.Join(data, adventure.DOSDir), filepath.Join(data, adventure.AmigaDir), filepath.Join(data, adventure.AssetsDir, adventure.ArtDir))
}

// ImportAssets is the assets import: the art (ImportArt) and the data (ImportData).
func ImportAssets(data string) error {
	if err := ImportArt(data); err != nil {
		return err
	}
	return ImportData(data)
}

// ImportAssetsIfStale runs the art and the data import where stale.
func ImportAssetsIfStale(data string) error {
	if err := ImportArtIfStale(data); err != nil {
		return err
	}
	return ImportDataIfStale(data)
}
