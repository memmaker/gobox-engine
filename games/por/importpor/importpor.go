// Package importpor registers POR's import with gobox/importer: the
// decompiled scripts with POR's fixes, the art sheets and the data (maps,
// monsters, items, POOL.CFG, START.EXE's sound tables) into
// adventures/por/{scripts,art,data}. The Amiga art is taken from a folder
// amiga beside the originals when there is one (POR's data/dos, data/amiga).
package importpor

import (
	"gobox/importer/pictures"
	"os"
	"path/filepath"

	"gobox/dax"
	"gobox/importer"
	"gobox/pool"
)

func init() {
	importer.Register(importer.Game{
		ID: "por", Title: "Pool of Radiance", Ruleset: "rules1e", Record: pool.RecordLayout,
		Sources: []string{"*.DAX", "ITEMS", "POOL.CFG", "START.EXE"},
		Import:  Import,
		Start:   importer.Fixed(start()),
		End:     importer.Fixed(end()),
	})
}

// whole is the 40x25 page.
var whole = importer.Cells{X1: 0, Y1: 0, X2: 39, Y2: 24}

// start is POR's title (ovr02:0626 title_screen, as pool.Engine.State's
// SceneTitle builds it): TITLE 1 with the Amiga intro's song (titleSong 1),
// TITLE 2 with its lightning, the credits page (ovr02:0206: credits in
// its two boxes); a short intro (POOL.CFG's, Intro.Full off) is TITLE 1
// only. The Amiga painting art/TITLE.PNG lies under the main menu.
func start() *importer.Sequence {
	q := &importer.Sequence{
		Steps: []importer.SequenceStep{
			{Kind: importer.StepPicture, Layers: []importer.Layer{{Art: importer.Art{File: "TITLE", Block: 1}, At: whole}}, Song: "1", Whole: true},
			{Kind: importer.StepPicture, Layers: []importer.Layer{{Art: importer.Art{File: "TITLE", Block: 2}, At: whole}}, Whole: true, Lightning: importer.FlashEGA},
			{Kind: importer.StepCredits, Page: 0},
		},
		Boxes:    []importer.Cells{{X1: 1, Y1: 1, X2: 38, Y2: 3}, {X1: 1, Y1: 5, X2: 38, Y2: 23}},
		Backdrop: []importer.Layer{{Art: importer.Art{File: "TITLE.PNG", Kind: importer.ArtPNG}, At: whole}},
		Short:    []int{0},
		Full:     true,                                                                        // POOL.CFG's default
		Remaster: [2]string{"Opinionated Remaster", "directed by Felix Ruzzoli using Claude"}, // extra: centred, yellow, from three rows over the prompt row, on a page tall enough for them
		Credits:  credits,
	}
	return q
}

// credits is ovr02:0206's page: each Text at cell X, Y in EGA colour Colour
// (DisplayString's way, upcased).
var credits = []importer.Credit{
	{X: 5, Y: 1, Colour: 0xa, Text: "scenario created by:"},
	{X: 0x1a, Y: 1, Colour: 0xe, Text: "tsr, inc."},
	{X: 0xf, Y: 2, Colour: 0xa, Text: "jim ward"},
	{X: 2, Y: 3, Colour: 0xa, Text: "david cook,steve winter,mike breault"},
	{X: 1, Y: 5, Colour: 0xa, Text: "game created by:"},
	{X: 0x12, Y: 5, Colour: 0xe, Text: "ssi special projects"},
	{X: 0xe, Y: 6, Colour: 0xb, Text: "version 1.3"},
	{X: 2, Y: 7, Colour: 0xe, Text: "programming:"},
	{X: 0x14, Y: 7, Colour: 0xe, Text: "original"},
	{X: 2, Y: 8, Colour: 0xb, Text: "scot bayless"},
	{X: 0x14, Y: 8, Colour: 0xe, Text: "programming:"},
	{X: 2, Y: 9, Colour: 0xb, Text: "brad myers"},
	{X: 0x14, Y: 9, Colour: 0xb, Text: "keith brors"},
	{X: 2, Y: 0xa, Colour: 0xb, Text: "russ brown"},
	{X: 0x14, Y: 0xa, Colour: 0xb, Text: "brad myers"},
	{X: 2, Y: 0xb, Colour: 0xb, Text: "ted greer"},
	{X: 2, Y: 0xd, Colour: 0xe, Text: "graphic arts:"},
	{X: 0x14, Y: 0xd, Colour: 0xe, Text: "encounter coding:"},
	{X: 2, Y: 0xe, Colour: 0xb, Text: "tom wahl"},
	{X: 0x14, Y: 0xe, Colour: 0xb, Text: "paul murray"},
	{X: 2, Y: 0xf, Colour: 0xb, Text: "fred butts"},
	{X: 0x14, Y: 0xf, Colour: 0xb, Text: "russ brown"},
	{X: 2, Y: 0x10, Colour: 0xb, Text: "darla marasco"},
	{X: 0x14, Y: 0x10, Colour: 0xb, Text: "victor penman"},
	{X: 2, Y: 0x11, Colour: 0xb, Text: "susan halbleib"},
	{X: 0x14, Y: 0x11, Colour: 0xb, Text: "dave shelley"},
	{X: 2, Y: 0x13, Colour: 0xe, Text: "project manager"},
	{X: 0x14, Y: 0x13, Colour: 0xe, Text: "developer:"},
	{X: 2, Y: 0x14, Colour: 0xb, Text: "victor penman"},
	{X: 0x14, Y: 0x14, Colour: 0xb, Text: "george mac donald"},
	{X: 2, Y: 0x16, Colour: 0xe, Text: "testing:"},
	{X: 0xb, Y: 0x16, Colour: 0xb, Text: "joel billings,steve salyer"},
	{X: 2, Y: 0x17, Colour: 0xb, Text: "james kucera,robert daly,rick white"},
}

// hostFiles are the port's own files beside the originals (the host -data
// directory, dos's parent) the folder carries, from -> to (folder path).
// fonts/ and data/music/ stay the host's. TestImportContents checks them.
var hostFiles = []struct{ From, To string }{
	{"amiga/SOUND.BIN", "sound/SOUND.BIN"},
	{"assets/dasp", "sound/dasp"},
	{"assets/speech", "speech"},
	{"assets/soundtrack", "soundtrack"},
	{"amiga/TITLE.PNG", "art/TITLE.PNG"},
}

// Import writes POR's folder dst from the DOS originals in dos.
func Import(dos, dst string) error {
	scripts := filepath.Join(dst, "scripts")
	if err := writeScriptData(scripts); err != nil {
		return err
	}
	if err := ImportScripts(dos, scripts); err != nil {
		return err
	}
	amiga := filepath.Join(filepath.Dir(dos), pool.AmigaDir)
	if err := dax.ImportArt(dos, amiga, filepath.Join(dst, pool.ArtDir)); err != nil {
		return err
	}
	if err := pictures.WritePictures("por", dst); err != nil {
		return err
	}
	for _, h := range hostFiles {
		if err := copyEntry(filepath.Join(filepath.Dir(dos), h.From), filepath.Join(dst, h.To)); err != nil {
			return err
		}
	}
	return ImportDataTo(dos, filepath.Join(dst, pool.DataDir))
}

// copyEntry copies file or directory src to dst.
func copyEntry(src, dst string) error {
	fi, err := os.Stat(src)
	if err != nil {
		return err
	}
	if fi.IsDir() {
		return os.CopyFS(dst, os.DirFS(src))
	}
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, b, 0o644)
}
