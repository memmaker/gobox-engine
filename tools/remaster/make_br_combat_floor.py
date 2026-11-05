#!/usr/bin/env python3
"""Buck Rogers 1 combat floor: a dark steel panel per cell under the original walls.

Reads the original DOS tiles from git (commit ORIG, so reruns are idempotent), cuts each
tile's wall/object pixels away from its floor (hitek: the black 8px diagonal hatch,
lotek: flat grey 82), and writes <set>_tile_NNN.png = new floor + those pixels.
With --walls DIR, DIR/<set>_tile_NNN.png (RGBA, remastered wall layer) replaces the
original wall pixels where present. --mask DIR writes each wall layer (RGBA) there.

    python3 tools/remaster/make_br_combat_floor.py [--walls DIR] [--mask DIR]
"""
import io, os, random, subprocess, sys
from PIL import Image

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
DIR = 'sprites/series/buck-rogers/combat'
ORIG = '7a6aa23'
SETS = {'hitek': 53, 'lotek': 52}
GREY = (82, 82, 82)


def original(name):
    return Image.open(io.BytesIO(subprocess.check_output(
        ['git', 'show', f'{ORIG}:{DIR}/{name}.png'], cwd=ROOT))).convert('RGB')


def old_floor(s, x, y):
    if s == 'lotek':
        return GREY
    return (0, 0, 0) if y % 8 == 0 or x % 8 == y % 8 else GREY


def wall_layer(s, im):
    """RGBA: the tile's pixels that are not its floor; hatch-black pixels count when 2+ neighbours do."""
    w = Image.new('RGBA', im.size)
    on = {(x, y) for y in range(24) for x in range(24) if im.getpixel((x, y)) != old_floor(s, x, y)}
    black = {(x, y) for y in range(24) for x in range(24) if (x, y) not in on and im.getpixel((x, y)) == (0, 0, 0)}
    on |= {p for p in black if sum((p[0] + dx, p[1] + dy) in on for dx, dy in ((1, 0), (-1, 0), (0, 1), (0, -1))) >= 2}
    for x, y in on:
        w.putpixel((x, y), im.getpixel((x, y)) + (255,))
    return w


def floor(seed):
    """24x24 steel panel: lit top/left edge, dark bottom/right groove (the grid), corner rivets, faint seams."""
    r = random.Random(seed)
    im = Image.new('RGB', (24, 24))
    for y in range(24):
        for x in range(24):
            n = r.randint(-2, 2)
            im.putpixel((x, y), (40 + n, 45 + n, 56 + n))
    for i in range(24):
        im.putpixel((i, 0), (62, 70, 86)); im.putpixel((0, i), (62, 70, 86))
        im.putpixel((i, 23), (20, 22, 28)); im.putpixel((23, i), (20, 22, 28))
    for i in range(3, 21):  # inset plate line
        im.putpixel((i, 3), (34, 38, 48)); im.putpixel((3, i), (34, 38, 48))
        im.putpixel((i, 20), (48, 54, 66)); im.putpixel((20, i), (48, 54, 66))
    for x, y in ((2, 2), (21, 2), (2, 21), (21, 21)):
        im.putpixel((x, y), (74, 82, 98))
    kind = seed % 4
    if kind == 1:  # vent slots
        for y in (9, 11, 13):
            for x in range(8, 16):
                im.putpixel((x, y), (28, 31, 40))
    elif kind == 2:  # faint cyan status light
        im.putpixel((17, 6), (60, 110, 120)); im.putpixel((18, 6), (44, 80, 92))
    return im


def main():
    args = sys.argv[1:]
    walls = args[args.index('--walls') + 1] if '--walls' in args else None
    mask = args[args.index('--mask') + 1] if '--mask' in args else None
    if mask:
        os.makedirs(mask, exist_ok=True)
    for s, n in SETS.items():
        for t in range(n):
            name = f'{s}_tile_{t:03d}'
            w = wall_layer(s, original(name))
            if mask:
                w.save(f'{mask}/{name}.png')
            if walls and os.path.exists(f'{walls}/{name}.png'):
                new = Image.open(f'{walls}/{name}.png').convert('RGBA')
                w = Image.composite(new, Image.new('RGBA', (24, 24)), w.getchannel('A'))  # keep the original footprint
            out = floor(t).convert('RGBA')
            out.alpha_composite(w)
            out.convert('RGB').save(f'{ROOT}/{DIR}/{name}.png')


if __name__ == '__main__':
    main()
