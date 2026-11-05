#!/usr/bin/env python3
"""Buck Rogers 1 combat cut-outs: the EXTRUDE props (organic or flat, no 3D shape to infer) lose their baked-in
ground so gobox extrudes them upright. Load24x24Set (O157.PAS:53) copies frames raw, so every tile is opaque.

Reads the original tiles from git (commit ORIG, so reruns are idempotent). Per tile: flood-fill from the border
through the set's main ground colour; a ground speckle colour joins the fill when 3 of its 4 neighbours are already
ground (or off the tile), so speckles in open ground go and ground colours enclosed by the object stay. Black (the
outline) is never ground. Writes <set>_tile_NNN.png in place (RGBA, ground alpha 0).

    python3 tools/remaster/make_br_cutout.py [--out DIR]   # DIR: write there instead of in place
"""
import io, os, sys, subprocess
from collections import Counter
from PIL import Image

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
DIR = 'sprites/series/buck-rogers/combat'
ORIG = '7a6aa23'
GROUND = {'marscom': (20, 21), 'venuscom': (6, 7), 'lotek': (0,)}  # bare-ground frames (marscom 004 is rock art)
EXTRUDE = {'lotek': [38], 'marscom': [13, 14, 15, 16, 17, 18, 19, 23, 24, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40], 'venuscom': [0, 1, 2, 3, 4, 8, 9, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 36, 37, 38, 39, 40, 41, 42]}
BLACK = (0, 0, 0, 255)


def original(name):
    return Image.open(io.BytesIO(subprocess.check_output(
        ['git', 'show', f'{ORIG}:{DIR}/{name}.png'], cwd=ROOT))).convert('RGBA')


def cutout(s, im):
    cnt = Counter(c for g in GROUND[s] for c in original(f'{s}_tile_{g:03d}').getdata())
    base = cnt.most_common(1)[0][0]
    speck = set(cnt) - {base, BLACK}
    px = im.load()
    inside = lambda p: 0 <= p[0] < 24 and 0 <= p[1] < 24
    nb = lambda p: [(p[0] + 1, p[1]), (p[0] - 1, p[1]), (p[0], p[1] + 1), (p[0], p[1] - 1)]
    ground = set()

    def flood(starts):
        stack = list(starts)
        while stack:
            p = stack.pop()
            if p in ground or not inside(p) or px[p] != base:
                continue
            ground.add(p)
            stack += nb(p)

    flood([(x, y) for x in range(24) for y in (0, 23)] + [(x, y) for y in range(24) for x in (0, 23)])
    while True:
        new = [p for p in ((x, y) for x in range(24) for y in range(24)) if p not in ground and px[p] in speck
               and sum(1 for q in nb(p) if not inside(q) or q in ground) >= 3]
        if not new:
            break
        ground.update(new)
        flood([q for p in new for q in nb(p)])
    out = im.copy()
    for p in ground:
        out.putpixel(p, (0, 0, 0, 0))
    return out


def main():
    out = sys.argv[sys.argv.index('--out') + 1] if '--out' in sys.argv else os.path.join(ROOT, DIR)
    os.makedirs(out, exist_ok=True)
    for s, frames in EXTRUDE.items():
        for t in frames:
            name = f'{s}_tile_{t:03d}'
            cutout(s, original(name)).save(os.path.join(out, name + '.png'))
            print(name)


if __name__ == '__main__':
    main()
