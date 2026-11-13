package importer

import (
	"errors"
	"gobox/adventure"
	"gobox/asset"
	"image/color"
	"os"
	"path/filepath"

	"gobox/asset/orig"
)

// init registers POR's import with gobox/importer: the
// decompiled scripts with POR's fixes, the art sheets and the data (maps,
// monsters, items, POOL.CFG, START.EXE's sound tables) into
// adventures/por/{scripts,art,data}. The Amiga art is taken from a folder
// amiga beside the originals when there is one (POR's data/dos, data/amiga).
func init() {
	Register(Game{
		Cover: true,
		ID:    "por", Title: "Pool of Radiance", Release: 1988, Ruleset: "rules1e", Record: adventure.RecordLayout, Combat4e: true, Version: 4, // 4: art/LOGO.png, the cut menu lettering (Data.Logo gone); 3: Combat4e in the manifest (playpool offers rules4e on it); 2: Window, UserSaves, Foreign in the manifest (playpool reads them); 1: 1: no Sequence.Remaster (the remaster credit line is gobox's), SequenceStep.Credits

		Window:  Size{W: asset.ScreenW, H: asset.ScreenH},
		Sources: []string{"*.DAX", "ITEMS", "POOL.CFG", "START.EXE"},
		Import:  PorImport, Amiga: &porAmiga,
		Start: Fixed(porStart()),
		End:   Fixed(porEnd()),
	})
}

// whole is the 40x25 page.
var porWhole = adventure.Cells{X1: 0, Y1: 0, X2: 39, Y2: 24}

// start is POR's title (ovr02:0626 title_screen, as pool.Engine.State's
// SceneTitle builds it): TITLE 1 with the Amiga intro's song (titleSong 1),
// TITLE 2 with its lightning, the credits page (ovr02:0206: credits in
// its two boxes); a short intro (POOL.CFG's, Intro.Full off) is TITLE 1
// only. The Amiga painting art/TITLE.PNG lies under the main menu.
func porStart() *adventure.Sequence {
	q := &adventure.Sequence{
		Steps: []adventure.SequenceStep{
			{Kind: adventure.StepPicture, Layers: []adventure.Layer{{Art: adventure.Art{File: "TITLE", Block: 1}, At: porWhole}}, Song: "1", Whole: true},
			{Kind: adventure.StepPicture, Layers: []adventure.Layer{{Art: adventure.Art{File: "TITLE", Block: 2}, At: porWhole}}, Whole: true, Lightning: adventure.FlashEGA},
			{Kind: adventure.StepCredits, Page: 0},
		},
		Boxes:    []adventure.Cells{{X1: 1, Y1: 1, X2: 38, Y2: 3}, {X1: 1, Y1: 5, X2: 38, Y2: 23}},
		Backdrop: []adventure.Layer{{Art: adventure.Art{File: "TITLE.PNG", Kind: adventure.ArtPNG}, At: porWhole}},
		Short:    []int{0},
		Full:     true, // POOL.CFG's default
		Credits:  porCredits,
	}
	return q
}

// credits is ovr02:0206's page: each Text at cell X, Y in EGA colour Colour
// (DisplayString's way, upcased).
var porCredits = []adventure.Credit{
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
func PorImport(dos, dst string) error {
	scripts := filepath.Join(dst, "scripts")
	if err := writeScriptData(scripts); err != nil {
		return err
	}
	if err := ImportScripts(dos, scripts); err != nil {
		return err
	}
	amiga := filepath.Join(filepath.Dir(dos), adventure.AmigaDir)
	if err := orig.ImportArt(dos, amiga, filepath.Join(dst, adventure.ArtDir)); err != nil {
		return err
	}
	if err := writeLogo(filepath.Join(dst, adventure.ArtDir)); err != nil {
		return err
	}
	if err := asset.WritePictures("por", dst); err != nil {
		return err
	}
	for _, h := range hostFiles {
		if err := copyEntry(filepath.Join(filepath.Dir(dos), h.From), filepath.Join(dst, h.To)); err != nil {
			return err
		}
	}
	return PorImportDataTo(dos, filepath.Join(dst, adventure.DataDir))
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

// porLogo cuts TITLE picture 2's "Pool of Radiance" (yellow with a light blue
// shadow, above row 108) and its subtitle (light magenta, rows 108 to 160) in
// the Amiga intro's colours: gold (0xF90) with a dark red (0x600) outline, the
// intro's credits palette entries 2 and 1.
var porLogo = asset.LogoCut{Sub: 108, Bottom: 160, Keep: []uint8{asset.EGAYellow, asset.EGALightBlue}, SubInk: asset.EGALightMagenta,
	Gold: color.RGBA{0xff, 0x99, 0, 0xff}, Outline: color.RGBA{0x66, 0, 0, 0xff}}

// writeLogo cuts the menu lettering from the imported art in artDir
// (TITLE 2, DOS) and writes it as adventure.LogoFile.
func writeLogo(artDir string) error {
	b := asset.OpenArt(artDir).Block(asset.DOS, "TITLE", 2)
	if b == nil {
		return errors.New("no TITLE 2 in the imported art for the menu logo")
	}
	im := b.Image(asset.NoMask)
	if im == nil {
		return errors.New("TITLE 2 is no picture")
	}
	logo, err := asset.CutLogo(im.Frame(0), im.Width*8, porLogo)
	if err != nil {
		return err
	}
	return asset.WritePNG(filepath.Join(artDir, adventure.LogoFile), logo)
}
