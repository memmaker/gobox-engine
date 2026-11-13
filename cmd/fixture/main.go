// Command fixture builds a save fixture: save A of SAVE/ (por: a named slot)
// with its party, moved into another area. The area's ECL block is the script
// the load runs, as when a script enters it (ovr04:0BD1 opNewECL), and the
// party stands at x,y.
//
//	fixture -game por|cotab|sotsb -out DIR [-data DIR] [-block N -walls A,B,C [-geo G]] [-x X -y Y -dir D] [-set FIELD=VAL ...] [-strong]
//
// -walls are the WALLDEF blocks of wall slots 1..3 as SaveGame records them
// (255 none; a block of two chunks fills its slot and the next, which is
// then 255), as the area's LOAD PIECES leaves them. LoadGame reloads them.
// -set sets a script field, by its layout name, element 0. -strong
// makes every member hit always and take a long time to die, so a scripted Quick fight ends in a win.
// The other flags are gobox/fixture's (see its API.md).
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gobox/asset/orig"
	eclscript "gobox/ecl/script"
	"gobox/engine"
	"gobox/fixture"
	"gobox/importer/games"
	"gobox/rules1e"
)

// rawLayout is the byte layout of SAVGAMA.DAT shared by cotab and sotsb.
var rawLayout = fixture.RawLayout{Size: 0x3351, Disk: 0, Area: 1, Game: 0x801, ECL: 0x1401, ECLSize: 0x1E00,
	Pos: 0x3201, Count: 0x3208, Names: 0x3209, NameSize: 0x29, GameSymSet: 0x624,
	AreaBlock: 0x1E4, AreaGeo: 0x18A, AreaWallBlock: 0x3F4, AreaWallSlot: 0x3FA}

// rawScriptLayouts names the script fields -set sets, per raw-save game.
var rawScriptLayouts = map[string]func() (*eclscript.Layout, error){
	"cotab": games.CotabScriptLayout,
	"sotsb": games.SotsbScriptLayout,
}

// eclBlock is the first disk's ECL file that has block, without its 2-byte header.
func eclBlock(l *orig.Loader, block int, shared bool) (int, []byte, error) {
	for d := 1; d <= 8; d++ {
		name := fmt.Sprintf("ECL%d", d)
		var b []byte
		if shared {
			b = l.BlockShared(name, block)
		} else {
			b = l.Block(name, block)
		}
		if len(b) > 2 {
			return d, b[2:], nil
		}
	}
	return 0, nil, fmt.Errorf("no ECL file has block %d", block)
}

// raw is the fixture of a game whose original SAVE/SAVGAMA.DAT is edited
// byte-wise (cotab, sotsb): the data folder is the parent of the game folder.
func raw(id string) fixture.Game {
	return fixture.Game{
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
			vars, err := rawScriptLayouts[id]()
			if err != nil {
				return nil, err
			}
			layout := rawLayout
			layout.Vars = vars
			return fixture.OpenRawSave(sv, layout)
		},
		ECL: func(root string, block int) (int, []byte, error) {
			l, err := orig.NewLoader(root)
			if err != nil {
				return 0, nil, err
			}
			return eclBlock(l, block, false)
		},
		Member: func(root, name, ext string) ([]byte, error) {
			return os.ReadFile(filepath.Join(root, "SAVE", name+ext))
		},
		Strong: fixture.MakeStrong,
	}
}

// porGame is the fixture of por, whose canonical named slots are edited.
func porGame() fixture.Game {
	var slotDir string // the slot Open read: its members are beside its save
	return fixture.Game{
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
			slotDir = filepath.Join(root, engine.SaveDir, "saves", slot)
			s, err := engine.OpenSave(rules1e.Rules1e, slotDir, filepath.Join(root, engine.AssetsDir))
			if err != nil {
				return nil, fmt.Errorf("slot %s: %v", slot, err)
			}
			return s, nil
		},
		ECL: func(root string, block int) (int, []byte, error) {
			l, err := orig.NewLoader(filepath.Join(root, engine.DOSDir))
			if err != nil {
				return 0, nil, err
			}
			return eclBlock(l, block, true)
		},
		Member: func(_, name, ext string) ([]byte, error) {
			return os.ReadFile(filepath.Join(slotDir, name+ext))
		},
		Strong: engine.MakeStrong,
	}
}

// gamesByID maps -game to its fixture.
var gamesByID = map[string]func() fixture.Game{
	"por":   porGame,
	"cotab": func() fixture.Game { return raw("cotab") },
	"sotsb": func() fixture.Game { return raw("sotsb") },
}

// splitGame removes -game ID from args (fixture.Run knows no such flag).
func splitGame(args []string) (id string, rest []string) {
	for i := 0; i < len(args); i++ {
		a := strings.TrimLeft(args[i], "-")
		switch {
		case a == "game" && i+1 < len(args):
			id = args[i+1]
			i++
		case strings.HasPrefix(a, "game="):
			id = strings.TrimPrefix(a, "game=")
		default:
			rest = append(rest, args[i])
		}
	}
	return id, rest
}

func main() {
	id, args := splitGame(os.Args[1:])
	mk, ok := gamesByID[id]
	if !ok {
		log.Fatalf("usage: fixture -game por|cotab|sotsb ... (got -game %q)", id)
	}
	if err := fixture.Run(mk(), args); err != nil {
		log.Fatal(err)
	}
}
