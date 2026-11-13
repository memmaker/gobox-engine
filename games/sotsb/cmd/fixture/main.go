// Command fixture builds a save fixture for a trace scenario
// (decomp/emu/saves/NAME): save A of SAVE/ with its party, moved into
// another area. The area's ECL block replaces the saved script, as when a
// script enters it (Pool ovr04:0BD1 opNewECL), and the party stands at x,y.
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

	"sotsb/importsotsb"
)

func main() {
	err := fixture.Run(fixture.Game{
		Root: func(data string) (string, error) {
			if data == "" {
				data = ".."
			}
			return data, nil
		},
		Open: func(root, _ string) (fixture.State, error) {
			sv, err := os.ReadFile(filepath.Join(root, "SAVE", "SAVGAMA.DAT"))
			if err != nil {
				return nil, err
			}
			vars, err := importsotsb.ScriptLayout()
			if err != nil {
				return nil, err
			}
			return fixture.OpenRawSave(sv, fixture.RawLayout{Vars: vars, Size: 0x3351, Disk: 0, Area: 1, Game: 0x801, ECL: 0x1401, ECLSize: 0x1E00, Pos: 0x3201, Count: 0x3208, Names: 0x3209, NameSize: 0x29, GameSymSet: 0x624, AreaBlock: 0x1E4, AreaGeo: 0x18A, AreaWallBlock: 0x3F4, AreaWallSlot: 0x3FA})
		},
		ECL: func(root string, block int) (int, []byte, error) {
			l, err := dax.NewLoader(root)
			if err != nil {
				return 0, nil, err
			}
			for d := 1; d <= 8; d++ { // the first disk whose ECL file has it
				if f := l.File(fmt.Sprintf("ECL%d", d)); f != nil {
					if b := f.Block(block); len(b) > 2 {
						return d, b[2:], nil
					}
				}
			}
			return 0, nil, fmt.Errorf("no ECL file has block %d", block)
		},
		Member: func(root, name, ext string) ([]byte, error) {
			return os.ReadFile(filepath.Join(root, "SAVE", name+ext))
		},
		Strong: fixture.MakeStrong,
	}, os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
}
