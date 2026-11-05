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


def tile(im):
    """Crossfade the outer bands in both axes so the texture repeats without seams."""
    a = np.asarray(im.convert("RGB"), np.float32)
    for axis in (1, 0):
        n = a.shape[axis]
        b = n // 8
        t = np.linspace(0, 1, b).reshape((1, b, 1) if axis == 1 else (b, 1, 1))
        lo = np.take(a, range(b), axis=axis)
        hi = np.take(a, range(n - b, n), axis=axis)
        a = np.concatenate([np.take(a, range(b, n - b), axis=axis), hi * (1 - t) + lo * t], axis=axis)
    return Image.fromarray(a.astype(np.uint8), "RGB").resize((256, 256), Image.LANCZOS)


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
