"""Pixel index of the originals behind the pool textures (walls/series/pool), for reuse before paying the image API.
  python3 pool_index.py             rebuild pool_index.txt from por-go and cotab-go
  python3 pool_index.py DUMPDIR     list DUMPDIR's walls (gobox viewdump -transparent output) with the pool texture
                                    of the same original pixels, or "new"
pool_index.txt: <sha1 of the wall picture> <texture> <game> <dump name>. The key is the picture with fully transparent
pixels zeroed and the sky colour (13, 0x55FFFF) cleared, so por's viewdump, cotab's walldump and gobox's viewdump
of the same wall give the same key."""
import hashlib, importlib.util, subprocess, sys, tempfile
from pathlib import Path
import numpy as np
from PIL import Image

HERE = Path(__file__).resolve().parent
GOBOX = HERE.parents[1]
PROJ = GOBOX.parent
POOL = GOBOX / "walls" / "series" / "pool"
FILE = HERE / "pool_index.txt"
SKY = [(85, 255, 255)]


def key(p):
    a = np.array(Image.open(p).convert("RGBA"))
    for c in SKY:
        a[(a[..., :3] == c).all(axis=2), 3] = 0
    a[a[..., 3] == 0] = 0
    return hashlib.sha1(a.tobytes()).hexdigest()


def tables(game, path, dump):
    """(texture, dump png) of every pool texture an area table names: generated walls and reused ones."""
    spec = importlib.util.spec_from_file_location(f"areas_{game}", path)
    m = importlib.util.module_from_spec(spec)
    sys.path.insert(0, str(path.parent))
    spec.loader.exec_module(m)
    if hasattr(m, "load"):  # gobox tools/remaster/areas.py: the game dir
        m.load(str(PROJ / f"{game}-go"))
    for n, a in m.AREAS.items():
        for v in set(a["walls"]) | set(a.get("reuse", {})):
            name = a.get("reuse", {}).get(v) or (lambda o: f"wall_{o[0]}_v{o[1]:02}")(m.origin(n, v))
            p = dump / f"{m.src(n, v)}.png"
            if (POOL / f"{name}.png").exists() and p.exists():
                yield name, p


def build():
    rows = {}
    for game, path, dump in [("por", PROJ / "por-go/phlan-poc/areas.py", PROJ / "por-go/phlan-poc/viewdump"),
                             ("cotab", HERE / "areas.py", PROJ / "cotab-go/remaster/viewdump")]:
        for name, p in tables(game, path, dump):
            rows.setdefault(key(p), (name, game, p.stem))
    FILE.write_text("".join(f"{k} {n} {g} {s}\n" for k, (n, g, s) in sorted(rows.items(), key=lambda r: r[1])))
    print(len(rows), "originals,", len({r[0] for r in rows.values()}), "of", len(list(POOL.glob("wall_*.png"))), "pool walls")


def lookup(dump):
    idx = {l.split()[0]: l.split()[1] for l in FILE.read_text().splitlines()}
    hit = 0
    for p in sorted(Path(dump).glob("*.png")):
        t = idx.get(key(p))
        hit += bool(t)
        print(p.stem, t or "new")
    print(f"# {hit} reused, {len(list(Path(dump).glob('*.png'))) - hit} new", file=sys.stderr)


if __name__ == "__main__":
    lookup(sys.argv[1]) if sys.argv[1:] else build()
