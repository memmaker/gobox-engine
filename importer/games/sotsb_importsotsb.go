package games

import (
	"os"
	"path/filepath"

	"gobox/asset/orig"
	"gobox/audio"
	"gobox/importer"
)

// init registers Secret of the Silver Blades's import with gobox/importer: the
// decompiled scripts with the game's fixes, the art sheets, and the maps and records into
// adventures/sotsb/{scripts,art,data}. The Amiga art is taken from a folder amiga
// beside the originals when there is one (data/dos, data/amiga).
func init() {
	importer.Register(importer.Game{
		ID: "sotsb", Title: "Secret of the Silver Blades", Release: 1990, Ruleset: "rules1e", Record: SotsbRecordLayout, Version: 12, // 12: Window, UserSaves, Foreign in the manifest (playpool reads them); 11: 1: imported for pool (Sw-sotsb); 2: MON records as the rules' Sheet (Pf); 3: GameTables.Art, items as char.Item (Pg); 4: Hold tips (Rz); 5: groundIds, itemKinds, spiritualHammer (Ph); 6: randomTreasure, appraisal (Pi); 7: scripts gold only (Gc); 8: randomTreasure from rules1e's treasure/sotsb.json (Ct); 9: area 16 gem trade 1 per 500 gold (Df); 10: Word7B90, Word7BC2 16-cell text arrays (Tb); 11: no Sequence.Remaster (the remaster credit line is gobox's), SequenceStep.Credits on the Amiga card
		Window: importer.Size{W: 80 * 8, H: 50 * 8}, UserSaves: true, Foreign: []string{"cotab"},
		Sources: []string{"*.DAX", "ITEMS", "START.EXE", "GAME.OVR"},
		Import:  SotsbImport, Amiga: &sotsbAmiga,
		Start: importer.Fixed(sotsbStart()),
	})
}

// whole is the 40x25 page.
var sotsbWhole = importer.Cells{X2: 39, Y2: 24}

// start is SotSB's DOS title (GAME.OVR 0x644 title_screen, rules/title.go):
// TITLE 1 (SSI) with the title song (titleSongN 2), the credits page (0x36e,
// Credits), the painting (TITLE 2), the painting with the Silver Blades
// logo (TITLE 3 at cell 2, 5); the painting with the Forgotten Realms banner
// (TITLE 4 at cell 5, 5) stays under the main menu. With the Amiga art up
// the port's credits card (the Amiga TITLE 5, 256x161, centred) follows the
// credits page (EditionAmiga).
func sotsbStart() *importer.Sequence {
	painting := importer.Layer{Art: importer.Art{File: "TITLE", Block: 2}, At: sotsbWhole}
	q := &importer.Sequence{
		Steps: []importer.SequenceStep{
			{Kind: importer.StepPicture, Layers: []importer.Layer{{Art: importer.Art{File: "TITLE", Block: 1}, At: sotsbWhole}}, Song: "2"},
			{Kind: importer.StepCredits},
			{Kind: importer.StepPicture, Layers: []importer.Layer{{Art: importer.Art{File: "TITLE", Block: 5}, At: importer.Cells{X1: 4, Y1: 2, X2: 35, Y2: 22}, Centred: true}}, Edition: importer.EditionAmiga, Credits: true},
			{Kind: importer.StepPicture, Layers: []importer.Layer{painting}},
			{Kind: importer.StepPicture, Layers: []importer.Layer{painting, {Art: importer.Art{File: "TITLE", Block: 3}, At: importer.Cells{X1: 2, Y1: 5, X2: 37, Y2: 15}}}},
		},
		Backdrop: []importer.Layer{painting, {Art: importer.Art{File: "TITLE", Block: 4}, At: importer.Cells{X1: 5, Y1: 5, X2: 35, Y2: 15}}},
	}
	for _, c := range Credits {
		q.Credits = append(q.Credits, importer.Credit{X: c.X, Y: c.Y, Colour: c.Colour, Text: c.Text, Page: c.Page})
	}
	return q
}

// Import writes the folder dst from the DOS originals in dos (ImportTo)
// and the sound tables.
func SotsbImport(dos, dst string) error {
	if err := SotsbImportTo(dos, dst); err != nil {
		return err
	}
	if err := sotsbImportSound(dos, filepath.Join(dst, "data")); err != nil {
		return err
	}
	return nil
}

// SoundFile is the sound tables in the data dir (gobox audio's JSON).
const SotsbSoundFile = "sound.json"

// importSound decodes START.EXE's sound tables into dst/SoundFile.
func sotsbImportSound(dos, dst string) error {
	l, err := orig.NewLoader(dos)
	if err != nil {
		return err
	}
	exe, err := os.ReadFile(l.Path("START.EXE"))
	if err != nil {
		return err
	}
	snd, err := audio.DecodeSoundData(exe, sotsbSoundLayout)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dst, SotsbSoundFile), snd, 0o644)
}

// soundLayout is where START.EXE's driver keeps its tables (gobox audio API.md).
var sotsbSoundLayout = audio.SoundDataLayout{Notes: true, Segment: 0x7E4, JumpTable: 354, SpeakerSet: 366, TandySet: 776,
	SpeakerSongs: 1184, TandySongs: 1288, Songs: 13, Enabled: 334, SpeakerPeriods: 750, TandyPeriods: 1160,
	Commands: map[int]string{14112: "reset", 14206: "return", 14213: "call", 14103: "switch", 14080: "loop", 14161: "set"}}
