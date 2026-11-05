#!/usr/bin/env python3
"""Buck Rogers 1 combat floor: the PixelLab deck panel (br_combat_floor_panel.png, 24x24) under the original tiles.

Reads the original DOS tiles from git (commit ORIG, so reruns are idempotent) and classifies each:
- floor (only the old floor: hitek's black 8px diagonal hatch on grey 82, lotek's flat grey 82)
  -> the panel;
- separable object (lotek walls/props on flat grey, hitek props on hatch): its contiguous blobs
  (8-connected, >= MIN_BLOB px, holes filled) are pasted whole over the panel, stray pixels dropped;
- hitek 45 (the port's plain floor, hatch plus a small grate) -> the panel;
- hitek floor grates 46-48 -> the panel plus every pixel that differs from the bare hatch;
- hitek walls -> each pixel equal to the bare hatch at the same (x, y) (tile 45's pattern and phase,
  minus its small grate patch, whose black would otherwise eat wall black) becomes the panel pixel;
  hatch-line black touching off-line black (inside wall black) and all other pixels stay original.

    python3 tools/remaster/make_br_combat_floor.py
"""
import io, os, subprocess
from PIL import Image

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
DIR = 'sprites/series/buck-rogers/combat'
ORIG = '7a6aa23'
PANEL = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'br_combat_floor_panel.png')
SETS = {'hitek': 53, 'lotek': 52}
HITEK_PROPS = range(35, 45)
HITEK_FLOOR, HITEK_GRATES = 45, range(46, 49)  # consoles, chairs, machines standing on the hatch
GREY = (82, 82, 82)
MIN_BLOB = 12
N8 = [(dx, dy) for dx in (-1, 0, 1) for dy in (-1, 0, 1) if dx or dy]


def original(name):
    return Image.open(io.BytesIO(subprocess.check_output(
        ['git', 'show', f'{ORIG}:{DIR}/{name}.png'], cwd=ROOT))).convert('RGB')


def old_floor(s, x, y):
    if s == 'lotek':
        return GREY
    return (0, 0, 0) if y % 8 == 0 or x % 8 == y % 8 else GREY


def components(pts, nb):
    pts, out = set(pts), []
    while pts:
        stack, comp = [pts.pop()], set()
        while stack:
            p = stack.pop(); comp.add(p)
            for dx, dy in nb:
                q = (p[0] + dx, p[1] + dy)
                if q in pts:
                    pts.remove(q); stack.append(q)
        out.append(comp)
    return out


def object_mask(s, im):
    """Pixels of the tile's contiguous object blobs, holes filled; stray specks dropped."""
    on = [(x, y) for y in range(24) for x in range(24) if im.getpixel((x, y)) != old_floor(s, x, y)]
    keep = set().union(*[c for c in components(on, N8) if len(c) >= MIN_BLOB])
    off = [(x, y) for y in range(24) for x in range(24) if (x, y) not in keep]
    for c in components(off, [(1, 0), (-1, 0), (0, 1), (0, -1)]):  # enclosed pockets belong to the object
        if not any(x in (0, 23) or y in (0, 23) for x, y in c):
            keep |= c
    return keep


def in_black(s, im, p):
    """Whether a 4-neighbour off the hatch lines is black (only wall art is black there)."""
    for dx, dy in ((1, 0), (-1, 0), (0, 1), (0, -1)):
        x, y = p[0] + dx, p[1] + dy
        if 0 <= x < 24 and 0 <= y < 24 and old_floor(s, x, y) != (0, 0, 0) and im.getpixel((x, y)) == (0, 0, 0):
            return True
    return False


def tile(s, t, im, panel):
    if (s, t) == ('hitek', HITEK_FLOOR):
        return panel.copy(), 'floor'
    if all(im.getpixel((x, y)) == old_floor(s, x, y) for y in range(24) for x in range(24)):
        return panel.copy(), 'floor'
    if s == 'hitek' and t in HITEK_GRATES:
        out = panel.copy()
        for p in [(x, y) for y in range(24) for x in range(24) if im.getpixel((x, y)) != old_floor(s, x, y)]:
            out.putpixel(p, im.getpixel(p))
        return out, 'grate'
    if s == 'hitek' and t not in HITEK_PROPS:
        out = im.copy()
        for p in [(x, y) for y in range(24) for x in range(24) if im.getpixel((x, y)) == old_floor(s, x, y)]:
            if im.getpixel(p) == (0, 0, 0) and in_black(s, im, p):
                continue  # a hatch-line pixel inside a black wall area is wall black
            out.putpixel(p, panel.getpixel(p))
        return out, 'wall'
    out = panel.copy()
    for p in object_mask(s, im):
        out.putpixel(p, im.getpixel(p))
    return out, 'object'


def main():
    panel = Image.open(PANEL).convert('RGB')
    for s, n in SETS.items():
        for t in range(n):
            name = f'{s}_tile_{t:03d}'
            out, kind = tile(s, t, original(name), panel)
            out.save(f'{ROOT}/{DIR}/{name}.png')
            print(name, kind)


if __name__ == '__main__':
    main()
