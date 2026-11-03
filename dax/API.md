# gobox/dax (needs gobox/ega)
Superset of por's dax. Loader: `NewLoader(dir)` (auto-uses a DISK2 subdir as the Amiga set: cotab/sotsb),
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
- cotab: no call changes (Block copies, DISK2 auto-detect, dax.Palette alias). Image has extra fields (Alt, File, Block, Base) unused.
- sotsb: no call changes; its dax/amiga.go ega import is now gobox/ega internally.
- `Loader.OneDisk`: a block missing from its file stays missing (default: looked up on the other disks, GEO3 → GEO1..8). sotsb sets it.
