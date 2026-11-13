# gobox/asset/orig
Superset of por's dax. Loader: `NewLoader(dir)` (no Amiga auto-detect),
`UseAmiga(dir)` (por), `Path`, `Exists`, `File(name)`, `Block(name,id)`, `BlockShared(name,id)`, `DOSBlock(name,id)`.
File: `Open`, `Block(id)`, `BlockShared(id)`, `Entries`. Also `Decode`, `WriteDOS`, `OpenAmiga`,
`AmigaImage`, `AmigaSprite`, `ParseImage(b,masked,maskColour)`, `ParseSprite`, `SpriteFrame`, `NewImage`,
`Image` (Frame, Width, Height, Count, Mask, File, Block, Base, Alt, Clone, FlipLeftToRight, Fit), `Clear=16`,
`Palette` (copy of ega.Palette, kept for cotab).
Amiga "AM" tag (header bytes 9.., sprite frame bytes 12..): 'A','M',slot,8 = wide picture, one byte per pixel, 32 colours over slots slot, slot+1 (amigaimport writes it for Pools of Darkness).
Block vs BlockShared: `Block` returns a private COPY (cotab/sotsb semantics, safe to modify);
`BlockShared` returns the cached slice, read-only (por's old `Block`, with memo for Amiga overlay). Same on File and Loader.

All games: go.mod: `require gobox v0.0.0`, `replace gobox => ../gobox-engine` and `replace gobox/rules => ../gobox-engine/rules` (one module: every package below imports as `gobox/<dir>`).
import `por/dax` | `cotab/dax` | `sotsb/dax` -> `gobox/asset/orig`; delete the game's dax/ (and cotab's palette.go).
- por: rename every `.Block(` on a dax Loader/File to `.BlockShared(` (keeps exact old no-copy behaviour);
  UseAmiga/DOSBlock/Image.Alt/Fit unchanged. por's dax tests moved here.
- cotab: set `l.OneDisk = true` after every NewLoader; otherwise no call changes (Block copies, dax.Palette alias). Image has extra fields (Alt, File, Block, Base) unused.
- sotsb: set `l.OneDisk = true`; otherwise no call changes; its dax/amiga.go ega import is now gobox/ega internally.
- `Loader.OneDisk`: a block missing from its file stays missing (default: looked up on the other disks, GEO3 → GEO1..8, por's behaviour). cotab and sotsb set it.

## Migrate
- `NewLoader` no longer looks for a DISK2 subdir. A game that kept Amiga archives there calls `l.UseAmiga(filepath.Join(dir, "DISK2"))` (cotab, sotsb) right after `NewLoader`.


- cotab-go, sotsb-go engine/records.go: `wordAt`/`putWord` are not moved; switch to `binary.LittleEndian.Uint16(b[o:])` / `binary.LittleEndian.PutUint16(b[o:], v)` and delete them.
- Left in the games, not moved: engine/porstate.go `eclRun` (ovr04 interpreter globals still read by por-derived combat/camp/treasure code, in cotab-go and sotsb-go; por-go has no porstate.go).
