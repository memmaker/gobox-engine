# walls

Shared wall/floor/ceiling texture pool (`gobox/asset`), keyed by asset series: `series/<series>/<name>.png`, names without `.png`; see INDEX.md. `Image(series, name)`, `Names(series)`, `Textures(series, names...)`, `Series()`.

- `walls.Image(name) image.Image` - nil if absent
- `walls.Names() []string`
- `walls.Textures(names ...string) map[string]image.Image` - fill `blobber.View.Textures`

Each game keeps its own maps and character-to-texture table; only the textures are shared.

## Migrate

### por
Replace `floor_phlan_water` with `floor_water`, `ceiling_phlan_clouds` with `ceiling_clouds`. All other names are unchanged. Drop blobber/assets/walls PNGs in favour of the pool (keep walls.txt: your own character table).

### cotab
Replace `ceiling_phlan_clouds` with `ceiling_clouds` (in view3d.go's list). All other names (`wall_*`, `floor_cobble`, `ceiling_sky_phlan`) are unchanged; cmd/cotab/walls/*.png can be replaced by `walls.Image` (keep walls/index.txt, your own map table). The `wallN` fallbacks from WALLDEF stay generated.

### sotsb
Nothing to change: it ships no PNGs; its `wallN` textures stay decoded at run time. It may select pool textures by name.
