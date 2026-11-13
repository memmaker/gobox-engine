# gobox/ega
Constants `W=320`, `H=200`. `Palette [16][3]uint8` (EGA RGB), `RGBA(i int) color.RGBA`.
Amiga: `Amiga [3][16]uint16` (12-bit), slots `AmigaView/AmigaPortrait/AmigaTerrain`, `AmigaBase=32`,
`AmigaSaturation` (1.4), `AmigaRGBA(i uint8)` (pixel value >= AmigaBase, saturated).
Verified: Palette and Amiga tables are identical in por pal, cotab dax/palette.go (EGA only), sotsb ega.
Not here: sotsb's Screen/keyboard queue/Overlay (stay in sotsb ega); por's Regs.

Each game: go.mod: `require gobox v0.0.0`, `replace gobox => ../gobox-engine` and `replace gobox/rules => ../gobox-engine/rules` (one module: every package below imports as `gobox/<dir>`).
- por: `pal.W/H/Palette/EGA(i)/Amiga*` -> `ega.W/H/Palette/RGBA(i)/Amiga*`. `pal.Regs.RGBA` stays in
  por/pal, but its Amiga branch becomes `return ega.AmigaRGBA(i)` and `EGA(...)` -> `ega.RGBA(...)`.
- cotab: `dax.Palette` keeps working (alias var in gobox/asset/orig); or use `ega.Palette`. Drop `dax.AmigaBase` use -> `ega.AmigaBase`.
- sotsb: delete Palette/Amiga*/W/H/AmigaBase from sotsb/ega, keep Screen etc.; replace uses with
  `gobox/ega` (import under another name, e.g. `gega`, as sotsb's own package is also `ega`).
  Screen.RGBA's Amiga branch -> `gega.AmigaRGBA(i)`.

## Migrate
Colour numbers `ega.Black Blue Green Cyan Red Magenta Brown LightGrey DarkGrey LightBlue LightGreen LightCyan LightRed LightMagenta Yellow White` (0..15).
- cotab-go, sotsb-go engine/colours.go: delete the `colourBlack ... colourWhite` consts and use `ega.Black`, `ega.Blue`, `ega.Red`, `ega.Magenta`, `ega.LightGreen`, `ega.LightCyan`, `ega.LightRed`, `ega.LightMagenta`, `ega.Yellow`, `ega.White` (cotab uses the subset without Red/Magenta; both games' green and cyan are LightGreen and LightCyan, 10 and 11). A game whose own names clash wraps the API in a small adapter.. Keep ticksPerSecond, titleSound, states, scan codes.
- sotsb-go: import gobox/ega under another name (`gega`), as its own package is `ega`.

A game whose own names clash with this API wraps them in a small adapter.
