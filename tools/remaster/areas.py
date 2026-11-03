"""Generic area tables for the remaster pipeline. The per-game data is read from the
game's remaster dir (GAME/remaster/):
  remaster.json   settings (see README.md)
  areas_data.py   the AREAS table: executed here with area(), src(), origin(), AREAS, DUMP,
                  and the description fragments are the game's own.
Keys per area: chunks (viewdump name of wall values 1-5, 6-10, ...) or srcs (value -> dump name),
walls (value -> description for the image model), notrim, water, tile, noseam, arch, mask, sky,
shape, clear, cut, props, reuse (value -> texture of another game, copied from cfg "reuse_dir").
"""
import hashlib, json
from pathlib import Path

AREAS = {}
CFG = {}
GAME = REM = DUMP = None


def src(name, v):
    """walldump name of wall value v of an area."""
    if "srcs" in AREAS[name]:
        return AREAS[name]["srcs"][v]
    return f"{AREAS[name]['chunks'][(v - 1) // 5]}-wall{(v - 1) % 5 + 1}"


def _key(s):
    from PIL import Image
    return hashlib.sha1(Image.open(DUMP / f"{s}.png").convert("RGBA").tobytes()).digest()


def origin(name, v):
    """(area, value) of the first area here with the same original art as wall v of area name."""
    k = _key(src(name, v))
    for n, a in AREAS.items():
        for w in a["walls"]:
            if _key(src(n, w)) == k:
                return n, w
    return name, v


def area(name, digit, walls, **kw):
    """Several-map area: walls is value -> (dump name, description or None for a reused wall)."""
    a = dict(notrim=set(), water=set(), tile=set(), noseam=set(), arch=set(), mask=set(), sky=set(), reuse={})
    a.update(kw)
    AREAS[name] = a
    a["srcs"] = {v: s for v, (s, _) in walls.items()}
    a["walls"] = {v: d for v, (s, d) in walls.items() if d}
    for v, (s, d) in walls.items():
        if d is None:
            a["reuse"][v] = kw.get("por", {})[v]


def load(game):
    """Read GAME/remaster/remaster.json and areas_data.py; returns the settings."""
    global GAME, REM, DUMP
    GAME = Path(game).resolve()
    REM = GAME / "remaster"
    DUMP = REM / "viewdump"
    CFG.update(json.loads((REM / "remaster.json").read_text()))
    ns = dict(AREAS=AREAS, DUMP=DUMP, area=area, src=src, origin=origin)
    exec(compile((REM / "areas_data.py").read_text(), str(REM / "areas_data.py"), "exec"), ns)
    return CFG
