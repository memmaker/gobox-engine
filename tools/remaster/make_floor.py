"""The 3D floor: one OpenAI edit of the game's floor picture into a seamless top-down texture.
  python3 make_floor.py GAMEDIR
remaster.json "floor": {"ref": reference png, "prompt": text, "raw": raw output name, "out": texture path},
all relative to the game dir; "rw" as in make_textures.py."""
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


buf = io.BytesIO()
Image.open(ROOT / F["ref"]).convert("RGB").resize((1024, 1024), Image.NEAREST).save(buf, "PNG")
png = rw.openai(F["prompt"], buf.getvalue())
raw = areas.REM / "raw"
raw.mkdir(exist_ok=True)
(raw / F["raw"]).write_bytes(png)
out = ROOT / F["out"]
if out.exists() and "--force" not in sys.argv:  # out may be in the shared gobox/walls pool
    sys.exit(f"{out} exists; rerun with --force to overwrite")
tile(Image.open(io.BytesIO(png))).save(out)
print("floor: ok")
