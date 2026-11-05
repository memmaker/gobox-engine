"""The 3D floor: one OpenAI edit of the game's floor picture into a seamless top-down texture.
  python3 make_floor.py GAMEDIR [--force] [NAME...]
remaster.json "floor": {"ref": reference png, "prompt": text, "raw": raw output name, "out": texture path},
all relative to the game dir; "rw" as in make_textures.py. "floor" may be a list of such entries (NAMEs pick
entries by raw name); an entry may add "crop": [x0, y0, x1, y1] of ref, and "sky": [w, h] to keep the picture
untiled at that size (a panning sky_* backdrop) instead of a 256x256 tile. Pixel variants ("pixel": path)
come from PixelLab only (make_pixel.py); never derive them from this output by shrinking."""
import io, sys, types
from pathlib import Path
import numpy as np
from PIL import Image
import areas

CFG = areas.load(sys.argv[1])
F = CFG["floor"]
ROOT = areas.GAME
RW = ROOT / CFG.get("rw", "tools/remaster-walls.py")
rw = types.ModuleType("rw")
rw.__file__ = str(RW)
exec(compile(RW.read_text().replace('    nm.CONFIG["imageConfig"]["aspectRatio"] = "1:1"\n', ""), str(RW), "exec"), rw.__dict__)


def _cut(a, axis):
    """Crop along axis where the picture repeats: the span [s, s+p) (p >= 5/8 of the side) whose end lines
    best match its start lines, then fade the last n/32 lines into the lines just before s (their true
    continuation), so the wrap is seamless without the ghost bands of a wide crossfade of unrelated content."""
    a = np.moveaxis(a, axis, 0)
    n = a.shape[0]
    L = a[:, ::4].reshape(n, -1)
    D = np.empty((n, n), np.float32)
    for i in range(0, n, 128):
        D[i:i + 128] = np.abs(L[i:i + 128, None] - L[None]).mean(2)
    b = n // 32
    best = None
    for p in range(n * 5 // 8, n - b - 1):
        s = np.arange(b, n - p - 1)
        sc = sum(D[s + k, s + p + k] for k in (-1, 0, 1))
        i = sc.argmin()
        if best is None or sc[i] < best[0]:
            best = (sc[i], s[i], p)
    _, s, p = best
    out = a[s:s + p].copy()
    t = np.linspace(0, 1, b + 2)[1:-1].reshape(-1, *[1] * (a.ndim - 1))
    out[p - b:] = out[p - b:] * (1 - t) + a[s - b:s] * t
    return np.moveaxis(out, 0, axis)


def tile(im):
    """A seamless 256x256 tile: pattern-aligned crop in both axes (_cut)."""
    a = _cut(_cut(np.asarray(im.convert("RGB"), np.float32), 1), 0)
    return Image.fromarray(a.clip(0, 255).astype(np.uint8), "RGB").resize((256, 256), Image.LANCZOS)


names = [a for a in sys.argv[2:] if not a.startswith("--")]
for F in (F if isinstance(F, list) else [F]):
    if names and F["raw"] not in names:
        continue
    out = ROOT / F["out"]
    if out.exists() and "--force" not in sys.argv:  # out may be in the shared gobox/walls pool
        print(f"{out} exists; skipped (--force to overwrite)")
        continue
    ref = Image.open(ROOT / F["ref"]).convert("RGB")
    if "crop" in F:
        ref = ref.crop(F["crop"])
    buf = io.BytesIO()
    ref.resize((1024, 1024), Image.NEAREST).save(buf, "PNG")
    raw = areas.REM / "raw"
    raw.mkdir(exist_ok=True)
    if (raw / F["raw"]).exists() and "--force" not in sys.argv:  # a paid call already made: reuse it
        png = (raw / F["raw"]).read_bytes()
    else:
        png = rw.openai(F["prompt"], buf.getvalue())
        (raw / F["raw"]).write_bytes(png)
    im = Image.open(io.BytesIO(png))
    im = im.convert("RGB").resize(tuple(F["sky"]), Image.LANCZOS) if "sky" in F else tile(im)
    out.parent.mkdir(parents=True, exist_ok=True)
    im.save(out)
    print(F["raw"], "ok")
