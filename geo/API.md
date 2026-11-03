# gobox/geo API
Pure parser: block bytes in (caller's own dax), plain values out. No dax, no engine state, no drawing.

- `ParseMap(b) (*Map, error)`: GEO block (at least 0x402 bytes, else `ErrSize`; longer ones are cut). `ParseWalldef` of a block shorter than one 0x30C chunk is an empty Walldef, not an error. `Map.Plane` is the 4 planes; `Map.SolidEdges` set by game.
- `Map.WallAt/DoorAt/MapCell(x,y[,dir])`: dir 0 N,2 E,4 S,6 W. Off map: 0 if SolidEdges, else wrap (>15->0, <0->15). DoorAt: 1 for a side without wall; 2-bit state otherwise.
- `OnMap(x,y)`, `ValidWallSlot(s)` (1..3).
- `ParseWalldef(b) (*Walldef, error)`: split into 0x30C chunks (5 walls x 0x9C ids); `Sets`, `Wall(k,i)`, `Fits(slot)`, `Relocate(k,delta)` (ids >= 0x2D), `SymbolBlock(block,k,n)` (block*10+k if n>1).
- Decided (DESIGN.md): chunks = floor(len/780); a trailing partial chunk is dropped, as the later games did. Bad WALLDEF: ignore (error). Bad GEO: game reports "Unable to load geo in Load3DMap."

## Per-game differences
- por: SolidEdges=true always; keeps only block+slot of WALLDEF (no parse needed; block 10 recorded as 0 quirk stays game-side); GEO size mismatch tolerated if len>=0x402 (game-side).
- cotab: SolidEdges = areaKind 0 or 0xA; also keeps WallSlot{Block,Chunk,Symbols,Loaded}.
- sotsb: SolidEdges = script block 0 or 80; copies chunks into walldef[3*0x30C], relocates by the difference of its own per-slot symbol base table, passed as `Relocate`'s delta.
- Stay game-side: sky colours, overhead flag, the 0x7F second-sky-word rule (not in cotab's current geo.go; unverified), Draw3DView, event/area-word writes, OxXXX addresses.

## Switch steps (per game)
1. go.mod: `require gobox/geo v0.0.0` + `replace gobox/geo => ../gobox-engine/geo`.
2. geoState.geo -> `*geo.Map`; LoadGeo: `m, err := geo.ParseMap(daxBlock)`; on err report as now; set `m.SolidEdges` per query/area change.
3. WallAt/DoorAt/MapCell delegate to Map; delete onMap/wrap/validWallSlot copies.
4. LoadWalldef: `geo.ParseWalldef` (+Fits, Relocate, SymbolBlock); delete walldefSet/SetSize constants.
5. go test in the game; compare a few areas' WallAt/DoorAt before/after.

## blobber
Blobber's `maps` layout can be built from a `*geo.Map` by looping x,y 0..15 over WallAt/DoorAt/MapCell; or accept an interface `{WallAt;DoorAt;MapCell}` which `*geo.Map` satisfies (no blobber import of geo needed).

## Migrate
- No code change: the symbol base table was never in geo, only named in comments and docs; the game keeps it and passes the delta to `Relocate`. The dropped partial WALLDEF chunk is now documented as decided.
