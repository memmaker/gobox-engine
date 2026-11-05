"""The 3D floor: one OpenAI edit of the game's floor picture into a seamless top-down texture.
  python3 make_floor.py GAMEDIR [--force] [NAME...]
remaster.json "floor": {"ref": reference png, "prompt": text, "raw": raw output name, "out": texture path},
all relative to the game dir; "rw" as in make_textures.py. "floor" may be a list of such entries (NAMEs pick
entries by raw name); an entry may add "crop": [x0, y0, x1, y1] of ref, and "sky": [w, h] to keep the picture
untiled at that size (a panning sky_* backdrop) instead of a 256x256 tile. Each entry with "pixel": path also
writes the pixel-style version: 256 wide, 48 colours (8 kept for highlights), no dither."""
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


def pixel(im):
    """The pixel style: 256 wide, 48 colours without dither. Small lights survive: a shrink keeps a cell's
    brightest source pixel where it stands out (a lit window would average into the dark), and 8 of the 48
    colours are taken from the brightest pixels, so median cut cannot spend them all on large areas."""
    im = im.convert("RGB")
    w, h = im.size
    size = (256, max(1, h * 256 // w))
    out = np.asarray(im.resize(size, Image.LANCZOS), np.float32)
    k = w // 256
    if k > 1 and w == 256 * k and h == size[1] * k:
        a = np.asarray(im, np.float32).reshape(size[1], k, 256, k, 3).transpose(0, 2, 1, 3, 4).reshape(size[1], 256, k * k, 3)
        lum = a @ [0.299, 0.587, 0.114]
        top = a[np.arange(size[1])[:, None], np.arange(256), lum.argmax(2)]
        spot = (lum.max(2) - out @ [0.299, 0.587, 0.114]) > 60
        out[spot] = top[spot]
    small = Image.fromarray(out.astype(np.uint8), "RGB")
    px = out.reshape(-1, 3)
    lum = px @ [0.299, 0.587, 0.114] + (px.max(1) - px.min(1)) * 0.5
    hi = px[lum >= np.percentile(lum, 99.5)].astype(np.uint8)
    hipal = Image.fromarray(hi.reshape(1, -1, 3), "RGB").quantize(8, Image.Quantize.MEDIANCUT).getpalette()[:24]
    cols = small.quantize(40, Image.Quantize.MEDIANCUT).getpalette()[:120] + hipal
    pal = Image.new("P", (1, 1))
    pal.putpalette(cols + cols[:3] * (256 - len(cols) // 3))  # padded with a used colour, not black
    return small.quantize(palette=pal, dither=Image.Dither.NONE).convert("RGBA")


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
    if "pixel" in F:
        (ROOT / F["pixel"]).parent.mkdir(parents=True, exist_ok=True)
        pixel(im).save(ROOT / F["pixel"])
    print(F["raw"], "ok")
