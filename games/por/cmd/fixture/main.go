// Command fixture builds a save fixture: save A of SAVE/ with its party, moved into
// another area. The area's ECL block is the script the load runs, as when a
// script enters it (ovr04:0BD1 opNewECL), and the party stands at x,y.
//
//	fixture -out DIR [-data DIR] [-block N -walls A,B,C [-geo G]] [-x X -y Y -dir D] [-set FIELD=VAL ...] [-strong]
//
// -walls are the WALLDEF blocks of wall slots 1..3 as SaveGame records them
// (255 none; a block of two chunks fills its slot and the next, which is
// then 255), as the area's LOAD PIECES leaves them. LoadGame reloads them.
// -set sets a script field, by its layout name, element 0. -strong
// makes every member hit always and take a long time to die, so a scripted Quick fight ends in a win.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gobox/dax"
	"gobox/fixture"
	"gobox/pool"
	"por/porgame"
)

func main() {
	var slotDir string // the slot Open read: its members are beside its save
	err := fixture.Run(fixture.Game{
		NeedSlot: true,
		Root: func(data string) (string, error) {
			if data != "" { // default: data/por/sources in the nearest folder above holding data/por/sources/dos/TITLE.DAX
				return data, nil
			}
			wd, _ := os.Getwd()
			for d := wd; d != filepath.Dir(d); d = filepath.Dir(d) {
				if _, err := os.Stat(filepath.Join(d, "data", "por", "sources", "dos", "TITLE.DAX")); err == nil {
					return filepath.Join(d, "data", "por", "sources"), nil
				}
			}
			return "", nil
		},
		Open: func(root, slot string) (fixture.State, error) {
			slotDir = filepath.Join(root, pool.SaveDir, "saves", slot)
			s, err := porgame.OpenSave(slotDir, filepath.Join(root, pool.AssetsDir))
			if err != nil {
				return nil, fmt.Errorf("slot %s: %v", slot, err)
			}
			return s, nil
		},
		ECL: func(root string, block int) (int, []byte, error) {
			l, err := dax.NewLoader(filepath.Join(root, pool.DOSDir))
			if err != nil {
				return 0, nil, err
			}
			for d := 1; d <= 8; d++ {
				if b := l.BlockShared(fmt.Sprintf("ECL%d", d), block); len(b) > 2 {
					return d, b[2:], nil
				}
			}
			return 0, nil, fmt.Errorf("no ECL file has block %d", block)
		},
		Member: func(_, name, ext string) ([]byte, error) {
			return os.ReadFile(filepath.Join(slotDir, name+ext))
		},
		Strong: pool.MakeStrong,
	}, os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
}
