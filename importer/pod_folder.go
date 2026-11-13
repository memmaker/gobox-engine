package importer

import (
	"gobox/adventure"
	"gobox/asset"
	"gobox/asset/orig"
	"gobox/rules1e"
	"path/filepath"
)

// The adventure folder's parts (gobox import pod writes them).
const (
	animDir   = "anim"
	savesDir  = "saves"
	secretDir = "secret" // the shipped SECRET folder: Silver Blades' saves to bring over
)

// libraryRecord is an item as ITEM0 block 1 and MON<n>ITM hold it: the
// last libraryItemSize bytes of an itemRecord (libraryItem).
type libraryRecord [libraryItemSize]byte

// PodFolder is what the engine reads of an imported adventure folder: its art
// sheets and its maps and records.
type PodFolder struct {
	Art  *asset.Art
	Data *orig.Data[rules1e.Sheet, affectRecord, rules1e.Item] // as dax.ImportData writes them
}

// OpenPodFolder is the adventure folder adv.
func OpenPodFolder(adv string) PodFolder {
	return PodFolder{Art: asset.OpenArt(filepath.Join(adv, adventure.ArtDir)), Data: orig.OpenData[rules1e.Sheet, affectRecord, rules1e.Item](filepath.Join(adv, adventure.DataDir))}
}
