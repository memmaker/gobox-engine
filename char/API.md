# char

A character record described as data. Games keep their own record structs;
`Rec` reads and writes the same bytes through a `Layout`.

- `Layout{Size, Fields map[string]Field}`; `Field{Off, Size(1/2/4, LE), N, Stride, Signed}`.
- `LoadLayout(path)`, `ParseLayout(json)`. Layouts: `testdata/cotab.json` (422 bytes),
  `testdata/sotsb.json` (439). Examples of the field names: score/scoreUse (6, stride 2),
  str00/str00Use, saves, saveBonus, thac0(Base), ac, hp, maxHP, hpRoll, levels, oldLevels,
  multiclassLevel, race, class, alignment, status, okay, exp, coins, weight, itemList,
  slots (the 52-byte readied item pointer area). Absent fields read 0 and ignore writes.
- `Rec{B, L}`: `Get(name, i)`, `Set(name, i, v)`, `Bytes(name)`, `Has`, `Len`, `New(layout)`.
- `Stats` interface: what rules needs (levels, hp, thac0, ac, saves, status and setters). `Rec` implements it.

Tests check both layouts against copies of the games' `charRecord` structs with encoding/binary.

## Migrate

### cotab-go
Keep `charRecord` for now. Load `char.Rec{B: c.record(), L: layout}` where a rules call needs a `Stats`,
or make `Character` implement `char.Stats` over its fields (a dozen one-line methods). The layout file
ships with the game (copy `testdata/cotab.json` to the game's data).

### sotsb-go
Same, with `testdata/sotsb.json`. `Levels`/`OldLevels` have 7 entries; `Classes()` reports it.

### por-go
Character is a Go struct: implement `char.Stats` on it (`Level(k)` = `int(c.Levels[k])` ...). No layout needed.
