// Command fixture builds a save fixture for a trace scenario
// (decomp/emu/saves/NAME): save A of SAVE/ with its party, moved into
// another area (gobox/fixture; see its API.md for the flags).
//
//	fixture -out DIR [-data ..] [-block N -walls A,B,C [-geo G]] [-x X -y Y -dir D] [-set FIELD=VAL ...] [-strong]
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gobox/asset/orig"
	"gobox/fixture"

	"gobox/games/cotab/importcotab"
)

func main() {
	err := fixture.Run(fixture.Game{
		Root: func(data string) (string, error) {
			if data == "" {
				return "..", nil
			}
			return data, nil
		},
		Open: func(root, _ string) (fixture.State, error) {
			sv, err := os.ReadFile(filepath.Join(root, "SAVE", "SAVGAMA.DAT"))
			if err != nil {
				return nil, err
			}
			vars, err := importcotab.ScriptLayout()
			if err != nil {
				return nil, err
			}
			return fixture.OpenRawSave(sv, fixture.RawLayout{Vars: vars, Size: 0x3351, Disk: 0, Area: 1, Game: 0x801,
				ECL: 0x1401, ECLSize: 0x1E00, Pos: 0x3201, Count: 0x3208, Names: 0x3209, NameSize: 0x29,
				GameSymSet: 0x624, AreaBlock: 0x1E4, AreaGeo: 0x18A, AreaWallBlock: 0x3F4, AreaWallSlot: 0x3FA})
		},
		ECL: func(root string, block int) (int, []byte, error) {
			l, err := orig.NewLoader(root)
			if err != nil {
				return 0, nil, err
			}
			for d := 1; d <= 8; d++ {
				if b := l.Block(fmt.Sprintf("ECL%d", d), block); len(b) > 2 {
					return d, b[2:], nil
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
