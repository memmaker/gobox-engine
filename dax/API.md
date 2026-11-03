# gobox/dax (needs gobox/ega)
Superset of por's dax. Loader: `NewLoader(dir)` (no Amiga auto-detect),
`UseAmiga(dir)` (por), `Path`, `Exists`, `File(name)`, `Block(name,id)`, `BlockShared(name,id)`, `DOSBlock(name,id)`.
File: `Open`, `Block(id)`, `BlockShared(id)`, `Entries`. Also `Decode`, `WriteDOS`, `OpenAmiga`,
`AmigaImage`, `AmigaSprite`, `ParseImage(b,masked,maskColour)`, `ParseSprite`, `SpriteFrame`, `NewImage`,
`Image` (Frame, Width, Height, Count, Mask, File, Block, Base, Alt, Clone, FlipLeftToRight, Fit), `Clear=16`,
`Palette` (copy of ega.Palette, kept for cotab).
Block vs BlockShared: `Block` returns a private COPY (cotab/sotsb semantics, safe to modify);
`BlockShared` returns the cached slice, read-only (por's old `Block`, with memo for Amiga overlay). Same on File and Loader.

All games: `require gobox/dax v0.0.0`, `replace gobox/dax => ../gobox-engine/dax` (+ `gobox/ega` replace, indirect);
import `por/dax` | `cotab/dax` | `sotsb/dax` -> `gobox/dax`; delete the game's dax/ (and cotab's palette.go).
- por: rename every `.Block(` on a dax Loader/File to `.BlockShared(` (keeps exact old no-copy behaviour);
  UseAmiga/DOSBlock/Image.Alt/Fit unchanged. por's dax tests moved here.
- cotab: set `l.OneDisk = true` after every NewLoader; otherwise no call changes (Block copies, dax.Palette alias). Image has extra fields (Alt, File, Block, Base) unused.
- sotsb: set `l.OneDisk = true`; otherwise no call changes; its dax/amiga.go ega import is now gobox/ega internally.
- `Loader.OneDisk`: a block missing from its file stays missing (default: looked up on the other disks, GEO3 → GEO1..8, por's behaviour). cotab and sotsb set it.

## Migrate
- `NewLoader` no longer looks for a DISK2 subdir. A game that kept Amiga archives there calls `l.UseAmiga(filepath.Join(dir, "DISK2"))` (cotab, sotsb) right after `NewLoader`.

## cmd (dev tools)
- `dax/cmd/blockdump FILE.DAX ID` (was cmd/fontdump, identical in cotab and sotsb): block to stdout.
- `dax/cmd/ecldump [-onedisk] DATA [AREA [BLOCK]]` (was cmd/ecldump): counts ECL commands. cotab's copy hard-coded OneDisk=true: pass `-onedisk`.

## Migrate (tools)
- cotab-go, sotsb-go: delete cmd/fontdump and cmd/ecldump; run `go run gobox/dax/cmd/blockdump` / `ecldump` from the dax module (cotab: add `-onedisk`).
- The tools also need nothing in the game's go.mod.
