package importer

import (
	"gobox/rules1e"
)

// charRecord is the game's character record, the shared 439-byte layout
// (gobox/char, field for field: CHRDATA/SAVGAM and the monster records).
type sotsbCharRecord = rules1e.Record439

// RecordLayout is the layout of the game's character records (its adventure's manifest names it).
const SotsbRecordLayout = rules1e.Layout439
