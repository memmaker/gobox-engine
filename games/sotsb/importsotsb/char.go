package importsotsb

import (
	"gobox/char"
)

// charRecord is the game's character record, the shared 439-byte layout
// (gobox/char, field for field: CHRDATA/SAVGAM and the monster records).
type charRecord = char.Record439

// RecordLayout is the layout of the game's character records (its adventure's manifest names it).
const RecordLayout = char.Layout439
