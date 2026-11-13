package importpod

import (
	"gobox/asset"
	"gobox/char"
	"gobox/dax"
	"path/filepath"
)

// The adventure folder's parts (gobox import pod writes them).
const (
	scriptsDir = "scripts"
	artDir     = "art"
	dataDir    = "data"
	soundDir   = "sound"
	animDir    = "anim"
	savesDir   = "saves"
	secretDir  = "secret" // the shipped SECRET folder: Silver Blades' saves to bring over
)

// libraryRecord is an item as ITEM0 block 1 and MON<n>ITM hold it: the
// last libraryItemSize bytes of an itemRecord (libraryItem).
type libraryRecord [libraryItemSize]byte

// Folder is what the engine reads of an imported adventure folder: its art
// sheets and its maps and records.
type Folder struct {
	Art  *asset.Art
	Data *dax.Data[char.Sheet, affectRecord, char.Item] // as dax.ImportData writes them
}

// OpenFolder is the adventure folder adv.
func OpenFolder(adv string) Folder {
	return Folder{Art: asset.OpenArt(filepath.Join(adv, artDir)), Data: dax.OpenData[char.Sheet, affectRecord, char.Item](filepath.Join(adv, dataDir))}
}
