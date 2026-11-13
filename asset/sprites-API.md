# sprites API

    import "gobox/sprites"
    sprites.Image(pool, name string) image.Image // nil if absent
    sprites.Names(pool string) []string          // sorted

Pools: `portraits`, `combat`, `items` (names without .png). `portraits` and `combat` are empty for now.

## Migrate

### por-go
- `present/remaster/items/atlas.png` cell N (via `items.rec` `icon:`) -> `sprites.Image("items", "<item_name_snake_case>")`, e.g. `battle_axe`. Look up the name in INDEX.md by atlas cell. Duplicate-icon items (see INDEX.md) share the kept name.
- `present/remaster/doll/doll.png`, `bars.png`: UI, not pooled.

### cotab-go
Nothing to migrate (no portrait, combat or item PNGs).

### sotsb-go
Nothing to migrate (no portrait, combat or item PNGs).
