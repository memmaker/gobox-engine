package orig

import (
	"fmt"
	"io/fs"

	"gobox/ecl"
)

// DAXSource reads the blocks of the game's ECL*.DAX, for the import.
func DAXSource(l *Loader) ecl.Source {
	return ecl.Source{Block: func(disk, block int) []byte {
		name := fmt.Sprintf("ECL%d", disk)
		if !l.Exists(name + ".DAX") {
			return nil
		}
		return l.BlockShared(name, block)
	}}
}

// ImportDAXFolder decompiles the ECL*.DAX in dos into an adventure folder's
// scripts dir (ecl.ImportFolder with the DAX source and gate folding): the
// games' importer step.
func ImportDAXFolder(v *ecl.Version, dos string, data fs.FS, dir string, fixes []ecl.Fix) error {
	l, err := NewLoader(dos)
	if err != nil {
		return err
	}
	return ecl.ImportFolder(v, DAXSource(l), data, dir, ecl.FoldGates, fixes)
}
