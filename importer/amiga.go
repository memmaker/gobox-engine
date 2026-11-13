package importer

import (
	"gobox/amigaimport"
	"gobox/asset"
)

// The Amiga releases' disks, as `gobox import -amiga <game> <adfdir>`
// converts them (importer.Game.Amiga; were the per-game launchers'
// -import-disk2 / -import-amiga).

// porAmiga: the Pool of Radiance data disk. Each archive with the palette
// (0x54080 for PIC/FINAL, which the Amiga sets before them; 0x54020 for the
// ground tiles and the combat icons, which the combat screen (0x39F0E) and
// the icon editor (0x2EB02) set; 0x54040 for the portraits) it is drawn in,
// whether its frames pack two tiles side by side, and whether it is an
// animation block (PIC/FINAL frames XORed against the first).
// ponytail: 8X8D (the 3D-view symbols) stays DOS; the Amiga set is laid out
// for its own wall renderer and does not match WALLDEF's symbol numbers.
var porAmiga = amigaimport.Spec{
	Gfx: map[string]amigaimport.Gfx{
		"icon": {Slot: asset.AmigaTerrain}, "comspr": {Slot: asset.AmigaTerrain}, "cpic": {Slot: asset.AmigaTerrain},
		"bacpac": {Slot: asset.AmigaTerrain, Split: true}, "sqrpaci": {Slot: asset.AmigaTerrain, Split: true}, "wildcom": {Slot: asset.AmigaTerrain, Split: true},
		"dungcom": {Slot: asset.AmigaTerrain, Split: true}, "randcom": {Slot: asset.AmigaTerrain, Split: true},
		"head": {Slot: asset.AmigaPortrait}, "body": {Slot: asset.AmigaPortrait}, "chead": {Slot: asset.AmigaTerrain}, "cbody": {Slot: asset.AmigaTerrain},
		"pic": {Slot: asset.AmigaView, Sprite: true, XOR: true}, "final": {Slot: asset.AmigaView, Sprite: true, XOR: true}, "sprit": {Slot: asset.AmigaPortrait, Sprite: true},
	},
	Disks: []string{"poolgame", "POOLDATA"}, Program: "program", Intro: "intro", Font: "fonts/poolrad/8",
	Skip:       map[string]string{"8x8d.dax": "the Amiga set is laid out for its own wall renderer and does not match WALLDEF's symbol numbers: the 3D view stays DOS (porAmiga)"},
	ProgramGfx: map[string]amigaimport.ProgramGfx{porFrameFile: {Slot: asset.AmigaView, Blocks: porFrameBitMaps}},
}

// porFrameBitMaps are the window frame's pieces: no disk file holds them,
// the program does (struct BitMaps in code hunk 27, 0x5c000; their planes
// in data hunk 28, 0x65000). The frame routine 0x576d4 blits each with
// BltBitMap (0x5a9a0, minterm 0xC0) from its top-left: the bars as wide
// as the window, the sides and the corner 8 pixels wide.
var porFrameBitMaps = map[int]amigaimport.BitMap{
	porFrameTop:    {Addr: 0x637ec},                   // 320x8, blitted at 0x57750
	porFrameBottom: {Addr: 0x63810},                   // 320x8, 0x577a6
	porFrameRight:  {Addr: 0x63834, Width: frameCell}, // 16x190, 0x57856
	porFrameLeft:   {Addr: 0x63858, Width: frameCell}, // 16x190, 0x57800
	porFrameCorner: {Addr: 0x6387c, Width: frameCell}, // 16x8, all four corners, 0x5789c..0x57966
}

// laterGfx is how the later Pool games' Amiga archives decode: pal is the
// offset of the fixed palette in the program (Spec.Exe) for the pictures
// that carry none of their own; mask the colour written for the transparent
// pixels of the 3D view's wall pieces and sky (Curse: 0xD; else 0).
// Base/Frames say how the unindexed libraries map to DOS block ids.
func laterGfx(pal int, mask byte) map[string]amigaimport.Gfx {
	return map[string]amigaimport.Gfx{
		"8x8d": {Mask: mask, Pal: pal}, "bigpic": {Pal: pal}, "body": {Pal: pal}, "head": {Pal: pal}, "cbody": {Pal: pal}, "chead": {Pal: pal},
		"cpic": {Pal: pal}, "comspr": {Pal: pal}, "tiles": {Pal: pal}, "back": {Mask: mask, Pal: pal}, "skygrnd": {Mask: mask, Pal: pal}, "gen": {Pal: pal},
		"dungcom": {Frames: true, Base: 1, Pal: pal}, "wildcom": {Frames: true, Base: 1, Pal: pal}, "randcom": {Frames: true, Base: 1, Pal: pal},
		"sky":   {Base: 0xfa, Mask: mask, Pal: pal},
		"title": {Pal: pal}, "intro": {Pal: pal},
		"pic": {Sprite: true, XOR: true, Pal: pal}, "final": {Sprite: true, XOR: true, Pal: pal}, "frame": {Sprite: true, Pal: pal},
		"sprit": {Sprite: true, Pal: pal},
	}
}

// withFrame is gfx with the program's window-frame library ALWAYS.TLB
// converted too (its pieces drawn as AmigaFrame screens), its masked
// pieces' clear pixels written as colour mask.
func withFrame(gfx map[string]amigaimport.Gfx, pal int, mask byte) map[string]amigaimport.Gfx {
	gfx["always"] = amigaimport.Gfx{Pal: pal, Mask: mask}
	return gfx
}

// alwaysMask is the colour the Silver Blades' masked frame piece (ALWAYS 41,
// the rule) leaves clear: no frame piece (ALWAYS 37..50, 66, 68 use colours
// 0, 1, 6..9, 0xD, 0xF) draws it. Curse's frame pieces carry no mask.
const alwaysMask = 0xE

// cotabAmiga: the two Curse disks; SOUNDS.GLB the sampled effects (CotabAmigaMusic).
var cotabAmiga = amigaimport.Spec{Gfx: withFrame(laterGfx(0x47410, 0xD), 0x47410, alwaysMask), Disks: []string{"CurseA", "CurseB"}, Exe: "Curse", Font: "DISKA/CURSE.FON", Files: []string{"SOUNDS.GLB"}}

// podAmiga: the three Pools of Darkness disks; the music (pod.mx) and the
// sampled effects (SOUNDS.GLB) copied as they are.
var podAmiga = amigaimport.Spec{Gfx: laterGfx(0x51354, 0), Disks: []string{"POD_Disk1_v.1.00", "POD_Disk2_v.1.00", "POD_Disk3_v.1.00"}, Exe: "Pools of Darkness", Font: "DISK1/pod.fon",
	Files: []string{"pod.mx", "SOUNDS.GLB"}, Alias: map[string]string{"8x8da": "8x8d"}}

// sotsbAmiga: the two Silver Blades disks with their music.
var sotsbAmiga = amigaimport.Spec{Gfx: withFrame(laterGfx(0x4fdea, 0), 0x4fdea, alwaysMask), Exe: "Secret", Font: "DISK1/SECRET.FON", Disks: []string{"Secret 1", "Secret 2"}, Files: []string{SotsbAmigaMusic.Songs, SotsbAmigaMusic.Effects}}
