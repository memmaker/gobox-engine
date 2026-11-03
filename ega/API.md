# gobox/ega
Constants `W=320`, `H=200`. `Palette [16][3]uint8` (EGA RGB), `RGBA(i int) color.RGBA`.
Amiga: `Amiga [3][16]uint16` (12-bit), slots `AmigaView/AmigaPortrait/AmigaTerrain`, `AmigaBase=32`,
`AmigaSaturation` (1.4), `AmigaRGBA(i uint8)` (pixel value >= AmigaBase, saturated).
Verified: Palette and Amiga tables are identical in por pal, cotab dax/palette.go (EGA only), sotsb ega.
Not here: sotsb's Screen/keyboard queue/Overlay (stay in sotsb ega); por's Regs.

Each game: add `require gobox/ega v0.0.0` + `replace gobox/ega => ../gobox-engine/ega`.
- por: `pal.W/H/Palette/EGA(i)/Amiga*` -> `ega.W/H/Palette/RGBA(i)/Amiga*`. `pal.Regs.RGBA` stays in
  por/pal, but its Amiga branch becomes `return ega.AmigaRGBA(i)` and `EGA(...)` -> `ega.RGBA(...)`.
- cotab: `dax.Palette` keeps working (alias var in gobox/dax); or use `ega.Palette`. Drop `dax.AmigaBase` use -> `ega.AmigaBase`.
- sotsb: delete Palette/Amiga*/W/H/AmigaBase from sotsb/ega, keep Screen etc.; replace uses with
  `gobox/ega` (import under another name, e.g. `gega`, as sotsb's own package is also `ega`).
  Screen.RGBA's Amiga branch -> `gega.AmigaRGBA(i)`.
