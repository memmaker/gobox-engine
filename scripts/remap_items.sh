#!/bin/bash
# remap_items.sh <series>: edit a series' item icons in remapper
# (asset/sprites/series/<series>/itemicons: atlas.png, items.rec).
# The cell size is the atlas width over its 24 cells a row (ui atlasCols).
set -euo pipefail
cols=24
series=${1:?usage: remap_items.sh <series> ($(ls "$(dirname "$0")/../asset/sprites/series" | tr '\n' ' '))}
dir="$(cd "$(dirname "$0")/.." && pwd)/asset/sprites/series/$series/itemicons"
[ -f "$dir/atlas.png" ] && [ -f "$dir/items.rec" ] || { echo "no item icons for series $series in $dir" >&2; exit 1; }
width=$(sips -g pixelWidth "$dir/atlas.png" | awk '/pixelWidth/ {print $2}')
cell=$((width / cols))
cd "$dir"
exec remapper "$cell" "$cell" atlas.png items.rec
