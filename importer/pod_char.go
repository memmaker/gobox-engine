package importer

import (
	"gobox/rules1e"
	"strconv"
)

// charRecord is the game's character record, the shared 510-byte layout
// (gobox/char, field for field: CHRDATA/SAVGAM and the monster records).
type podCharRecord = rules1e.Record510

// RecordLayout is the layout of the game's character records (its adventure's manifest names it).
const PodRecordLayout = rules1e.Layout510

// PodArt is the opaque art IDs gobox picked for a character, head and body
// separately; the rules store them and never interpret them.
type PodArt struct{ PortraitHead, PortraitBody, IconHead, IconBody string }

// IntToStr = O0EC:14A5.
func itoa(n int) string { return strconv.Itoa(n) }
