# Buck Rogers series art recipe (BR1 and BR2)

New pixel-scale art for the buck-rogers series comes from PixelLab's pixel
art model (never a downscaled hi-res picture), made the same way every time
so the series stays one style.

## Style

- Endpoint: `POST /v2/generate-with-style-v2`. One generation returns a grid
  of variants at the style images' size: 64 at 16-32 px, 16 at 51-64 px.
- Style images: two of the original game's own 24x24 combat icons from
  `sprites/series/buck-rogers/combat`, cropped (not scaled) to the target
  size. The bar icons use `combat_icon_000.png`, `combat_icon_001.png`, and
  the item icons use the same two at 64x64 (the user granted 64x64 VGA-colour item icons, 2026-11-05).
- style_description: `1990 SSI Buck Rogers XXVc, EGA 16-colour pixel art,
  hard black outline, flat shading, retro pulp sci-fi`
- target_palette: the 16 EGA colours, `max_colors` 16:
  000000 0000AA 00AA00 00AAAA AA0000 AA00AA AA5500 AAAAAA
  555555 5555FF 55FF55 55FFFF FF5555 FF55FF FFFF55 FFFFFF
- seed: 2591 (a retry adds 1 and notes it here).
- no_background: true.

## Prompt template

`{subject}, one {kind} icon, centred, readable at small size, Buck Rogers
XXVc rocket-age sci-fi`

- kind `combat button` for the bar icons, at 64x64 (`combat/icons/buck-rogers/<word>.png`)
- kind `inventory item` for the item icons, at 64x64 in a 64-colour VGA palette from BR art (`itemicons/atlas.png`, 24 cells a row, 3 variants per item, mapped by `items.rec`; edit with buckrogers-go/remap_items.sh)

## Assets

- Bar icons (64x64, the word is the button's label lowercased): move, target, view,
  intimidate, quick, aid, sprint, dodge, jetpack, wait, guard, end.
- Item icons (64x64): the BR1/BR2 item names, mapped by name in items.rec.

## Budget log

BR1: 50 generations in all. The tally is in the notes folder's br1-gaps.md.
